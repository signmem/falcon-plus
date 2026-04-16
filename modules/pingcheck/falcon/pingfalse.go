package falcon

import (
	"fmt"
	"github.com/signmem/falcon-plus/common/redisdb"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"github.com/signmem/falcon-plus/modules/pingcheck/selector"
	"github.com/signmem/falcon-plus/modules/pingcheck/tools"
	"github.com/signmem/falcon-plus/modules/pingcheck/cmdb"
	"strings"
	"sync"
	"time"
)


var (
	//SkipCMDBPing 用于记录当前故障主机信息 "domain@host@ip" 
	// 每次从 cmdb 获取主机时候，都需要过滤这个 ping 
	SkipCmdbPing []string
)

func PingFalseRecheck() {

	for {

		if selector.Role == "master" {

			service := "falcon.ping"
			timeOut := int64(g.Config().PingDeadExpire)

			pingRetry, pingDieHard, alarm, err := redisdb.RedisServiceExprieScan(service, timeOut)

			// 访问 /falcon.ping/
			// 返回格式为 domain@hostname@ipaddress

			// pingDieHard == 已经过期，需要报警服务器, (redis delete already)
			// pingRetry == 还没有过期, 但无法 ping 通服务器，需要继续 ping check
			// alarm == redisdb.RedisServiceExprieScan 函数出错，需要告警
			// err == redisdb.RedisServiceExprieScan 函数错误信息

			// pingRetry and pingDieHard 格式 ==>  domain@hostname@ipaddress

			if alarm == true && err != nil {
				SendRedisAlarm(fmt.Sprintf("%s", err))
				time.Sleep(time.Second * time.Duration(60))
				continue
			}

			// 当 pingDieHard > 1 时， 需要执行收敛告警

			if len(pingDieHard) > 0 {

				g.FalconPingDie.Set(len(pingDieHard))

				appNameDieHard := mixPingFalseDomain(pingDieHard)

				// pingDieHard   /falcon.ping   已经过期的 主机信息
				// 1 执行报警
				// 2 从 /falcon.ping 删除对应主机
				// 3 写入 /falcon.pingdead  用于记录检测

				for domain, hostList := range appNameDieHard {

					//hostList = hostname@ipaddr  * notice **
					// terry.zeng  ( internal alarm enable )

					if g.SkipPingAlarm == false && g.Config().AlarmEnable == true {
						SendPingDieHardAlarm(domain, hostList) // 这里做了收敛告警
					}
					delRedisDiePingKey(domain, hostList)  // del   /falcon.ping/domain@hostname@ip
					putDiePingKey(domain, hostList)       // put   /falcon.pingdead/domain@hostname@ip
				}

			}


			if len(pingRetry) > 0 {

				g.Logger.Debugf("[REPING] going to reping %d hosts", len(pingRetry))

				task_chan := make(chan bool, 10)
				wg := sync.WaitGroup{}
				defer close(task_chan)
				for _, info := range pingRetry {

					var host cmdb.HostInfo
					hostinfo := strings.Split(info, "@")
					host.DomainName = hostinfo[0]
					host.HostName = hostinfo[1]
					host.IPAddr = hostinfo[2]
					wg.Add(1)
					task_chan <- true

					go func(host cmdb.HostInfo) {
						<-task_chan
						go reCheckHost(host)
						defer wg.Done()
					}(host)
				}
				wg.Wait()
			}
		}

		time.Sleep(15 * time.Second)
	}
}


func mixPingFalseDomain(redisInfo []string) (mixDomain map[string][]string) {

	// 组合所有无法 ping 信息 返回格式 gap[domain][hostname@ipaddr, hostname@ipaddr]
	// redisInfo == [ domain@hostname@ipaddress ]

	// gap return map[abc.com:[aaa@1.1.1.1 ] bbb.com:[aaa@2.2.2.1 bbb@2.2.2.2 ccc@2.2.2.3]]

	gap := make(map[string][]string)

	for _, info := range redisInfo {
		detail := strings.Split(info, "@")
		domain := detail[0]
		hostname := detail[1]
		ipaddr := detail[2]

		if tools.SliceContains(hostname + "@" + ipaddr, gap[domain]) == false {
			gap[domain] = append(gap[domain], hostname + "@" + ipaddr)
		}
	}
	return gap
}

func checkIpInDomain(gap map[string][]string, domain string) bool {

	for info, _ := range gap {
		if info == domain {
			return true
		}
	}
	return false
}

func delRedisDiePingKey(domain string, hostList []string) {

	// 删除  /falcon.ping/domain@hostname@ipaddr

	for _, info := range hostList {

		key := domain + "@" + info
		service := "falcon.ping"

		if g.Config().Debug {
			g.Logger.Debugf("[REPING] 删除 redis service: %s 记录 host:%s", service, key)
		}

		err := redisdb.RedisServerDelete(service, key)

		if err != nil {
			g.Logger.Debugf("[REPING] 删除 redis 主机记录错误 service: %s host:%s error:%s",
				service, key, err)
		}
	}
}

func putDiePingKey(domain string, hostList []string) {

	// 写入 redis /falcon.pingdead/domain@hostname@ipaddr 

	for _, hostInfo := range hostList {
		hostinfo := domain + "@" + hostInfo
		service := "falcon.pingdead"
		//
		g.Logger.Debugf("[REDIS WRITE] %s to %s", hostinfo, service)
		redisdb.RedisServiceWrite(service, hostinfo)
	}

}
