package http

import (
	"github.com/signmem/falcon-plus/modules/trend/proc"
	"io"
	"net/http"
)

func configProcHttpRoutes() {
	// counter
	http.HandleFunc("/counter/all", func(w http.ResponseWriter, r *http.Request) {
		RenderDataJson(w, proc.GetAll())
	})
}

func healthCheck() {
	http.HandleFunc("/_health_check", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok")
	})
}