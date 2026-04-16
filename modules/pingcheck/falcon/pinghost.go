package falcon

import (
	"github.com/signmem/falcon-plus/common/redisdb"
	"github.com/signmem/falcon-plus/modules/pingcheck/cmdb"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"github.com/signmem/falcon-plus/modules/pingcheck/net"
	"github.com/signmem/falcon-plus/modules/pingcheck/selector"
	"sync"
	"time"
)



func PingCheckHost() {

	for {

		falconRePingConcurrent := g.Config().FalconRePingConcurrent

		if selector.Role == "master" {

			t := time.Now().Unix()
			if t % int64(g.Config().CheckPingInterval) == 0 {

				if g.Config().Debug == true {
					start := time.Now().Format("2006-01-02 15:04:05")
					g.Logger.Debugf("[PING START] at %s", start)
				}

				if len(cmdb.PingCheckRecord) > 0 {

					allowPingRecord := getSkipPingHosts(cmdb.PingCheckRecord)

					g.Logger.Debugf("[PING] %d hosts", len(allowPingRecord))

					g.FalconPingTotal.Set(len(allowPingRecord))

					taskChan := make(chan struct{}, falconRePingConcurrent)
					var wg sync.WaitGroup

					// defer close(taskChan)
					for _, host := range allowPingRecord {

						hostCopy := host
						taskChan <- struct{}{}
						wg.Add(1)


						go func(h cmdb.HostInfo) {

							defer func() {
								<-taskChan
								wg.Done()
							}()

							checkHost(h)

						}(hostCopy)
					}
					wg.Wait()
					close(taskChan)
				}

				if g.Config().Debug == true {
					start := time.Now().Format("2006-01-02 15:04:05")
					g.Logger.Debugf("[PING END] at %s", start)
					g.Logger.Debugf("[PING CHECK] 在 %d 分钟内，发生了 %d " +
						"台服务器无法 ping 检测, 当前降级设定为 %v", g.Config().Degrade.Period,
						g.FalconPingLru.GetLength(0), g.SkipPingAlarm)
				}

				if g.SkipPingAlarm == true {
					if g.FalconPingLru.GetLength(2) > 0 {
						SendPingDegardAlarm()
					}
				}
			}
		}

		time.Sleep( time.Duration(1) * time.Second)
	}
}

func getSkipPingHosts(allHost []cmdb.HostInfo) (allowHost []cmdb.HostInfo) {

	// allHost == CMDB 所有 ping only 主机
	// allowHost == 排除了 ping pingdaed 主机

	// SkipCmdbPing = make([]string,0)
	// SkipCmdbPing = append(SkipCmdbPing, pingRetry...)
	// SkipCmdbPing = append(SkipCmdbPing, pingDieHard...)

	pingService := "falcon.ping"
	skipPingHost, _ := redisdb.RedisServiceScan(pingService)


	deadService := "falcon.pingdead"
	timeOut := int64(3600)
	skikDieHost, expireDieHost, _, _ := redisdb.RedisServiceExprieScan(deadService, timeOut)

	if len(expireDieHost) > 0 {

		// 对过期 pingdead 主机执行处理，需要重置处理 falcon agent 检测 (1小时处理报告 pigeon 一次)

		for _, hostInfo := range expireDieHost {
			err := redisdb.RedisServerDelete(deadService, hostInfo)

			if err != nil {
				g.Logger.Errorf("getSkipPingHosts() delete service: %s host: %s error:%s",
					deadService, hostInfo, err)
			}

			g.Logger.Infof("[REDIS DELETE] service:%s, host:%s succeful.",
				deadService, hostInfo)
		}
	}

	skipAllHost := append(skipPingHost, skikDieHost...)

	if len(skipAllHost) > 0 {
		allowHost = skipHostFromCmdb(skipAllHost, allHost)
	} else {
		allowHost =  allHost
	}

	return allowHost
}


func skipHostFromCmdb(hostList []string, allCmdbHost []cmdb.HostInfo) (pingAllowHost []cmdb.HostInfo) {

	if len(allCmdbHost) > 0 && len(hostList) > 0 {

		for _, host := range allCmdbHost {
			hostInfo := host.DomainName + "@" + host.HostName + "@" + host.IPAddr
			if g.SliceContains(hostInfo, hostList) == false {
				pingAllowHost = append(pingAllowHost, host)
			}
		}

	} else {
		return allCmdbHost
	}
	return  pingAllowHost
}


func checkHost(host cmdb.HostInfo) {

	// 只会被 PingCheckHost() 调用
	// 用于检测主机主机是否可 ping 
	// ping 通则不做任何处理
	// ping 不通则写入 /falcon.ping/domain@hostname@ipaddr

	pingStatus, _ := net.PingFromProxy(host.IPAddr)
	if pingStatus == false {

		if g.Config().Debug == true {
			g.Logger.Debugf("[PING CHECK FALSE] host:%s ip:%s ping false.", host.HostName, host.IPAddr,)
		}

		g.FalconPingFailuer.Incr()

		hostinfo := host.DomainName + "@" + host.HostName + "@" + host.IPAddr
		service := "falcon.ping"
		redisdb.RedisServiceWrite(service, hostinfo)
	} else {
		g.FalconPingSuccess.Incr()
	}

}

func reCheckHost(host cmdb.HostInfo) {

	pingStatus, _ := net.PingFromProxy(host.IPAddr)
	if pingStatus == true {

		if g.Config().Debug == true {
			g.Logger.Debugf("[PING CHECK RESTORE] host:%s ip:%s ping restore.", host.HostName, host.IPAddr,)
		}

		hostinfo := host.HostName + "@" + host.IPAddr
		service := "falcon.ping"
		err := redisdb.RedisServerDelete(service, hostinfo)

		if err != nil {
			fullKey := "/" + service + "/" + hostinfo
			g.Logger.Errorf("reCheckHost() delete %s error.", fullKey)
		}
	}
}
