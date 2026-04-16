package cron

import (
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"github.com/signmem/falcon-plus/modules/agent/funcs"
	"time"
)

func AgentAlive() {

	agentInterval := 15

	if !g.Config().Transfer.Enabled {
		return
	}

	if len(g.Config().Transfer.Addrs) == 0 {
		return
	}

	for {
		time.Sleep( time.Duration(agentInterval) * time.Second)
		var mappers = []funcs.FuncsAndInterval{
			{
				Fs: []func() []*model.MetricValue {
						funcs.AgentMetrics,
				},
				Interval: agentInterval,
			},
		}

		for _, v := range mappers {
			collect(int64(agentInterval), v.Fs)
		}
	}

}
