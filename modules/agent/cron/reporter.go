package cron

import (
	"fmt"
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"log"
	"strings"
	"time"
)

func ReportAgentStatus() {
	if g.Config().Heartbeat.Enabled && g.Config().Heartbeat.Addr != "" {
		go reportAgentStatus(time.Duration(g.Config().Heartbeat.Interval) * time.Second)
	}
}

func reportAgentStatus(interval time.Duration) {
	for {
		// use to replace space in hostname strings . edit by terry.zeng
		tmpHostname, err := g.Hostname()
		replacer := strings.NewReplacer( " ", "", "\t", "" )
		hostname := replacer.Replace(tmpHostname)

		if err != nil {
			hostname = fmt.Sprintf("error:%s", err.Error())
		}

		req := model.AgentReportRequest{
			Hostname:      hostname,
			IP:            g.IP(),
			AgentVersion:  g.VERSION,
			PluginVersion: g.GetPluginVersion(),
		}

		var resp model.SimpleRpcResponse
		err = g.HbsClient.Call("Agent.ReportStatus", req, &resp)
		if err != nil || resp.Code != 0 {
			log.Println("call Agent.ReportStatus fail:", err, "Request:", req, "Response:", resp)
		}

		time.Sleep(interval)
	}
}
