package http

import (
	"github.com/signmem/falcon-plus/modules/judge/proc"
	"github.com/signmem/falcon-plus/modules/judge/store"
	"net/http"
	"time"
)


type MetricResponse struct {
	Name 		string 		`json:"name"`
	Value 		float64		`json:"value"`
	Time 		string 		`json:"time"`
}



func metricGet() {
	http.HandleFunc("/counter/all",
		func(w http.ResponseWriter, r *http.Request) {
			ret := proc.GetAll()
			bigMapValue := store.GetBigMapCount()
			revc := proc.GenMap("RevcTransferMetric", bigMapValue)
			ret = append(ret, revc)
			RenderDataJson(w, ret)
			return
		})
}


func metricCheck() {
	http.HandleFunc("/count/all",
		func(w http.ResponseWriter, r *http.Request) {
			var localProc []*proc.SCount
			localProc = append(localProc, proc.SendToRedisCntTotal)
			localProc = append(localProc, proc.SendToRedisCntDrop)
			localProc = append(localProc, proc.SendToRedisCntSuccess)

			var metricRespo []MetricResponse

			for _, procName := range localProc {
				var metricProc MetricResponse

				info := procName.Get()
				metricProc.Name =  info.Name
				metricProc.Value = info.Cnt
				// timestamp := info.Time
				t := time.Now()
				metricProc.Time = t.Format("2006-01-02 15:04:05")

				metricRespo = append(metricRespo,metricProc)
			}

			RenderJson(w, metricRespo)
			return
		})
}
