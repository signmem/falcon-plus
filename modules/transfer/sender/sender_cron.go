package sender

import (
	"github.com/open-falcon/falcon-plus/modules/transfer/g"
	"github.com/open-falcon/falcon-plus/modules/transfer/proc"
	"github.com/toolkits/container/list"
	"log"
	"strings"
	"time"
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
	proc.JudgeQueuesCnt.SetCnt(calcSendCacheSize(JudgeQueues))
	proc.GraphQueuesCnt.SetCnt(calcSendCacheSize(GraphQueues))
	//added by vincent.zhang for tsdb and kafka
	if cfg.Tsdb.Enabled {
		proc.TsdbQueuesCnt.SetCnt(int64(TsdbQueue.Len()))
	}
	if cfg.Kafka.LogEnabled {
		proc.KafkaLogQueuesCnt.SetCnt(int64(KafkaLogQueue.Len()))
	}
	//added by qimin.xu
	if cfg.Kafka.Enabled {
		proc.KafkaQueuesCnt.SetCnt(int64(KafkaQueue.Len()))
		if len(cfg.Kafka.Filter) > 0 {
			for f := range cfg.Kafka.Filter {
				proc.KafkaQueuesCntMap[f].SetCnt(int64(KafkaFilterQueues[f].Len()))
			}
		}
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
	log.Printf("connPools proc: \n%v", strings.Join(GraphConnPools.Proc(), "\n"))
	//added by vincent.zhang for sending to kafka and producer pool method is neccessary
	//log.Printf("kafkaProducerPool proc: \n%v", KafkaPool.Proc())
}
