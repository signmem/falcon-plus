package proc

import (
	nproc "github.com/toolkits/proc"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/g"
)

// 统计指标的整体数据
var (
	// 计数统计,正确计数,错误计数, ...
	ConsumeCnt     = nproc.NewSCounterQps("ConsumeCnt")
	ConsumeDropCnt = nproc.NewSCounterQps("ConsumeDropCnt")

	SendToTrendCnt    = nproc.NewSCounterQps("SendToTrendCnt")
	SendToTransferCnt = nproc.NewSCounterQps("SendToTransferCnt")

	SendToTrendDropCnt    = nproc.NewSCounterQps("SendToTrendDropCnt")
	SendToTransferDropCnt = nproc.NewSCounterQps("SendToTransferDropCnt")

	SendToTrendFailCnt    = nproc.NewSCounterQps("SendToTrendFailCnt")
	SendToTransferFailCnt = nproc.NewSCounterQps("SendToTransferFailCnt")

	// 发送缓存大小
	TrendQueuesCnt    = nproc.NewSCounterBase("TrendSendCacheCnt")
	TransferQueuesCnt = nproc.NewSCounterBase("TransferSendCacheCnt")
)

func Start() {
	g.Logger.Println("proc.Start, ok")
}

func GetAll() []interface{} {
	ret := make([]interface{}, 0)

	// recv cnt
	ret = append(ret, ConsumeCnt.Get())
	ret = append(ret, ConsumeDropCnt.Get())

	// send cnt
	ret = append(ret, SendToTrendCnt.Get())
	ret = append(ret, SendToTransferCnt.Get())

	// drop cnt
	ret = append(ret, SendToTrendDropCnt.Get())
	ret = append(ret, SendToTransferDropCnt.Get())

	// send fail cnt
	ret = append(ret, SendToTrendFailCnt.Get())
	ret = append(ret, SendToTransferFailCnt.Get())

	// cache cnt
	ret = append(ret, TrendQueuesCnt.Get())
	ret = append(ret, TransferQueuesCnt.Get())

	return ret
}
