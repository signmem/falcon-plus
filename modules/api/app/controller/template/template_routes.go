package template

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/open-falcon/falcon-plus/modules/api/app/utils"
	"github.com/open-falcon/falcon-plus/modules/api/config"
)

var db config.DBPool

const badstatus = http.StatusBadRequest

func Routes(r *gin.Engine) {
	db = config.Con()
	tmpr := r.Group("/api/v1/template")
	tmpr_vip := r.Group("/api/vip/template")
	tmpr.Use(utils.AuthSessionMidd)
	tmpr_vip.Use(utils.AuthSessionMidd)
	tmpr_vip.GET("/:tpl_id", GetOneTemplateNameByID)
	tmpr.GET("", GetTemplates)
	tmpr.POST("", CreateTemplate)
	tmpr.GET("/:tpl_id", GetATemplate)
	tmpr.PUT("", UpdateTemplate)
	tmpr.DELETE("/:tpl_id", DeleteTemplate)
	tmpr.POST("/action", CreateActionToTmplate)
	tmpr.PUT("/action", UpdateActionToTmplate)
	tmpr.POST("/owner_update", UpdateTemplateOwner)

	actr := r.Group("/api/v1/action")
	actr.GET("/:act_id", GetActionByID)

	//simple list for ajax use
	tmpr2 := r.Group("/api/v1/template_simple")
	tmpr.Use(utils.AuthSessionMidd)
	tmpr2.GET("", GetTemplatesSimple)
}
