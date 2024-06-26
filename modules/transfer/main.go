package main

import (
	"flag"
	"fmt"
	"github.com/signmem/falcon-plus/modules/transfer/g"
	"github.com/signmem/falcon-plus/modules/transfer/http"
	"github.com/signmem/falcon-plus/modules/transfer/proc"
	"github.com/signmem/falcon-plus/modules/transfer/receiver"
	"github.com/signmem/falcon-plus/common/redisdb"
	"github.com/signmem/falcon-plus/modules/transfer/sender"
	"os"
)

func init () {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	versionGit := flag.Bool("vg", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}
	if *versionGit {
		fmt.Println(g.VERSION, g.COMMIT)
		os.Exit(0)
	}

	// global config
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

	proc.Start()
	proc.InitKafkaCntSet()  // add by qimin.xu 初始化kafka发送计数器

	sender.Start()
	receiver.Start()

	// http
	http.Start()

	select {}
}
