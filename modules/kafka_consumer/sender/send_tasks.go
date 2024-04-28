package sender

import (
	"math/rand"
	"time"

	pfc "github.com/niean/goperfcounter"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/kafka_consumer/g"
	"github.com/open-falcon/falcon-plus/modules/kafka_consumer/proc"
	nsema "github.com/toolkits/concurrent/semaphore"
	"github.com/toolkits/container/list"
)

// send
const (
	DefaultSendTaskSleepInterval = time.Millisecond * 50 //默认睡眠间隔为50ms
)

// TODO 添加对发送任务的控制,比如stop等
func startSendTasks() {
	cfg := g.Config()
	// init semaphore
	trendConcurrent := cfg.Trend.MaxConns
	transferConcurrent := cfg.Transfer.MaxConns * len(cfg.Transfer.Cluster)

	if trendConcurrent < 1 {
		trendConcurrent = 1
	}

	if transferConcurrent < 1 {
		transferConcurrent = 1
	}

	// init send go-routines
	for node := range cfg.Trend.Cluster {
		// terrytsang
		if g.Config().DebugTraffer {
			g.Logger.Debugf("startSendTasks() node is:%s", node)
		}
		queue := TrendQueues[node]
		go forward2TrendTask(queue, node, trendConcurrent)
	}

	if cfg.Transfer.Enabled {
		go forward2TransferTask(TransferQueue, transferConcurrent)
	}
}

// Trend定时任务, 将Trend发送缓存中的数据 通过rpc连接池 发送到Trend
func forward2TrendTask(Q *list.SafeListLimited, node string, concurrent int) {
	batch := g.Config().Trend.Batch // 一次发送,最多batch条数据
	addr := g.Config().Trend.Cluster[node]
	// terrytsang
	if g.Config().DebugTraffer {
		g.Logger.Debugf("forward2TrendTask() addr is %v", addr)
	}
	sema := nsema.NewSemaphore(concurrent)

	for {
		items := Q.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}

		plusItems := make([]*model.TrendItem, count)
		for i := 0; i < count; i++ {
			plusItems[i] = items[i].(*model.TrendItem)
		}

		//	同步Call + 有限并发 进行发送s
		sema.Acquire()
		go func(addr string, plusItems []*model.TrendItem, count int) {
			defer sema.Release()

			resp := &model.SimpleRpcResponse{}
			var err error
			sendOk := false
			for i := 0; i < 3; i++ { //最多重试3次
				err = TrendConnPools.Call(addr, "Trend.Send", plusItems, resp)
				if err == nil {
					sendOk = true
					break
				}
				time.Sleep(time.Millisecond * 10)
				break
			}

			// statistics
			if !sendOk {
				g.Logger.Printf("send Trend %s:%s fail: %v", node, addr, err)
				proc.SendToTrendFailCnt.IncrBy(int64(count))
			} else {
				proc.SendToTrendCnt.IncrBy(int64(count))
			}
		}(addr, plusItems, count)
	}
}

// Transfer定时任务, 将MetricItem发送缓存中的数据 通过rpc连接池 发送到Transfer
func forward2TransferTask(Q *list.SafeListLimited, concurrent int) {
	cfg := g.Config()
	batch := cfg.Transfer.Batch
	maxConns := int64(cfg.Transfer.MaxConns)
	retry := cfg.Transfer.Retry
	if retry < 1 {
		retry = 1
	}

	sema := nsema.NewSemaphore(concurrent)
	transNum := len(TransferHostnames)

	for {
		items := Q.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(time.Millisecond * 50)
			continue
		}

		transItems := make([]*model.MetricValue, count)
		for i := 0; i < count; i++ {
			transItems[i] = items[i].(*model.MetricValue)
		}

		sema.Acquire()
		go func(transItems []*model.MetricValue, count int) {
			defer sema.Release()
			var err error

			// 随机遍历transfer列表，直到数据发送成功 或者 遍历完;随机遍历，可以缓解慢transfer
			resp := &model.TransferResponse{}
			sendOk := false

			for j := 0; j < retry && !sendOk; j++ {
				rint := rand.Int()
				for i := 0; i < transNum && !sendOk; i++ {
					idx := (i + rint) % transNum
					host := TransferHostnames[idx]
					addr := TransferMap[host]

					// 过滤掉建连缓慢的host, 否则会严重影响发送速率
					cc := pfc.GetCounterCount(host)
					if cc >= maxConns {
						continue
					}

					pfc.Counter(host, 1)
					err = TransferConnPools.Call(addr, "Transfer.Update", transItems, resp)
					pfc.Counter(host, -1)

					if err == nil {
						sendOk = true
						// statistics
					} else {
						// statistics
						g.Logger.Printf("transfer update fail, items size:%d, error:%v, resp:%v", len(transItems), err, resp)
					}
				}
			}

			// statistics
			if !sendOk {
				proc.SendToTransferFailCnt.IncrBy(int64(count))
			} else {
				proc.SendToTransferCnt.IncrBy(int64(count))
			}
		}(transItems, count)
	}
}
