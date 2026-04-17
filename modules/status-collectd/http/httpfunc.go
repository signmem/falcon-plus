package http

import (
	_ "encoding/json"
	"github.com/signmem/falcon-plus/modules/status-collectd/proc"
	"github.com/signmem/falcon-plus/modules/status-collectd/selector"
	_ "io/ioutil"
	_ "log"
	"net/http"
	"time"
)


func healthCheck() {
	http.HandleFunc("/_health_check", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
}

func metricCheck() {
	http.HandleFunc("/counter/all",
		func(w http.ResponseWriter, r *http.Request) {
			var localProc []*proc.SCount
			localProc = append(localProc, proc.SendToKafkaCntTotal)
			localProc = append(localProc, proc.SendToKafkaCntDrop)
			localProc = append(localProc, proc.SendToKafkaCntSuccess)

			var metricRespo []MetricResponse

			for _, procName := range localProc {
				var metricProc MetricResponse

				info := procName.Get()
				metricProc.Name =  info.Name
				metricProc.Value = info.Cnt
				timestamp := info.Time
				t := time.Unix(timestamp, 0)
				metricProc.Time = t.Format("2006-01-02 15:04:05")

				metricRespo = append(metricRespo,metricProc)
			}

			RenderJson(w, metricRespo)
			return
		})
}

func roleCheck() {
	http.HandleFunc("/role/check", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(selector.Role))
	})
}