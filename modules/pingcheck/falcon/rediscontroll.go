package falcon

import (
	"fmt"
	"github.com/signmem/falcon-plus/common/redisdb"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"sync"
	"time"
)


func GetRedisHostsExpire() {

	g.FalconAgentLru.Init()

	for {

		if time.Now().Unix() % int64(g.Config().CheckInterval) == 0 {

			g.Logger.Debugf("[Redis Check Agent Alive] start")
			g.LruMaintain()     // init lru and remove out of lru

			service := "agent.alive"
			timeOut := int64(g.Config().AgentExpire)
			normalHost, expireHost, redisAlarm, err := redisdb.RedisServiceExprieScan(service, timeOut)


			// redisAlarm 意味着 redisdb.RedisServiceExprieScan 函数产生故障，需要告警 redis 访问出错信息
			if redisAlarm  {
				SendRedisAlarm(fmt.Sprintf("%s", err))
				g.Logger.Errorf("[Redis Check Agent Alive] service scan failed: %v", err)
				// time.Sleep(time.Second * time.Duration(g.Config().CheckInterval))
				continue
			}

			if err != nil {
				g.Logger.Debugf("[Redis Check Agent Alive Debug] %s", err)
				// time.Sleep(time.Second * time.Duration(g.Config().CheckInterval))
				continue
			}

			g.RedisNormalHost = normalHost

			if g.Config().Debug {
				g.Logger.Debugf("[AGENT CHECK] Redis 中正常上报 falcon alive 主机 Total: %d", len(normalHost))
				g.Logger.Debugf("[AGENT CHECK] Redis 中过期 falcon alive 主机 Total: %d", len(expireHost))
			}

			g.FalconAgentAliveReportSuccess.Set(len(normalHost))
			g.FalconAgentAliveReportTimeOut.Set(len(expireHost))

			// 清理 redis 中过期的主机信息
			// 意图: 120 秒不上报 falcon agent 人为 agent 故障
			// 发送 pigeon 告警信息  SendAlarm()

			if len(expireHost) > 0 {

				taskChan := make(chan struct{}, 10)
				var wg sync.WaitGroup

				for _, host := range expireHost {

					hostCopy := host

					taskChan <- struct{}{}
					wg.Add(1)

					go func(h string) {

						defer func() {
							<-taskChan
							wg.Done()
						}()

						if g.Config().Debug {
							g.Logger.Debugf("[AGENT CHECK] - [REDIS DELETE] 删除 redis 记录 host:%s", host)
						}

						err := redisdb.RedisServerDelete(service, h)

						if err != nil {
							g.Logger.Debugf("[AGENT CHECK] 删除 redis 主机记录错误  host:%s error:%s", host, err)
						}

						if g.Config().AlarmEnable == true && g.SkipAgentAlarm == false {
							SendAlarm(h,"", false, "redis")
						}

					}(hostCopy)
				}
				wg.Wait()
				close(taskChan)
			}

			// terry.zeng 计算时间窗口内告警次数
			alarmTotalCount := g.FalconAgentLru.GetLength(0)

			g.FalconAgentAliveAlarmCounter.Set(alarmTotalCount)

			if g.Config().Debug == true {
				g.Logger.Debugf("[告警信息 Total] 在 %d 个时间窗口内已经发生过 %d 次告警",
					g.Config().Degrade.Period, alarmTotalCount )
			}

			// terry.zeng
			if g.SkipAgentAlarm == true {
				g.FalconAgentAliveAlarmDegarded.Set(1)

				if alarmTotalCount  > 0 {
					SendInternalAlarm()
				}
			} else {
				g.FalconAgentAliveAlarmDegarded.Set(0)
			}
			g.Logger.Debugf("[Redis Check Agent Alive end.]")
		}

		time.Sleep( time.Second * 1 )
	}
}

