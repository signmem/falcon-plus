package falcon

import (
	"github.com/signmem/falcon-plus/common/redisdb"
	"github.com/signmem/falcon-plus/modules/pingcheck/cmdb"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"github.com/signmem/falcon-plus/modules/pingcheck/selector"
	"time"
)

func CompireCmdbAndRedis() {

	// 检测 cmdb 中主机信息没有上报至 redis 中的服务器信息
	// 避免启动时候 redis 没有信息，建议 sleep 3 mins
	// 当配置文件中 ForceCheck == true 则纯检， 不告警，用于测试
	// 当配置文件中 Redis.Enabled == true 时才可以检测出当前有多少 falcon agent 没有启动

	// all falcon check = cmdb.FalconCheckRecord


	for {

		if cmdb.FistTime == true  {
			time.Sleep(60 * time.Second)
			continue
		}

		if selector.Role == "master" {

			service := "agent.alive"
			hostList, err := redisdb.RedisServiceScan(service)
			if err != nil {
				g.Logger.Errorf("scan agent.alive failed: %v", err)
				time.Sleep(60 * time.Second)
				continue
			}

			noFalconHost := cmdb.FalconNotReport

			if len(noFalconHost) == 0 || len(hostList) == 0 {

				time.Sleep(60 * time.Second)

			} else {

				g.Logger.Debugf("[CMDB AGENT PREPARE] 没有启动 Falcon Agent " +
					"Total %d", len(noFalconHost))

				g.Logger.Debugf("[CMDB DEBUG FALCON AGENT NOT REPORT] %v", noFalconHost)
				g.Logger.Debugf("[CMDB AGENT PREPARE] 当前 Falcon Agent " +
					"上报服务器数量 Total %d", len(hostList))


				g.FalconAgentCMDBReportFailure.Set(len(noFalconHost))

				if len(noFalconHost) > 0 {

					for _, host := range noFalconHost {

						if g.Config().AlarmEnable == true {
							// 只有 force check == false 才会发送报警信息
							SendAlarm(host, "", false, "cmdb")
							// 为了预防报警降级，因此每次报警 sleep 2 秒
							time.Sleep(200 * time.Millisecond)
						}
					}
				}
				time.Sleep(60 * time.Minute)
			}
		}

	}
}
