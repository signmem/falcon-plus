package http

import (
	"github.com/open-falcon/falcon-plus/modules/kafka_consumer/proc"
	"net/http"
)

func configProcHttpRoutes() {
	// counter
	http.HandleFunc("/counter/all", func(w http.ResponseWriter, r *http.Request) {
		RenderDataJson(w, proc.GetAll())
	})
}
