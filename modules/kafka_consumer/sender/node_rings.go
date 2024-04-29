package sender

import (
	cutils "github.com/signmem/falcon-plus/common/utils"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/g"
	rings "github.com/signmem/consistent/rings"
)

func initNodeRings() {
	cfg := g.Config()

	TrendNodeRing = rings.NewConsistentHashNodesRing(int32(cfg.Trend.Replicas),
		cutils.KeysOfMap(cfg.Trend.Cluster))
	// cutils.KeysOfMap --> ["cluster-01", "cluster-02", "cluster-03"] terrytsang
}
