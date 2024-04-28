package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/open-falcon/falcon-plus/modules/agent/cron"
	"github.com/open-falcon/falcon-plus/modules/agent/funcs"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"github.com/open-falcon/falcon-plus/modules/agent/http"
)

func main() {

	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	check := flag.Bool("check", false, "check collector")
	showpid := flag.Bool("showpid",false, "show process pid")

	flag.Parse()

	if *version {
		pluginInfo := g.GetVersionFileInfo()
		fmt.Printf("falcon-agent version is %s\n", g.VERSION)
		fmt.Printf("%s", pluginInfo)
		os.Exit(0)
	}

	if *check {
		funcs.CheckCollector()
		os.Exit(0)
	}

	if *showpid {
		pid, err := g.ReadPid()
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}else{
			fmt.Printf("%d\n", pid)
			os.Exit(0)
		}
	}

	g.ParseConfig(*cfg)
	g.VersionDir = g.Config().Plugin.Dir + "/version"
	g.VersionFile = g.VersionDir + "/version"
	g.InitLog()

	// g.WritePid()  ## write pid to pidfile setting. not necessary. terry.zeng
	g.InitRootDir()
	g.InitLocalIp()
	g.InitPoolAndVMTags()    // initial g.POOLNAME g.VirtualMachine variable
	g.CheckSwapMonitor()     // check swap monitor status  SkipSwapMonitor -> false ( do monitor )
	g.InitRpcClients()

	g.GenCronTime()
	log.Println("cron job is: ", g.JobTime.Hour, ":" , g.JobTime.Minite)

	funcs.BuildMappers()

	go cron.InitDataHistory()

	go cron.SyncPlugin()       // use to rsync to sync plugin
	go cron.ReInitLocalIp()    // retry to get ipaddr every day
	go cron.GetMetricCount()
	cron.ReportAgentStatus()
	cron.SyncMinePlugins()
	cron.SyncBuiltinMetrics()
	cron.SyncTrustableIps()
	cron.Collect()

	go http.Start()
	go http.StartLoopback()

	select {}

}
