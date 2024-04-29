package sender

import (
	backend "github.com/signmem/falcon-plus/common/backend_pool"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/g"
	nset "github.com/toolkits/container/set"
)

func initConnPools() {
	cfg := g.Config()

	if cfg.Trend.Enabled {
		trendInstances := nset.NewStringSet()
		for _, instance := range cfg.Trend.Cluster {
			if g.Config().DebugTraffer {
				// terrytsang
				g.Logger.Debugf("initConnPools() Trend cluster instance %s", instance)
			}
			trendInstances.Add(instance)
		}
		TrendConnPools = backend.CreateSafeRpcConnPools(cfg.Trend.MaxConns, cfg.Trend.MaxIdle,
			cfg.Trend.ConnTimeout, cfg.Trend.CallTimeout, trendInstances.ToSlice())

		// trendInstances.Delete(instance)  terrytsang
	}

	// transfer
	if cfg.Transfer.Enabled {
		// init transfer global configs
		addrs := make([]string, 0)
		for hn, addr := range cfg.Transfer.Cluster {
			TransferHostnames = append(TransferHostnames, hn)
			addrs = append(addrs, addr)
			TransferMap[hn] = addr
		}
		transferInstances := nset.NewSafeSet()
		for _, instance := range cfg.Transfer.Cluster {
			transferInstances.Add(instance)
		}
		TransferConnPools = backend.CreateSafeJsonrpcConnPools(cfg.Transfer.MaxConns, cfg.Transfer.MaxIdle,
			cfg.Transfer.ConnTimeout, cfg.Transfer.CallTimeout, transferInstances.ToSlice())
	}
}

func DestroyConnPools() {
	TrendConnPools.Destroy()
	TransferConnPools.Destroy()
}
