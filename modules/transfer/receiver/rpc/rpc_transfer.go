package rpc

import (
	"fmt"
	cmodel "github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/common/redisdb"
	cutils "github.com/signmem/falcon-plus/common/utils"
	"github.com/signmem/falcon-plus/modules/transfer/g"
	"github.com/signmem/falcon-plus/modules/transfer/proc"
	"github.com/signmem/falcon-plus/modules/transfer/sender"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
    illegalCharRegexp *regexp.Regexp
)

type Transfer int

type TransferResp struct {
	Msg        string
	Total      int
	ErrInvalid int
	Latency    int64
}

func (t *TransferResp) String() string {
	s := fmt.Sprintf("TransferResp total=%d, err_invalid=%d, latency=%dms",
		t.Total, t.ErrInvalid, t.Latency)
	if t.Msg != "" {
		s = fmt.Sprintf("%s, msg=%s", s, t.Msg)
	}
	return s
}

func (this *Transfer) Ping(req cmodel.NullRpcRequest, resp *cmodel.SimpleRpcResponse) error {
	return nil
}

func (t *Transfer) Update(args []*cmodel.MetricValue, reply *cmodel.TransferResponse) error {
	return RecvMetricValues(args, reply, "rpc")
}


func InitIllegalCharRegexp() {
    IllegalChar := g.Config().IllegalChar
    IllegalCharString := strings.Join(IllegalChar, "|")
    illegalCharRegexp = regexp.MustCompile(IllegalCharString)
}



// process new metric values
func RecvMetricValues(args []*cmodel.MetricValue, reply *cmodel.TransferResponse, from string) error {
	start := time.Now()
	reply.Invalid = 0

	//IllegalMetric := g.Config().IllegalMetric

	items := make([]*cmodel.MetaData, 0, len(args))
	log_items := make([]*sender.LogMetricItem,0, len(args))

	for _, v := range args {

		if v == nil {
			reply.Invalid += 1
			continue
		}

		newV := *v

		// 历史遗留问题.
		// 老版本agent上报的metric=kernel.hostname的数据,其取值为string类型,现在已经不支持了;所以,这里硬编码过滤掉
		if newV.Metric == "kernel.hostname" {
			reply.Invalid += 1
			continue
		}

		if newV.Metric == "" || newV.Endpoint == "" {
			reply.Invalid += 1
			continue
		}

		newV.Metric = illegalCharRegexp.ReplaceAllString(newV.Metric, "")
		// v.Metric = re.ReplaceAllString(v.Metric, "")

		if newV.Type != g.COUNTER && newV.Type != g.GAUGE && newV.Type != g.DERIVE && newV.Type != g.LOG {
			reply.Invalid += 1
			continue
		}

		if newV.Value == nil {
			reply.Invalid += 1
			continue
		}

		if len(newV.Tags) > 0 {

			var validTags []string
			tags := strings.Split(newV.Tags, ",")

			skipMetric := false
			for _, tag := range tags {
				if cutils.IsChinese(tag) {
					reply.Invalid += 1
					skipMetric = true
					break
				}

				splitTag := strings.Split(tag, "=")
				if len(splitTag) != 2 || len(splitTag[1]) == 0 {
					continue
				}

				validTags = append(validTags, illegalCharRegexp.ReplaceAllString(tag,""))
			}

			if skipMetric {
				continue
			}

			if len(validTags) > 0 {
				newV.Tags = strings.Join(validTags,",")
			} else {
				newV.Tags = ""
			}

		}

		if newV.Step <= 0 {
			reply.Invalid += 1
			continue
		}

		if len(newV.Metric)+len(newV.Tags) > 510 {
			reply.Invalid += 1
			continue
		}

		// TODO 呵呵,这里需要再优雅一点
		now := start.Unix()
		maxtime := now + 172800
		mintime := now - 172800
		if newV.Timestamp <= mintime || newV.Timestamp > maxtime {
			newV.Timestamp = now
		}


		// 要处理 agent.alive 需要在这里增加处理方法  
		// url api || redis || etcd 都可以  

		if newV.Metric == "agent.alive" {
			service := newV.Metric
			hostname := newV.Endpoint

			go func(service, hostname string) {
				conn := redisdb.Pool.Get()
				defer conn.Close()

				// timeoutCtx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)
				// defer cancel()

				key := "/" + service + "/" + hostname
				_, err := conn.Do("SET", key, time.Now().Unix(), "EX", 3600)
				if err != nil {
					proc.SendToRedisFailCnt.Incr()
					log.Printf("redis write error: service=%s, hostname=%s, err=%v", service, hostname, err)
				} else {
					proc.SendToRedisCnt.Incr()
				}

				/*
				_, err := redisdb.RedisServiceWriteTimeout(timeoutCtx, service, hostname)
				if err != nil {
					proc.SendToRedisFailCnt.Incr()
					log.Printf("redis write error: service=%s, hostname=%s, err=%v", service, hostname, err)
				} else {
					proc.SendToRedisCnt.Incr()
				}
				*/
			}(service, hostname)
		}
		// edit by terry.zeng

		var err error
		if newV.Type == g.LOG {
			fs := &sender.LogMetricItem{
				Metric:    newV.Metric,
				Endpoint:  newV.Endpoint,
				Timestamp: newV.Timestamp,
				Step:      int(newV.Step),
				Tags:      cutils.DictedTagstring(newV.Tags),
			}
			switch cv := newV.Value.(type) {
			case string:
				fs.Value = cv
			case float64:
				fs.Value = strconv.FormatFloat(cv, 'f', -1, 64)
			case int64:
				fs.Value = strconv.FormatInt(cv, 64)
			default:
				continue
			}
			log_items = append(log_items, fs)
		} else {
			fv := &cmodel.MetaData{
				Metric:      newV.Metric,
				Endpoint:    newV.Endpoint,
				Timestamp:   newV.Timestamp,
				Step:        newV.Step,
				CounterType: newV.Type,
				Tags:        cutils.DictedTagstring(newV.Tags), //TODO tags键值对的个数,要做一下限制
			}
			valid := true
			var vv float64
			switch cv := newV.Value.(type) {
			case string:
				vv, err = strconv.ParseFloat(cv, 64)
				if err != nil {
					fs := &sender.LogMetricItem{
						Metric:    newV.Metric,
						Endpoint:  newV.Endpoint,
						Timestamp: newV.Timestamp,
						Value:     cv,
						Step:      int(newV.Step),
						Tags:      cutils.DictedTagstring(newV.Tags),
					}
					log_items = append(log_items, fs)
					continue
				}
			case float64:
				vv = cv
			case int64:
				vv = float64(cv)
			default:
				valid = false
			}

			if !valid {
				reply.Invalid += 1
				continue
			}
			fv.Value = vv
			items = append(items, fv)
		}
	}

	// statistics
	cnt := int64(len(items) + len(log_items))
	proc.RecvCnt.IncrBy(cnt)
	if from == "rpc" {
		proc.RpcRecvCnt.IncrBy(cnt)
	} else if from == "http" {
		proc.HttpRecvCnt.IncrBy(cnt)
	}

	cfg := g.Config()

	if cfg.Graph.Enabled {
		sender.Push2GraphSendQueue(items)
	}

	if cfg.Judge.Enabled {
		sender.Push2JudgeSendQueue(items)
	}

	if cfg.Tsdb.Enabled {
		sender.Push2TsdbSendQueue(items)
	}
	//added by vincent.zhang for sending to kafka
	if cfg.Kafka.Enabled {
		sender.Push2KafkaSendQueue(items)
	}

	if cfg.Kafka.LogEnabled && len(log_items) > 0 {
		sender.Push2KafkaLogSendQueue(log_items)
	}

	reply.Message = "ok"
	reply.Total = len(args)
	reply.Latency = (time.Now().UnixNano() - start.UnixNano()) / 1000000

	return nil
}
