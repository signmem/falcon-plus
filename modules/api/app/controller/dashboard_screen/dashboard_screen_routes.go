package dashboard_screen

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
	//for unauth GET requests
	unauthapi := r.Group("/api/v1/dashboard_screen")
	unauthapi.GET("/getall", ScreenAll)
	unauthapi.GET("/get/:screen_name", ParentScreenGet)
	unauthapi.GET("/get_byname/:screen_name", ScreenGetByName)
	unauthapi.GET("/get_byip/:screen_ip", ScreenGetByIP)
	unauthapi.GET("/redirect/:screen_name", ScreenRedirect)

	// for auth API
	authapi := r.Group("/api/v1/dashboard")
	authapi.Use(utils.AuthSessionMidd)
	authapi.POST("/screen", ScreenCreate)
	authapi.GET("/screen/:screen_id", ScreenGet)
	authapi.GET("/screenall", ScreenAll)
	authapi.GET("/screen_byname/:screen_name", ScreenGetByName)
	authapi.GET("/screen_byip/:screen_ip", ScreenGetByIP)
	authapi.GET("/screens/pid/:pid", ScreenGetsByPid)
	authapi.GET("/screens", ScreenGetsAll)
	authapi.DELETE("/screen/:screen_id", ScreenDelete)
	authapi.PUT("/screen/:screen_id", ScreenUpdate)
}
