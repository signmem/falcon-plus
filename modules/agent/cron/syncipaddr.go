package cron

import (
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"log"
	"time"
)

func ReInitLocalIp() {

	// use to get ipaddr every day.
	// prevent IDC change ifcfg-eth0 and restart network
	// but g.LocalIp get old ip set to stable variable.

	for {

		nowTime := g.GetNow()
		if nowTime == g.JobTime {
			if g.Config().Debug == true {
				log.Println("[DEBUG] Local ip is: ", g.LocalIp )
			}
			g.InitLocalIp()
			g.InitPoolAndVMTags()   // reinitial g.POOLNAME  variable
		}
		time.Sleep(time.Second * 60)
	}
}

