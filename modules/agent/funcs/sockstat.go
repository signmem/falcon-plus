package funcs

import (
	"github.com/open-falcon/falcon-plus/common/model"
	"log"
)

//SocketStatSummaryMetrics  rewrite ss function skip access /proc/slabinfo

func SocketStatSummaryMetrics() (L []*model.MetricValue) {
	ssMap, err := getSS()
	if err != nil {
		log.Println("SocketStatSummaryMetrics() error ", err)
		return
	}

	for k, v := range ssMap {
		L = append(L, GaugeValue("ss."+k, v))
	}

	return
}
