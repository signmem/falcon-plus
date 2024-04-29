package main

import (
	"flag"
	"fmt"
	"github.com/signmem/falcon-plus/modules/hbs/cache"
	"github.com/signmem/falcon-plus/modules/hbs/db"
	"github.com/signmem/falcon-plus/modules/hbs/g"
	"github.com/signmem/falcon-plus/modules/hbs/http"
	"github.com/signmem/falcon-plus/modules/hbs/rpc"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	g.ParseConfig(*cfg)

	db.Init()
	cache.Init()

	go cache.DeleteStaleAgents()

	go http.Start()
	go rpc.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println()
		db.DB.Close()
		os.Exit(0)
	}()

	select {}
}
