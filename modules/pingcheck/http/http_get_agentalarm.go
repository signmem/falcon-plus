package http

import (
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"net/http"
)

func agentAlarm() {
	http.HandleFunc("/api/v1/agentcheck",	func(w http.ResponseWriter, r *http.Request) {

		var alarmInfo []AgentAlarm
		for time, hostList := range g.FalconAgentLru.TotalLru {
			var alarm  AgentAlarm
			alarm.TimeStamp = time
			alarm.HostList = hostList
			alarmInfo = append(alarmInfo, alarm)
		}

		g.Logger.Debugf("http get agent info: %s", g.FalconAgentLru.GetLruDetail(0))
		RenderJson(w, alarmInfo)
		return

	})
}

func pingAlarm() {
	http.HandleFunc("/api/v1/pingcheck",	func(w http.ResponseWriter, r *http.Request) {

		var alarmInfo []AgentAlarm
		for time, hostList := range g.FalconPingLru.TotalLru {
			var alarm  AgentAlarm
			alarm.TimeStamp = time
			alarm.HostList = hostList
			alarmInfo = append(alarmInfo, alarm)
		}

		g.Logger.Debugf("http get ping check info: %s", g.FalconPingLru.GetLruDetail(0))
		RenderJson(w, alarmInfo)
		return

	})
}