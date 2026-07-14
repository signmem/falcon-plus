package main

import (
	"flag"
	"fmt"
	"github.com/signmem/falcon-plus/common/redisdb"
	"github.com/signmem/falcon-plus/modules/pingcheck/cmdb"
	"github.com/signmem/falcon-plus/modules/pingcheck/falcon"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"github.com/signmem/falcon-plus/modules/pingcheck/http"
	"github.com/signmem/falcon-plus/modules/pingcheck/selector"
	"os"
	"runtime"
)

func init() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")

	flag.Parse()

	if *version {
		version := g.Version
		fmt.Printf("pingcheck version: %s\n", version)
		os.Exit(0)
	}

	var err error

	g.ParseConfig(*cfg)
	g.Logger = g.InitLog()
	g.Alarmer = g.InitAlarmLog()

	redisdb.Server = g.Config().Redis.Server + ":" + g.Config().Redis.Port
	redisdb.MaxIdle = g.Config().Redis.MaxIdle
	redisdb.MaxActive = g.Config().Redis.MaxActive
	redisdb.IdleTimeOut = g.Config().Redis.IdleTimeOut
	redisdb.Pool = redisdb.NewPool(redisdb.MaxIdle, redisdb.MaxActive,
		redisdb.IdleTimeOut, redisdb.Server)
	redisdb.CleanupHook()

	g.FalconAgentLru.TotalLru = make(map[string][]string)
	g.FalconPingLru.TotalLru = make(map[string][]string)

	redisdb.Client, err = redisdb.RedisClient(redisdb.MaxIdle,redisdb.MaxActive,redisdb.Server)
	if err != nil {
		g.Logger.Fatalf("create redis client failed: %v", err)
	}
}

func main() {


	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	if g.Config().Redis.Enabled {
		// 用于检测 redis /agent.alive/host 主机过期入口 ( 一秒一次 ) 
		go falcon.GetRedisHostsExpire() 
	}

	go http.Start()

	// 从 cmdbv3 入口获取所有物理机入口
	go cmdb.GetCMDBHostLoop()

	go falcon.CompireCmdbAndRedis()	// 检测没有运行 falcon-agent 的入口

	go falcon.CheckReplicaHostname() //  alarm multi hostname
	go selector.Start()			// use to get redis lock


	// PingCheckHost PingFalseRecheck 需要获取 select.Role == master 才可以执行
	go falcon.PingCheckHost()	  // pingcheck 主流程，只检测 cmdb.PingCheckRecord 中包含的主机
	go falcon.PingFalseRecheck()  // 15s 一次从 /falcon.ping 中检测主机，after interval 无法 ping 则告警
    

	if g.Config().Degrade.Enabled  {
		go g.MonitorAgentPeriod()	// 降级 agent 控制
		go g.MonitorPingPeriod()	// 降级 ping 控制
	}

	// send metrics to falcon agent
	go falcon.UploadMetric()

	// check transfer health  alarm needed
	go http.CheckTransfer()

	// generate agent.ping metrics
	go falcon.FalconPingMetrics()

	g.Logger.Debugf("Program started with %d CPUs", numCPU)

	select {}
}
