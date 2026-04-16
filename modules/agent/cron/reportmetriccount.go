package cron

import (
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"log"
	"time"
)

func GetMetricCount()  {

	debug := g.Config().Debug

	mvs := model.MetricValue{}
	mvs.Metric = "uploadmetric.count"
	mvs.Step = 60
	mvs.Type = "GAUGE"

	for {
		mvs.Tags = ""
		mvs.Endpoint = g.HostName
		reportMVS := []*model.MetricValue{}
		metricCount := g.ReportMetricCounts()

		now := time.Now().Unix()
		mvs.Value = metricCount
		mvs.Timestamp = now

		reportMVS = append(reportMVS, &mvs)

		if debug {
			log.Println("total metric count: ", metricCount )
		}

		g.SendToTransfer(reportMVS)

		g.SetMetricCount(0)
		time.Sleep(time.Second * 60)
	}

}