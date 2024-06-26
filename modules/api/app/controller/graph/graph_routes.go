package graph

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/signmem/falcon-plus/modules/api/app/utils"
	"github.com/signmem/falcon-plus/modules/api/config"
)

var db config.DBPool

const badstatus = http.StatusBadRequest
const expecstatus = http.StatusExpectationFailed

func Routes(r *gin.Engine) {
	db = config.Con()
	authapi := r.Group("/api/v1")
	authapi.Use(utils.AuthSessionMidd)
	authapi.GET("/graph/endpointobj", EndpointObjGet) //added by vincent.zhang for screen bug of dashboard
	authapi.GET("/graph/endpoint", EndpointRegexpQuery)
	authapi.GET("/graph/endpoint_counter", EndpointCounterRegexpQuery)
	authapi.GET("/graph/endpoint_counter_distinct", EndpointCounterDistinct)
	authapi.GET("/graph/metrics/:ip", EndpointCounterlistByIP)
	authapi.POST("/graph/history", QueryGraphDrawData)
	authapi.POST("/graph/lastpoint", QueryGraphLastPoint)
	authapi.DELETE("/graph/endpoint", DeleteGraphEndpoint)
	authapi.DELETE("/graph/counter", DeleteGraphCounter)

	grfanaapi := r.Group("/api")
	grfanaapi.GET("/v1/grafana", GrafanaMainQuery)
	grfanaapi.GET("/v1/grafana/metrics/find", GrafanaMainQuery)
	grfanaapi.POST("/v1/grafana/render", GrafanaRender)
	grfanaapi.GET("/v1/grafana/render", GrafanaRender)

}
