package http

import (
	"github.com/gin-gonic/gin"
	"github.com/signmem/falcon-plus/modules/graph/proc"
)

func configProcRoutes() {
	// counter
	router.GET("/counter/all", func(c *gin.Context) {
		JSONR(c, 200, proc.GetAll())
	})
}
