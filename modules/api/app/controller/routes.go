package controller

import (
	"github.com/DeanThompson/ginpprof"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/signmem/falcon-plus/modules/api/app/controller/alarm"
	"github.com/signmem/falcon-plus/modules/api/app/controller/dashboard_graph"
	"github.com/signmem/falcon-plus/modules/api/app/controller/dashboard_screen"
	"github.com/signmem/falcon-plus/modules/api/app/controller/expression"
	"github.com/signmem/falcon-plus/modules/api/app/controller/graph"
	"github.com/signmem/falcon-plus/modules/api/app/controller/host"
	"github.com/signmem/falcon-plus/modules/api/app/controller/mockcfg"
	"github.com/signmem/falcon-plus/modules/api/app/controller/strategy"
	"github.com/signmem/falcon-plus/modules/api/app/controller/template"
	"github.com/signmem/falcon-plus/modules/api/app/controller/uic"
	"github.com/signmem/falcon-plus/modules/api/app/utils"
)

func StartGin(port string, r *gin.Engine) {
	r.Use(utils.CORS())
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, I'm Falcon+ (｡A｡)")
	})

	r.GET("/_health_check", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// automatically add routers for net/http/pprof
	// e.g. /debug/pprof, /debug/pprof/heap, etc.
	ginpprof.Wrap(r)

	graph.Routes(r)
	uic.Routes(r)
	template.Routes(r)
	strategy.Routes(r)
	host.Routes(r)
	expression.Routes(r)
	mockcfg.Routes(r)
	dashboard_graph.Routes(r)
	dashboard_screen.Routes(r)
	alarm.Routes(r)
	r.Run(port)
}
