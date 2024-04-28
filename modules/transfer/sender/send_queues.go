package sender

import (
	"github.com/open-falcon/falcon-plus/modules/transfer/g"
	nlist "github.com/toolkits/container/list"
)

func initSendQueues() {
	cfg := g.Config()
	for node := range cfg.Judge.Cluster {
		Q := nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
		JudgeQueues[node] = Q
	}

	for node, nitem := range cfg.Graph.ClusterList {
		for _, addr := range nitem.Addrs {
			Q := nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
			GraphQueues[node+addr] = Q
		}
	}

	if cfg.Tsdb.Enabled {
		TsdbQueue = nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
	}

	//added by vincent.zhang for sending to kafka
	if cfg.Kafka.Enabled {
		KafkaQueue = nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
		// added by qimin.xu for init filter queue
		for f := range cfg.Kafka.Filter {
			Q := nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
			KafkaFilterQueues[f] = Q
		}
	}

	if cfg.Kafka.LogEnabled {
		KafkaLogQueue = nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
	}
}
