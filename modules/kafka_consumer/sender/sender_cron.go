package sender

import (
	"strings"
	"time"

	"github.com/open-falcon/falcon-plus/modules/kafka_consumer/g"
	"github.com/open-falcon/falcon-plus/modules/kafka_consumer/proc"
	"github.com/toolkits/container/list"
)

const (
	DefaultProcCronPeriod = time.Duration(5) * time.Second    //ProcCron的周期,默认1s
	DefaultLogCronPeriod  = time.Duration(3600) * time.Second //LogCron的周期,默认300s
)

// send_cron程序入口
func startSenderCron() {
	go startProcCron()
	go startLogCron()
}

func startProcCron() {
	for {
		time.Sleep(DefaultProcCronPeriod)
		refreshSendingCacheSize()
	}
}

func startLogCron() {
	for {
		time.Sleep(DefaultLogCronPeriod)
		logConnPoolsProc()
	}
}

func refreshSendingCacheSize() {
	cfg := g.Config()
	proc.TrendQueuesCnt.SetCnt(calcSendCacheSize(TrendQueues))
	if cfg.Transfer.Enabled {
		proc.TransferQueuesCnt.SetCnt(int64(TransferQueue.Len()))
	}
}

func calcSendCacheSize(mapList map[string]*list.SafeListLimited) int64 {
	var cnt int64 = 0
	for _, list := range mapList {
		if list != nil {
			cnt += int64(list.Len())
		}
	}
	return cnt
}

func logConnPoolsProc() {
	g.Logger.Printf("connPools proc: \n%v", strings.Join(TrendConnPools.Proc(), "\n"))
}
