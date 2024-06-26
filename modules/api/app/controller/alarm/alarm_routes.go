package alarm

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
	alarmapi := r.Group("/api/v1/alarm")
	alarmapi.Use(utils.AuthSessionMidd)
	alarmapi.POST("/eventcases", AlarmLists)
	alarmapi.GET("/eventcases", AlarmLists)
	alarmapi.POST("/events", EventsGet)
	alarmapi.GET("/events", EventsGet)
	alarmapi.GET("/GetEventCases", GetEventCases)
	alarmapi.GET("/GetEventCasesV2", GetEventCasesV2)
	alarmapi.GET("/GetTotalOfEventCases", GetTotalOfEventCases)
	alarmapi.POST("/event_note", AddNotesToAlarm)
	alarmapi.GET("/event_note", GetNotesOfAlarm)
	alarmapi.GET("/eventcases/total", GetEventCasesTotal)
	alarmapi.GET("/eventcases/detail", GetEventCasesDetail)
}
