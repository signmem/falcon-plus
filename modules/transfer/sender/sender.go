package sender

import (
	"fmt"
	backend "github.com/signmem/falcon-plus/common/backend_pool"
	cmodel "github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/transfer/g"
	"github.com/signmem/falcon-plus/modules/transfer/proc"
	rings "github.com/signmem/consistent/rings"
	nlist "github.com/toolkits/container/list"
	"log"
	//added by vincent.zhang for sending to kafka
	//"github.com/IBM/sarama"
	cutils "github.com/signmem/falcon-plus/common/utils"
)

const (
	DefaultSendQueueMaxSize = 102400 //10.24w
)

// 默认参数
var (
	MinStep int //最小上报周期,单位sec
)

// 服务节点的一致性哈希环
// pk -> node
var (
	JudgeNodeRing *rings.ConsistentHashNodeRing
	GraphNodeRing *rings.ConsistentHashNodeRing
)

// 发送缓存队列
// node -> queue_of_data
var (
	TsdbQueue   *nlist.SafeListLimited
	JudgeQueues = make(map[string]*nlist.SafeListLimited)
	GraphQueues = make(map[string]*nlist.SafeListLimited)
	//added by vincent.zhang for sending to kafka
	KafkaQueue    *nlist.SafeListLimited
	KafkaLogQueue *nlist.SafeListLimited
	// added by qimin.xu
	KafkaFilterQueues = make(map[string]*nlist.SafeListLimited)
)

// 连接池
// node_address -> connection_pool
var (
	JudgeConnPools     *backend.SafeRpcConnPools
	TsdbConnPoolHelper *backend.TsdbConnPoolHelper
	GraphConnPools     *backend.SafeRpcConnPools
	//added by vincent.zhang for sending to kafka
	//KafkaPool          *KafkaProducerPool
	kafkaProducer    *KafkaProducer
	kafkaLogProducer *KafkaProducer
	// added by qimin.xu
	kafkaFilterProducer = make(map[string]*KafkaProducer)
	//end
)

// 初始化数据发送服务, 在main函数中调用
func Start() {
	// 初始化默认参数
	MinStep = g.Config().MinStep
	if MinStep < 1 {
		MinStep = 30 //默认30s
	}
	//
	initConnPools()
	initSendQueues()
	initNodeRings()
	// SendTasks依赖基础组件的初始化,要最后启动
	startSendTasks()
	startSenderCron()
	log.Println("send.Start, ok")
}

// 将数据 打入 某个Judge的发送缓存队列, 具体是哪一个Judge 由一致性哈希 决定
func Push2JudgeSendQueue(items []*cmodel.MetaData) {
	for _, item := range items {
		pk := item.PK()
		node, err := JudgeNodeRing.GetNode(pk)
		if err != nil {
			log.Println("E:", err)
			continue
		}

		// align ts
		step := int(item.Step)
		if step < MinStep {
			step = MinStep
		}
		ts := alignTs(item.Timestamp, int64(step))

		judgeItem := &cmodel.JudgeItem{
			Endpoint:  item.Endpoint,
			Metric:    item.Metric,
			Value:     item.Value,
			Timestamp: ts,
			JudgeType: item.CounterType,
			Tags:      item.Tags,
		}
		Q := JudgeQueues[node]
		isSuccess := Q.PushFront(judgeItem)

		// statistics
		if !isSuccess {
			proc.SendToJudgeDropCnt.Incr()
		}
	}
}

// 将数据 打入 某个Graph的发送缓存队列, 具体是哪一个Graph 由一致性哈希 决定
func Push2GraphSendQueue(items []*cmodel.MetaData) {
	cfg := g.Config().Graph

	for _, item := range items {
		graphItem, err := convert2GraphItem(item)
		if err != nil {
			log.Println("E:", err)
			continue
		}
		pk := item.PK()

		// statistics. 为了效率,放到了这里,因此只有graph是enbale时才能trace
		proc.RecvDataTrace.Trace(pk, item)
		proc.RecvDataFilter.Filter(pk, item.Value, item)

		node, err := GraphNodeRing.GetNode(pk)
		if err != nil {
			log.Println("E:", err)
			continue
		}

		cnode := cfg.ClusterList[node]
		errCnt := 0
		for _, addr := range cnode.Addrs {
			Q := GraphQueues[node+addr]
			if !Q.PushFront(graphItem) {
				errCnt += 1
			}
		}

		// statistics
		if errCnt > 0 {
			proc.SendToGraphDropCnt.Incr()
		}
	}
}

// 打到Graph的数据,要根据rrdtool的特定 来限制 step、counterType、timestamp
func convert2GraphItem(d *cmodel.MetaData) (*cmodel.GraphItem, error) {
	item := &cmodel.GraphItem{}

	item.Endpoint = d.Endpoint
	item.Metric = d.Metric
	item.Tags = d.Tags
	item.Timestamp = d.Timestamp
	item.Value = d.Value
	item.Step = int(d.Step)
	if item.Step < MinStep {
		item.Step = MinStep
	}
	item.Heartbeat = item.Step * 2

	if d.CounterType == g.GAUGE {
		item.DsType = d.CounterType
		item.Min = "U"
		item.Max = "U"
	} else if d.CounterType == g.COUNTER {
		item.DsType = g.DERIVE
		item.Min = "0"
		item.Max = "U"
	} else if d.CounterType == g.DERIVE {
		item.DsType = g.DERIVE
		item.Min = "0"
		item.Max = "U"
	} else {
		return item, fmt.Errorf("not_supported_counter_type")
	}

	item.Timestamp = alignTs(item.Timestamp, int64(item.Step)) //item.Timestamp - item.Timestamp%int64(item.Step)

	return item, nil
}

// 将原始数据入到tsdb发送缓存队列
func Push2TsdbSendQueue(items []*cmodel.MetaData) {
	for _, item := range items {
		tsdbItem := convert2TsdbItem(item)
		isSuccess := TsdbQueue.PushFront(tsdbItem)

		if !isSuccess {
			proc.SendToTsdbDropCnt.Incr()
		}
	}
}

// 转化为tsdb格式
func convert2TsdbItem(d *cmodel.MetaData) *cmodel.TsdbItem {
	t := cmodel.TsdbItem{Tags: make(map[string]string)}

	for k, v := range d.Tags {
		t.Tags[k] = v
	}
	t.Tags["endpoint"] = d.Endpoint
	t.Metric = d.Metric
	t.Timestamp = d.Timestamp
	t.Value = d.Value
	return &t
}

// added by vincent.zhang for sending to Kafka
// 转化为kafka格式
type KafkaItem struct {
	Endpoint  string  `json:"endpoint"`
	Metric    string  `json:"metric"`
	Tags      string  `json:"tags"`
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	DsType    string  `json:"dstype"`
	Step      int     `json:"step"`
}

func (this *KafkaItem) KafkaString() (s string) {
	s = fmt.Sprintf("%s\t%s\t%s\t%d\t%.3f\t%s\t%d", this.Endpoint, this.Metric, this.Tags, this.Timestamp, this.Value, this.DsType, this.Step)
	return s
}

// 将原始数据入到kafka发送缓存队列
func Push2KafkaSendQueue(items []*cmodel.MetaData) {
	for _, item := range items {
		kafkaItem := convert2KafkaItem(item)
		isSuccess := KafkaQueue.PushFront(kafkaItem)

		if !isSuccess {
			proc.SendToKafkaDropCnt.Incr()
		}
		// added by qimin.xu for push item to cache Q
		// 根据配置过滤指标，并打到相应的topic
		mf := matchKafkaFilter(item)
		if len(mf) > 0 {
			for _, v := range mf {
				Q := KafkaFilterQueues[v]
				if !Q.PushFront(kafkaItem) {
					proc.SendToKafkaDropCntMap[v].Incr()
					log.Println("push item to filter Q error")
				}
			}
		}

	}
}

// added by qimin.xu for match filter
func matchKafkaFilter(item *cmodel.MetaData) []string {
	ret := []string{}
	for filter_k, filter_v := range g.Config().Kafka.Filter {
		item_tagv, ok := item.Tags[filter_v["tagk"]]
		if ok {
			if item_tagv == filter_v["tagv"] {
				ret = append(ret, filter_k)
			}
		}
	}
	return ret
}

func convert2KafkaItem(d *cmodel.MetaData) *KafkaItem {
	kafkaItem := &KafkaItem{
		Endpoint:  d.Endpoint,
		Metric:    d.Metric,
		Tags:      cutils.SortedTags(d.Tags),
		Timestamp: d.Timestamp,
		Value:     d.Value,
		Step:      int(d.Step),
		DsType:    d.CounterType,
	}
	return kafkaItem
}

type LogMetricItem struct {
	Endpoint  string            `json:"endpoint"`
	Metric    string            `json:"metric"`
	Tags      map[string]string `json:"tags"`
	Timestamp int64             `json:"timestamp"`
	Value     string            `json:"value"`
	Step      int               `json:"step"`
}

func (this *LogMetricItem) KafkaString() (s string) {
	s = fmt.Sprintf("%s\t%s\t%s\t%d\t%s\t%d", this.Endpoint, this.Metric, cutils.SortedTags(this.Tags), this.Timestamp, this.Value, this.Step)
	return s
}

// 将原始数据入到kafka Log发送缓存队列
func Push2KafkaLogSendQueue(items []*LogMetricItem) {
	for _, item := range items {
		isSuccess := KafkaLogQueue.PushFront(item)

		if !isSuccess {
			proc.SendToKafkaLogDropCnt.Incr()
		}
	}
}

//kafka end

func alignTs(ts int64, period int64) int64 {
	return ts - ts%period
}
