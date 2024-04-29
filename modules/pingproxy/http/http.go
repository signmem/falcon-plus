package http

import (
	"github.com/signmem/falcon-plus/modules/pingproxy/g"
	"net/http"
	"log"
)

func Start() {
	go startHttpServer()
}

func startHttpServer() {
	if ! g.Config().Http.Enabled {
		return
	}

	addr := g.Config().Http.Listen
	port := g.Config().Http.Port
	listen_port := addr + ":" + port

	if listen_port == "" {
		return
	}

	configApiRoutes()

	s := &http.Server{
		Addr:           listen_port,
		MaxHeaderBytes: 1 << 30,
	}

	log.Println("http.startHttpServer ok, listening", addr)
	log.Fatalln(s.ListenAndServe())
}
