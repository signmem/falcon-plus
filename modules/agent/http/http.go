package http

import (
	"encoding/json"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/signmem/falcon-plus/modules/agent/g"
)

type Dto struct {
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func init() {
	configAdminRoutes()
	configCpuRoutes()
	configDfRoutes()
	configHealthRoutes()
	configIoStatRoutes()
	configKernelRoutes()
	configMemoryRoutes()
	configPageRoutes()
	configPluginRoutes()
	configPushRoutes()
	// configRunRoutes()
	configSystemRoutes()
	configCronRoutes()
	configShowIp()
}

func RenderJson(w http.ResponseWriter, v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(bs)
}

func RenderDataJson(w http.ResponseWriter, data interface{}) {
	RenderJson(w, Dto{Msg: "success", Data: data})
}

func RenderMsgJson(w http.ResponseWriter, msg string) {
	RenderJson(w, map[string]string{"msg": msg})
}

func AutoRender(w http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		RenderMsgJson(w, err.Error())
		return
	}

	RenderDataJson(w, data)
}

func Start() {
	if !g.Config().Http.Enabled {
		return
	}

	addr := g.Config().Http.Listen

	if addr == "" || strings.Split(addr, ":")[0] != "" {
		return
	} else {
		addr = g.LocalIp + addr
	}

	log.Println(addr)
	s := &http.Server{
		Addr:           addr,
		MaxHeaderBytes: 1 << 30,
	}

	log.Println("listening", addr)
	log.Fatalln(s.ListenAndServe())
}

func StartLoopback() {
	if !g.Config().Http.Enabled {
		return
	}


	addr_loop := "127.0.0.1" + g.Config().Http.Listen

	log.Println(addr_loop)

	s_loop := &http.Server{
		Addr:           addr_loop,
		MaxHeaderBytes: 1 << 30,
	}

	log.Println("listening", addr_loop)
	log.Fatalln(s_loop.ListenAndServe())
}
