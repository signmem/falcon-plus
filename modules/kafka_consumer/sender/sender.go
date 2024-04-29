package sender

import (
	"strconv"
	"strings"

	backend "github.com/signmem/falcon-plus/common/backend_pool"
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/g"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/proc"
	"github.com/signmem/consistent/rings"
	nlist "github.com/toolkits/container/list"
)

const (
	DefaultSendQueueMaxSize = 102400 //10.24wß
)

// 服务节点的一致性哈希环
// pk -> node
var (
	TrendNodeRing *rings.ConsistentHashNodeRing
)

// 发送缓存队列
// node -> queue_of_data
var (
	TrendQueues   = make(map[string]*nlist.SafeListLimited)
	TransferQueue *nlist.SafeListLimited
)

// 连接池
// node_address -> connection_pool
var (
	TrendConnPools    *backend.SafeRpcConnPools
	TransferConnPools *backend.SafeRpcConnPools
	TransferMap       = make(map[string]string, 0)
	TransferHostnames = make([]string, 0)
)

// 初始化数据发送服务, 在main函数中调用
func Start() {
	initConnPools()
	initSendQueues()
	initNodeRings()
	// SendTasks依赖基础组件的初始化,要最后启动
	startSendTasks()
	startSenderCron()
	g.Logger.Println("send.Start, ok")
}

func convert2TrendItem(kafkaStr string) *model.TrendItem {
	arr := strings.Split(kafkaStr, "\t")
	//Endpoint, Metric, Tags, Timestamp, Value, DsType, Step
	if len(arr) != 7 {
		g.Logger.Errorf("kafka string length is error, string: %d", kafkaStr)
		return nil
	}
	var ts int64
	var val float64
	var step int
	var err error
	ts, err = strconv.ParseInt(arr[3], 10, 64)
	if err != nil {
		g.Logger.Errorf("kafka string timestamp is invalid, string: %d", kafkaStr)
		return nil
	}
	val, err = strconv.ParseFloat(arr[4], 64)
	if err != nil {
		g.Logger.Errorf("kafka string value is invalid, string: %d", kafkaStr)
		return nil
	}
	step, err = strconv.Atoi(arr[6])
	if err != nil {
		g.Logger.Errorf("kafka string step is invalid, string: %d", kafkaStr)
		return nil
	}
	return &model.TrendItem{
		Endpoint:  arr[0],
		Metric:    arr[1],
		Tags:      arr[2],
		Timestamp: ts,
		Value:     val,
		DsType:    arr[5],
		Step:      step,
	}
}

func trendConvert2MetricItem(trend *model.TrendItem) *model.MetricValue {
	if trend.Metric == "cpu.idle" {
		return &model.MetricValue{
			Endpoint:  trend.Endpoint,
			Metric:    "cpu.used",
			Value:     100.0 - trend.Value,
			Step:      int64(trend.Step),
			Type:      trend.DsType,
			Tags:      trend.Tags,
			Timestamp: trend.Timestamp,
		}
	}
	return nil
}

func push(item *model.TrendItem) {
	if item == nil {
		return
	}
	pk := item.PK()
	node, err := TrendNodeRing.GetNode(pk)
	if err != nil {
		g.Logger.Errorf("Get node error: %s", err)
		return
	}

	Q := TrendQueues[node]
	isSuccess := Q.PushFront(item)

	// statistics
	if !isSuccess {
		proc.SendToTrendDropCnt.Incr()
	}
}

// 将原始数据入到transfer发送缓存队列
func Push2TransferSendQueue(item *model.TrendItem) {
	if item == nil {
		return
	}
	metricItem := trendConvert2MetricItem(item)
	if metricItem != nil {
		isSuccess := TransferQueue.PushFront(metricItem)
		if !isSuccess {
			proc.SendToTransferDropCnt.Incr()
		}
	}
}

// 将数据 打入 某个AggregaotrPlus的发送缓存队列, 具体是哪一个由一致性哈希决定
func Push2TrendSendQueue(val string) {
	item := convert2TrendItem(val)
	if item == nil {
		proc.ConsumeDropCnt.Incr()
		return
	}
	//add ignore items that metric include net.port.listen, alive suffix, hardware prefix and step>1800s.
	percentCheckMap := g.Config().PercentCheck
	ignoreHostMap := g.Config().IgnoreHost

	if b, ok := ignoreHostMap[item.Endpoint]; ok && b {
		g.Logger.Debugf("hostname is invalid, ignore, host: %s, %s/%s, %.3f", item.Endpoint, item.Metric, item.Tags, item.Value)
		proc.ConsumeDropCnt.Incr()
	} else if item.Metric == "net.port.listen" {
		g.Logger.Debug("net.port.listen ignored")
		proc.ConsumeDropCnt.Incr()
	} else if strings.HasSuffix(item.Metric, "alive") {
		g.Logger.Debugf("%s:%s/%s metric has alive suffix metric, ignored", item.Endpoint, item.Metric, item.Tags)
		proc.ConsumeDropCnt.Incr()
	} else if strings.HasPrefix(item.Metric, "hardware") {
		g.Logger.Debugf("%s:%s/%s step has hardware prefix metric, ignored", item.Endpoint, item.Metric, item.Tags)
		proc.ConsumeDropCnt.Incr()
	} else if strings.HasPrefix(item.Metric, "docker") {
		g.Logger.Debugf("%s:%s/%s step has docker prefix metric, ignored", item.Endpoint, item.Metric, item.Tags)
		proc.ConsumeDropCnt.Incr()
	} else if item.Step > 1800 {
		g.Logger.Debugf("%s:%s/%s step is larger than 1800s, ignored, step:%d", item.Endpoint, item.Metric, item.Tags, item.Step)
		proc.ConsumeDropCnt.Incr()
	} else if b, ok := percentCheckMap[item.Metric]; ok && b {
		if item.Value > 100 || item.Value < 0 {
			g.Logger.Infof("%s:%s/%s percent value is invalid, ignore, value: %.3f", item.Endpoint, item.Metric, item.Tags, item.Value)
			proc.ConsumeDropCnt.Incr()
		} else {
			//log.Debug("start push %s:%s/%s percent value value: %.3f", item.Endpoint, item.Metric, item.Tags, item.Value)
			push(item)
			if g.Config().Transfer.Enabled {
				if item.Metric == "cpu.idle" {
					Push2TransferSendQueue(item)
				}
			}
		}
	} else {
		push(item)
	}
}
