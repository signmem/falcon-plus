package cmdb

import (
	"strings"
	"github.com/signmem/falcon-plus/common/redisdb"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
)

func getRedisSkipPingHost() (hostName []string) {

	// 功能 获取当前不需要执行常规 ping 的服务器 list

	services := []string{"falcon.ping", "falcon.pingdead"}

	for _, service := range services {

		hostList, err := redisdb.RedisServiceScan(service)

		if err != nil {
			g.Logger.Errorf("getRedisSkipPingHost() 检测 RedisServiceScan service:%s " +
				"故障 %s ", service, err)
		}

		if len(hostList) > 0 {
			for _, host := range hostList  {
				detail := strings.Split(host, "@")
				if len(detail) < 3 {
					g.Logger.Errorf("invalid redis skip ping host format, raw=%s", host)
					continue
				}
				name := detail[2]
				hostName = append(hostName, name)
			}
		}
	}

	return hostName
}


func getAgentAvlive() (hosts []string, err error) {
	service := "agent.alive"
	hosts, err = redisdb.RedisServiceScan(service)
	if err != nil {
		g.Logger.Errorf("getAgentAvlive() err:%s", err)
		return
	}

	if g.Config().Debug {
		g.Logger.Debugf("getAgentAvlive() hosts length is:%d", len(hosts))
	}

	return
}