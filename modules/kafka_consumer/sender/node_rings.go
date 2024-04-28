package sender

import (
	cutils "github.com/open-falcon/falcon-plus/common/utils"
	"github.com/open-falcon/falcon-plus/modules/kafka_consumer/g"
	rings "github.com/toolkits/consistent/rings"
)

func initNodeRings() {
	cfg := g.Config()

	TrendNodeRing = rings.NewConsistentHashNodesRing(int32(cfg.Trend.Replicas),
		cutils.KeysOfMap(cfg.Trend.Cluster))
	// cutils.KeysOfMap --> ["cluster-01", "cluster-02", "cluster-03"] terrytsang
}
