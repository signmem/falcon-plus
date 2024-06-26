package http

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/signmem/falcon-plus/modules/graph/g"
	"github.com/signmem/falcon-plus/modules/graph/rrdtool"
)

type Dto struct {
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

var Close_chan, Close_done_chan chan int
var router *gin.Engine

func init() {
	router = gin.Default()
	configCommonRoutes()
	configProcRoutes()
	configIndexRoutes()
	Close_chan = make(chan int, 1)
	Close_done_chan = make(chan int, 1)

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

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type TcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln TcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func Start() {
	if !g.Config().Http.Enabled {
		log.Println("http.Start warning, not enabled")
		return
	}

	router.GET("/api/v2/counter/migrate", func(c *gin.Context) {
		cnt := rrdtool.GetCounter()
		log.Debug("migrating counter:", cnt)
		c.JSON(200, gin.H{"msg": "ok", "counter": cnt})
	})

	//compatible with signmem v0.1
	router.GET("/counter/migrate", func(c *gin.Context) {
		cnt := rrdtool.GetCounter()
		log.Debug("migrating counter:", cnt)
		c.JSON(200, gin.H{"msg": "ok", "counter": cnt})
	})

	addr := g.Config().Http.Listen
	if addr == "" {
		return
	}
	go router.Run(addr)

	select {
	case <-Close_chan:
		log.Info("http recv sigout and exit...")
		Close_done_chan <- 1
		return
	}

}
