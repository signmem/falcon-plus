package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	// _ "net/http/pprof"

	. "github.com/open-falcon/falcon-plus/common/model"
	"gitlab.tools.vipshop.com/vip-ops-sh/falcon-stats-collector/http"
	"gitlab.tools.vipshop.com/vip-ops-sh/falcon-stats-collector/g"
)

var (
	env         = flag.String("e", "test", "environment, test|prd")
	disableSend = flag.Bool("disableSend", false, "Disable send, for development")
	enableCnt   = flag.Bool("enableCnt", true, "除了QPS，将 Counter 值也作纪录，用于排查数据精度问题")
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func fetchCounter(ch chan *TsdbItem, wg *sync.WaitGroup, timeout time.Duration) {
	for _, falconType := range [5]string{"transfer", "graph", "kafka_consumer", "trend", "judge"} {

		sKey := fmt.Sprintf("falconServers.%s", falconType)
		for _, server := range Config.GetStringSlice(sKey) {

			c, _ := ClientFactory(falconType, server, timeout)
			wg.Add(1)
			go func(c FalconCommon, falconType string, server string) {
				defer wg.Done()
				defer c.Close()

				counters, err := c.Counter()
				if err != nil {
					log.Println(err)
					return
				}

				for _, counter := range counters {
					mKey := fmt.Sprintf("falconCounterMetrics.%s.%s", falconType, counter.Name)
					mType := Config.GetString(mKey)
					if mType == "" {
						continue
					}

					// 为 qps 指标发一份 Counter 数据
					if *enableCnt && mType == "qps" {
						ch <- counter.toTsdbItem4Cnt(falconType, mType, server)
					} else {
						ch <- counter.toTsdbItem(falconType, mType, server)
					}
				}
			}(c, falconType, server)

		}
	}
	wg.Wait()
}

func fetchHealth(ch chan *TsdbItem, wg *sync.WaitGroup, timeout time.Duration) {
	falconServers := Config.GetStringMapStringSlice("falconServers")

	if http.Debug {
		log.Printf("[DEBUG] falconServers: %s", falconServers)
	}

	for falconType, servers := range falconServers {
		for _, server := range servers {

			if http.Debug {
				log.Printf("[DEBUG] server: %s", server)
			}

			c, _ := ClientFactory(falconType, server, timeout)
			wg.Add(1)

			go func(c FalconCommon, falconType string, server string) {

				defer wg.Done()
				defer c.Close()

				if http.Debug {
					log.Printf("[DEBUG] type: %s", falconType )
				}

				health, err := c.Health()
				ch <- HealthToTsdbItem(falconType, server, err == nil && health)
			}(c, falconType, server)

		}
	}
}

func init () {
	cfg := flag.String("c", "cfg.json", "specify config file")
	flag.Parse()

	// global config
	g.ParseConfig(*cfg)
	g.Logger = g.InitLog()

	flag.Parse()
	// init config
	InitConfig(*env)

}


func main() {


	ch := make(chan *TsdbItem, 20)

	sender := NewTsdbSender(Config.GetString("tsdbServer"), *disableSend)

	httpListen := Config.GetString("httpLister")
	go http.Start(httpListen)
	http.Debug = Config.GetBool("debug")

	go sender.Start(ch, time.Second)

	ticker := time.NewTicker(time.Minute * 1)
	var wg sync.WaitGroup

	go func() {
		for ; true; <-ticker.C {
			fetchCounter(ch, &wg, time.Second*5)
		}
	}()

	ticker2 := time.NewTicker(time.Minute * 1)
	var wg2 sync.WaitGroup

	go func() {
		for ; true; <-ticker2.C {
			fetchHealth(ch, &wg2, time.Second*5)
		}
	}()

	// go func() {
	//     log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	s := <-sigCh

	// wait running collect goroutine
	log.Printf("Get signal: %s, Waiting collect...\n", s)

	// wait and stop collect counter
	wg.Wait()
	ticker.Stop()

	// wait and stop collect health
	wg2.Wait()
	ticker2.Stop()


	for i := 0; i < 10; i++ {
		if len(ch) == 0 && sender.IsClear() {
			break
		}
		log.Printf("Wait 200 ms for sender #%d", i)
		time.Sleep(time.Millisecond * 200)
	}
	select {}
}
