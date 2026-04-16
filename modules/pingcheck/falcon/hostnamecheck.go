package falcon

import (
	"github.com/signmem/falcon-plus/modules/pingcheck/cmdb"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"github.com/signmem/falcon-plus/modules/pingcheck/selector"
	"strconv"
	"time"
)

func CheckReplicaHostname() {

	allowWeek := []int{1,2,3,4,5}

	for {
		if selector.Role == "master" {

			t := time.Now()
			week := int(t.Weekday())

			if  g.IntInSlice(week, allowWeek) == true && timeNowAllow() == true {

				if len(cmdb.ReplicateHostInfo) == 0 {
					time.Sleep( 1 * time.Minute)
					continue
				}

				g.FalconCMDBHostNameReplicate.Set(len(cmdb.ReplicateHostInfo))

				for hostname, ipaddr := range cmdb.ReplicateHostInfo {
					sendMultiHostnameAlarm(hostname, ipaddr)
				}
			}
		}
		time.Sleep( 60 * time.Minute)
	}
}


func timeNowAllow() bool {

	hour, _ :=  strconv.Atoi(time.Now().Format("15"))
	if hour > 8 && hour < 17 {
		return true
	}
	return false
}
