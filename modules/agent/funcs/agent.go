package funcs

import (
	"github.com/signmem/falcon-plus/common/model"
)

func AgentMetrics() []*model.MetricValue {
	return []*model.MetricValue{GaugeValue("agent.alive", 1)}
}


func PingMetrics() []*model.MetricValue {
	return []*model.MetricValue{GaugeValue("agent.ping", 1)}
}
