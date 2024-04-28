package register

import (
	"github.com/open-falcon/falcon-plus/common/redisdb"
	"github.com/open-falcon/falcon-plus/common/utils"
	"github.com/open-falcon/falcon-plus/modules/trend/proc"
	"github.com/open-falcon/falcon-plus/modules/trend/g"
	"time"
)

func RegLocalService() {
	if g.Ipaddr == "" {
		g.Ipaddr = utils.GetLocalIp()
		g.Logger.Debugf("RegLocalService() get localIp ip is %s", g.Ipaddr)
	}

	if g.Ipaddr == "" {
		return
	}

	metricPort := g.Config().MetricPort
	localAddr := g.Ipaddr + ":" + metricPort

	_, err := redisdb.RedisServiceWrite("trend", localAddr)

	if err != nil {
		g.Logger.Errorf("RegLocalService() write to redis err:%s", err)
		proc.WriteToRedisFailCnt.Incr()
	} else {
		proc.WriteToRedisCnt.Incr()
	}


}

func RegCron() {
	for {
		RegLocalService()
		time.Sleep(time.Second * time.Duration(60))
	}
}