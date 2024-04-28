package falcon

import (
	"fmt"
	"github.com/open-falcon/falcon-plus/common/redisdb"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/g"
	"sync"
	"time"
)



func GetRedisHostList() ( hosts []string, err error) {
	service := "agent.alive"
	hosts, err = redisdb.RedisServiceScan(service)
	if err != nil {
		g.Logger.Errorf("GetRedisHostList() err:%s", err)
		return
	}

	if g.Config().Debug {
		g.Logger.Debugf("GetRedisHostList() hosts length is:%d", len(hosts))
	}

	return
}

func GetRedisHostsExpire() {

	for {

			var tempCache g.LruCache

			service := "agent.alive"
			timeOut := int64(g.Config().AgentExpire)
			normalHost, expireHost, alarm, err := redisdb.RedisServiceExprieScan(service, timeOut)

			if err != nil && alarm == true {
				SendRedisAlarm(fmt.Sprintf("%s", err))
				time.Sleep(time.Second * time.Duration(g.Config().CheckInterval))
				continue
			}

			g.RedisNormalHost = normalHost

			if g.Config().Debug {
				g.Logger.Debugf("GetRedisHostList() Redis 中正常上报 falcon alive 主机 Total: %d", len(normalHost))
				g.Logger.Debugf("GetRedisHostList() Redis 中过期 falcon alive 主机 Total: %d", len(expireHost))
			}

			/*  对过期主机执行删除操作，并删除 redis 记录  terry.zeng
			*/

			tempCache.HostList = expireHost
			tempCache.Timestamp = time.Now().Unix()

			g.TotalLru = g.PutData(g.TotalLru, tempCache)

			if len(expireHost) > 0 {

				task_chan := make(chan bool, 10)
				wg := sync.WaitGroup{}
				defer close(task_chan)

				for _, host := range expireHost {

					wg.Add(1)
					task_chan <- true

					go func(host string) {

						<-task_chan

						if g.Config().Debug {
							g.Logger.Debugf("GetRedisHostsExpire() 删除 redis 记录 host:%s", host)
						}

						err := redisdb.RedisServerDelete(service, host)

						if err != nil {
							g.Logger.Debugf("GetRedisHostsExpire() 删除 redis 主机记录错误  host:%s error:%s", host, err)
						}

						if g.SkipAlarm == false {
							SendAlarm(host,"")
						}

						defer wg.Done()
					}(host)
				}
			}

			if g.Config().Debug == true {
				g.Logger.Debugf("[告警信息] 在 %d 个时间窗口内已经发生过 %d 次告警",
					(g.Config().Degrade.Period+1), g.GetHostCount(g.TotalLru))
			}

			if g.SkipAlarm == true {
				var hostDetail []string
				for _, info := range g.TotalLru {
					hostDetail = append(hostDetail, info.HostList...)
				}

				if len(hostDetail) > 0 {
					SendInternalAlarm(hostDetail)
				}

				// 如果希望在降级期间保留所有的主机 list 则关闭下面 g.TotalLru 方法
				// 当前希望降级期间每分钟都清空主机列表一次 (每分钟都有独立的主机记录日志)
				g.TotalLru = make(map[int]g.LruCache, g.Config().Degrade.Period)
			}

			time.Sleep(time.Second * time.Duration(g.Config().CheckInterval))
	}
}

