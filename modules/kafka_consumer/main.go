package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/signmem/falcon-plus/modules/kafka_consumer/consumer"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/g"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/http"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/proc"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/sender"
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

			consumer.Stop()
			log.Println("kafka consumer group stop ok")

			log.Println(pid, "exit")
			os.Exit(0)
		}
	}
}

func main() {
	cfg := flag.String("c", "cfg.json", "specify config file")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	// global config
	g.ParseConfig(*cfg)

	g.Logger = g.InitLog()

	proc.Start()

	sender.Start()

	//http
	http.Start()

	//start kafka consumer group
	consumer.Start()

	start_signal(os.Getpid(), g.Config())
}
