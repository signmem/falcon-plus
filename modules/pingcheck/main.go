package main

import (
	"flag"
	"fmt"
	"github.com/open-falcon/falcon-plus/common/redisdb"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/cmdb"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/falcon"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/g"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/http"
	"os"
)

func init() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")

	flag.Parse()

	if *version {
		version := g.Version
		fmt.Printf("%s", version)
		os.Exit(0)
	}

	g.ParseConfig(*cfg)

	redisdb.Server = g.Config().Redis.Server + ":" + g.Config().Redis.Port
	redisdb.MaxIdle = g.Config().Redis.MaxIdle
	redisdb.MaxActive = g.Config().Redis.MaxActive
	redisdb.IdleTimeOut = g.Config().Redis.IdleTimeOut
	redisdb.Pool = redisdb.NewPool(redisdb.MaxIdle, redisdb.MaxActive,
		redisdb.IdleTimeOut, redisdb.Server)
	redisdb.CleanupHook()

}


func main() {

	g.Logger = g.InitLog()

	if g.Config().Redis.Enabled == true {
		go falcon.GetRedisHostsExpire() // 用于检测 redis /agent.alive/host 主机过期入口
	}

	go http.Start()

	go cmdb.GetCMDBHostEveryHour()    // 每小时从 cmdbv3 入口获取所有物理机入口

	go falcon.CompireCmdbAndRedis()    // 检测没有运行 falcon-agent 的入口

	if g.Config().Degrade.Enabled == true {
		go g.MonitorPeriod()
	}

	go http.CheckTransfer()          // check transfer health  alarm needed
	select {}
}
