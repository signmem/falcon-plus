package sender

import (
	"github.com/signmem/falcon-plus/modules/kafka_consumer/g"
	nlist "github.com/toolkits/container/list"
)

func initSendQueues() {
	cfg := g.Config()
	for node := range cfg.Trend.Cluster {
		Q := nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
		TrendQueues[node] = Q

		if g.Config().DebugTraffer {
			g.Logger.Debugf("initSendQueues() node is:%s", node)
			// terrytsang
		}
	}
	if cfg.Transfer.Enabled {
		TransferQueue = nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
	}
}
