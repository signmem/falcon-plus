package http

import (
	"encoding/json"
	"net/http"

	cmodel "github.com/signmem/falcon-plus/common/model"

	trpc "github.com/signmem/falcon-plus/modules/gateway/receiver/rpc"
)

func configApiHttpRoutes() {
	http.HandleFunc("/api/push", func(w http.ResponseWriter, req *http.Request) {
		if req.ContentLength == 0 {
			http.Error(w, "blank body", http.StatusBadRequest)
			return
		}

		decoder := json.NewDecoder(req.Body)
		var metrics []*cmodel.MetricValue
		err := decoder.Decode(&metrics)
		if err != nil {
			http.Error(w, "decode error", http.StatusBadRequest)
			return
		}

		reply := &cmodel.TransferResponse{}
		trpc.RecvMetricValues(metrics, reply, "http")

		RenderDataJson(w, reply)
	})
}
