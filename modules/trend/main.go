package main

import (
	"flag"
	"fmt"
	"github.com/signmem/falcon-plus/modules/trend/register"
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/signmem/falcon-plus/common/redisdb"
	"github.com/signmem/falcon-plus/modules/trend/cache"
	"github.com/signmem/falcon-plus/modules/trend/g"
	"github.com/signmem/falcon-plus/modules/trend/http"
	"github.com/signmem/falcon-plus/modules/trend/rpc"
	"github.com/signmem/falcon-plus/modules/trend/writer"
)

func start_signal(pid int, cfg *g.GlobalConfig) {
	sigs := make(chan os.Signal, 1)
	log.Println(pid, "register signal notify")
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		s := <-sigs
		log.Println("recv", s)

		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			log.Println("graceful shut down")
			if cfg.Http.Enabled {
				//ToDo close http
			}
			log.Println("http stop ok")
			if cfg.Rpc.Enabled {
				rpc.Close_chan <- 1
				<-rpc.Close_done_chan
			}
			log.Println("rpc stop ok")
			writer.Close()
			log.Println("writer stop ok")
			log.Println(pid, "exit")
			os.Exit(0)
		}
	}
}


func init () {
	cfg := flag.String("c", "cfg.json", "specify config file")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
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


	g.Logger = g.InitLog()

	//http
	go http.Start()

	go rpc.Start()

	go writer.Start()

	go cache.Start()
	go register.RegCron()
	start_signal(os.Getpid(), g.Config())
	select {}
}
