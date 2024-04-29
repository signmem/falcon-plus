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

// process new metric values
func RecvMetricValues(args []*cmodel.MetricValue, reply *cmodel.TransferResponse, from string) error {
	start := time.Now()
	reply.Invalid = 0

	//IllegalMetric := g.Config().IllegalMetric
	IllegalChar := g.Config().IllegalChar
	IllegalCharString := strings.Join(IllegalChar,"|")
	var re = regexp.MustCompile(IllegalCharString)

	items := []*cmodel.MetaData{}
	log_items := []*sender.LogMetricItem{}
	for _, v := range args {
		if v == nil {
			reply.Invalid += 1
			continue
		}

		// 历史遗留问题.
		// 老版本agent上报的metric=kernel.hostname的数据,其取值为string类型,现在已经不支持了;所以,这里硬编码过滤掉
		if v.Metric == "kernel.hostname" {
			reply.Invalid += 1
			continue
		}

		if v.Metric == "" || v.Endpoint == "" {
			reply.Invalid += 1
			continue
		}

		v.Metric = re.ReplaceAllString(v.Metric, "")

		if v.Type != g.COUNTER && v.Type != g.GAUGE && v.Type != g.DERIVE && v.Type != g.LOG {
			reply.Invalid += 1
			continue
		}

		if v.Value == "" {
			reply.Invalid += 1
			continue
		}

		if len(v.Tags) > 0 {
			var validTags []string

			tags := strings.Split(v.Tags, ",")

			for _, tag := range tags {
				if cutils.IsChinese(tag) {
					reply.Invalid += 1
					continue
				}

				splitTag := strings.Split(tag, "=")
				if len(splitTag) != 2 {
					continue
				}

				if  len(splitTag[1]) == 0 {
					continue
				}

				validTags = append(validTags, re.ReplaceAllString(tag,""))
			}
			if len(validTags) > 0 {
				v.Tags = strings.Join(validTags,",")
			} else {
				v.Tags = ""
			}

		}

		if v.Step <= 0 {
			reply.Invalid += 1
			continue
		}

		if len(v.Metric)+len(v.Tags) > 510 {
			reply.Invalid += 1
			continue
		}

		// TODO 呵呵,这里需要再优雅一点
		now := start.Unix()
		if v.Timestamp <= 0 || v.Timestamp > now*2 {
			v.Timestamp = now
		}


		// 要处理 agent.alive 需要在这里增加处理方法  
		// url api || redis || etcd 都可以  

		if v.Metric == "agent.alive" {
			service := v.Metric
			hostname := v.Endpoint
			_, err := redisdb.RedisServiceWrite(service, hostname)
			if err != nil {
				proc.SendToRedisFailCnt.Incr()
				log.Printf("redisdb.RedisServiceWrite() error: %s", err)
			} else {
				proc.SendToRedisCnt.Incr()
			}
		}
		// edit by terry.zeng


		/*
			fv := &cmodel.MetaData{
				Metric:      v.Metric,
				Endpoint:    v.Endpoint,
				Timestamp:   v.Timestamp,
				Step:        v.Step,
				CounterType: v.Type,
				Tags:        cutils.DictedTagstring(v.Tags), //TODO tags键值对的个数,要做一下限制
			}

			valid := true
			var vv float64
			var err error
			switch cv := v.Value.(type) {
			case string:
				vv, err = strconv.ParseFloat(cv, 64)
				if err != nil {
					valid = false
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
		*/

		var err error
		if v.Type == g.LOG {
			fs := &sender.LogMetricItem{
				Metric:    v.Metric,
				Endpoint:  v.Endpoint,
				Timestamp: v.Timestamp,
				Step:      int(v.Step),
				Tags:      cutils.DictedTagstring(v.Tags),
			}
			switch cv := v.Value.(type) {
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
				Metric:      v.Metric,
				Endpoint:    v.Endpoint,
				Timestamp:   v.Timestamp,
				Step:        v.Step,
				CounterType: v.Type,
				Tags:        cutils.DictedTagstring(v.Tags), //TODO tags键值对的个数,要做一下限制
			}
			valid := true
			var vv float64
			switch cv := v.Value.(type) {
			case string:
				vv, err = strconv.ParseFloat(cv, 64)
				if err != nil {
					fs := &sender.LogMetricItem{
						Metric:    v.Metric,
						Endpoint:  v.Endpoint,
						Timestamp: v.Timestamp,
						Value:     cv,
						Step:      int(v.Step),
						Tags:      cutils.DictedTagstring(v.Tags),
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
