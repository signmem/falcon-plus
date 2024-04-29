package proc

import (
	"github.com/signmem/falcon-plus/modules/transfer/g"
	nproc "github.com/toolkits/proc"
	"log"
)

// trace
var (
	RecvDataTrace = nproc.NewDataTrace("RecvDataTrace", 3)
)

// filter
var (
	RecvDataFilter = nproc.NewDataFilter("RecvDataFilter", 5)
)

// 统计指标的整体数据
var (
	// 计数统计,正确计数,错误计数, ...
	RecvCnt       = nproc.NewSCounterQps("RecvCnt")
	RpcRecvCnt    = nproc.NewSCounterQps("RpcRecvCnt")
	HttpRecvCnt   = nproc.NewSCounterQps("HttpRecvCnt")
	SocketRecvCnt = nproc.NewSCounterQps("SocketRecvCnt")

	SendToJudgeCnt = nproc.NewSCounterQps("SendToJudgeCnt")
	SendToTsdbCnt  = nproc.NewSCounterQps("SendToTsdbCnt")
	SendToGraphCnt = nproc.NewSCounterQps("SendToGraphCnt")
	//added by vincent.zhang for sending to kafka
	SendToKafkaCnt      = nproc.NewSCounterQps("SendToKafkaCnt")      // 转发到Kafka消息数(vipfalcon)
	SendToKafkaCntTotal = nproc.NewSCounterQps("SendToKafkaCntTotal") // 转发到Kafka消息总数
	SendToKafkaCntMap   = make(map[string]*nproc.SCounterQps)
	SendToKafkaLogCnt   = nproc.NewSCounterQps("SendToKafkaLogCnt") // 转发到Kafka Log消息数(vipfalcon-log)

	SendToJudgeDropCnt = nproc.NewSCounterQps("SendToJudgeDropCnt")
	SendToTsdbDropCnt  = nproc.NewSCounterQps("SendToTsdbDropCnt")
	SendToGraphDropCnt = nproc.NewSCounterQps("SendToGraphDropCnt")

	// added by terry.zeng for send to redis
	SendToRedisFailCnt = nproc.NewSCounterQps("SendToRedisFailCnt")
	SendToRedisCnt = nproc.NewSCounterQps("SendToRedisCnt")

	//added by vincent.zhang for sending to kafka
	SendToKafkaDropCnt    = nproc.NewSCounterQps("SendToKafkaDropCnt")    // Kafka缓存队列丢弃数(vipfalcon)
	SendToKafkaDropCntMap = make(map[string]*nproc.SCounterQps)           // 对应topic的缓存队列丢弃数
	SendToKafkaLogDropCnt = nproc.NewSCounterQps("SendToKafkaLogDropCnt") // Kafka Log队列丢弃数(vipfalcon-log)

	SendToJudgeFailCnt = nproc.NewSCounterQps("SendToJudgeFailCnt")
	SendToTsdbFailCnt  = nproc.NewSCounterQps("SendToTsdbFailCnt")
	SendToGraphFailCnt = nproc.NewSCounterQps("SendToGraphFailCnt")
	//added by vincent.zhang for sending to kafka
	SendToKafkaFailCnt      = nproc.NewSCounterQps("SendToKafkaFailCnt")      // 转发到Kafka失败数(vipfalcon)
	SendToKafkaFailCntTotal = nproc.NewSCounterQps("SendToKafkaFailCntTotal") // 转发到Kafka失败总数
	SendToKafkaFailCntMap   = make(map[string]*nproc.SCounterQps)
	SendToKafkaLogFailCnt   = nproc.NewSCounterQps("SendToKafkaLogFailCnt") // 转发到Kafka Log失败数(vipfalcon-log)

	// 发送缓存大小
	JudgeQueuesCnt = nproc.NewSCounterBase("JudgeSendCacheCnt")
	TsdbQueuesCnt  = nproc.NewSCounterBase("TsdbSendCacheCnt")
	GraphQueuesCnt = nproc.NewSCounterBase("GraphSendCacheCnt")
	//added by vincent.zhang for sending to kafka
	KafkaQueuesCnt    = nproc.NewSCounterBase("KafkaSendCacheCnt")    // Kafka缓存队列大小(vipfalcon)
	KafkaQueuesCntMap = make(map[string]*nproc.SCounterBase)          // 对应topic的缓存队列大小
	KafkaLogQueuesCnt = nproc.NewSCounterBase("KafkaLogSendCacheCnt") // Kafka Log缓存队列大小(vipfalcon-log)

	// http请求次数
	HistoryRequestCnt = nproc.NewSCounterQps("HistoryRequestCnt")
	InfoRequestCnt    = nproc.NewSCounterQps("InfoRequestCnt")
	LastRequestCnt    = nproc.NewSCounterQps("LastRequestCnt")
	LastRawRequestCnt = nproc.NewSCounterQps("LastRawRequestCnt")

	// http回执的监控数据条数
	HistoryResponseCounterCnt = nproc.NewSCounterQps("HistoryResponseCounterCnt")
	HistoryResponseItemCnt    = nproc.NewSCounterQps("HistoryResponseItemCnt")
	LastRequestItemCnt        = nproc.NewSCounterQps("LastRequestItemCnt")
	LastRawRequestItemCnt     = nproc.NewSCounterQps("LastRawRequestItemCnt")
)

func Start() {
	log.Println("proc.Start, ok")
}

func InitKafkaCntSet() {
	cfg := g.Config()
	if len(cfg.Kafka.Filter) > 0 {
		for f := range cfg.Kafka.Filter {
			SendToKafkaCntMap[f] = nproc.NewSCounterQps("SendToKafkaCnt_" + cfg.Kafka.Filter[f]["topic"])
			SendToKafkaFailCntMap[f] = nproc.NewSCounterQps("SendToKafkaFailCnt_" + cfg.Kafka.Filter[f]["topic"])
			SendToKafkaDropCntMap[f] = nproc.NewSCounterQps("SendToKafkaDropCnt_" + cfg.Kafka.Filter[f]["topic"])
			KafkaQueuesCntMap[f] = nproc.NewSCounterBase("KafkaSendCacheCnt_" + cfg.Kafka.Filter[f]["topic"])
		}
	}
}

func GetAll() []interface{} {
	cfg := g.Config()
	ret := make([]interface{}, 0)

	// recv cnt
	ret = append(ret, RecvCnt.Get())
	ret = append(ret, RpcRecvCnt.Get())
	ret = append(ret, HttpRecvCnt.Get())
	ret = append(ret, SocketRecvCnt.Get())

	// send cnt
	ret = append(ret, SendToJudgeCnt.Get())
	ret = append(ret, SendToTsdbCnt.Get())
	ret = append(ret, SendToGraphCnt.Get())

	ret = append(ret, SendToRedisCnt.Get())
	ret = append(ret, SendToRedisFailCnt.Get())

	//added by vincent.zhang for sending to kafka
	ret = append(ret, SendToKafkaCnt.Get())
	ret = append(ret, SendToKafkaLogCnt.Get())
	// added by qimin.xu
	ret = append(ret, SendToKafkaCntTotal.Get())
	if len(cfg.Kafka.Filter) > 0 {
		for f := range cfg.Kafka.Filter {
			ret = append(ret, SendToKafkaCntMap[f].Get())
			ret = append(ret, SendToKafkaFailCntMap[f].Get())
			ret = append(ret, SendToKafkaDropCntMap[f].Get())
			ret = append(ret, KafkaQueuesCntMap[f].Get())
		}
	}

	// drop cnt
	ret = append(ret, SendToJudgeDropCnt.Get())
	ret = append(ret, SendToTsdbDropCnt.Get())
	ret = append(ret, SendToGraphDropCnt.Get())
	//added by vincent.zhang for sending to kafka
	ret = append(ret, SendToKafkaDropCnt.Get())
	ret = append(ret, SendToKafkaLogDropCnt.Get())

	// send fail cnt
	ret = append(ret, SendToJudgeFailCnt.Get())
	ret = append(ret, SendToTsdbFailCnt.Get())
	ret = append(ret, SendToGraphFailCnt.Get())
	//added by vincent.zhang for sending to kafka
	ret = append(ret, SendToKafkaFailCnt.Get())
	ret = append(ret, SendToKafkaFailCntTotal.Get())
	ret = append(ret, SendToKafkaLogFailCnt.Get())

	// cache cnt
	ret = append(ret, JudgeQueuesCnt.Get())
	ret = append(ret, TsdbQueuesCnt.Get())
	ret = append(ret, GraphQueuesCnt.Get())
	//added by vincent.zhang for sending to kafka
	ret = append(ret, KafkaQueuesCnt.Get())
	ret = append(ret, KafkaLogQueuesCnt.Get())

	// http request
	ret = append(ret, HistoryRequestCnt.Get())
	ret = append(ret, InfoRequestCnt.Get())
	ret = append(ret, LastRequestCnt.Get())
	ret = append(ret, LastRawRequestCnt.Get())

	// http response
	ret = append(ret, HistoryResponseCounterCnt.Get())
	ret = append(ret, HistoryResponseItemCnt.Get())
	ret = append(ret, LastRequestItemCnt.Get())
	ret = append(ret, LastRawRequestItemCnt.Get())

	return ret
}
