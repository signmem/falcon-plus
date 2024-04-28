package falcon

import (
	"github.com/open-falcon/falcon-plus/modules/pingcheck/cmdb"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/g"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/tools"
	"time"
)

func CompireCmdbAndRedis() {

	// 检测 cmdb 中主机信息没有上报至 redis 中的服务器信息
	// 避免启动时候 redis 没有信息，建议 sleep 3 mins

	time.Sleep( 180 * time.Second)

	var runTime tools.TimeStruct

	runTime.Hour = "15"
	runTime.Minite = "30"

	allowWeek := []int{1,2,3,4,5}


	for {

		redisHostList := g.RedisNormalHost  // redis 中已经上报的主机

		t := time.Now()
		week := int(t.Weekday())
		if tools.IntInSlice(week, allowWeek) || g.Config().ForceCheck == true {

			nowTime := tools.GetNow()

			if runTime == nowTime || g.Config().ForceCheck == true {

				g.Logger.Debugf("CompireCmdbAndRedis() going to Compair info. firsttime %t," +
					" cmdbrecord: %d, redisrecord: %d", cmdb.FistTime, len(cmdb.CmdbHostInfoRecord), len(redisHostList) )

				if cmdb.FistTime == true || len(cmdb.CmdbHostInfoRecord) == 0 || len(redisHostList) == 0 {
					time.Sleep( 60 * time.Second)
				} else {

					g.Logger.Debugf("CompireCmdbAndRedis() CMDB 需要检测服务器数量 Total %d", len(cmdb.CmdbHostInfoRecord))
					g.Logger.Debugf("CompireCmdbAndRedis() REDIS 当前上报服务器数量 Total %d", len(redisHostList))

					var cmdbHostList []string
					for _, hostinfo := range cmdb.CmdbHostInfoRecord {
						cmdbHostList = append(cmdbHostList, hostinfo.HostName)
					}

					noFalconHost := tools.GetDstSliceNotInSrcSlice(redisHostList, cmdbHostList)

					if g.Config().Debug {
						g.Logger.Debugf("CompireCmdbAndRedis() 没有启动 falcon total %d", len(noFalconHost))
					}

					if len(noFalconHost) > 0 {
						for _, host := range noFalconHost {
							if g.Config().AlarmEnable == true {
								SendAlarm(host,"")   // open it when produce terry.zeng
							}

							if g.Config().Debug {
								g.Logger.Warningf("CompireCmdbAndRedis() falcon dead host: %s ", host)
							}
						}
					}

					time.Sleep(3600 * time.Second)
				}

			} else {
				time.Sleep(60 * time.Second)
			}

		} else {
			time.Sleep(60 * time.Second)
		}

	}
}