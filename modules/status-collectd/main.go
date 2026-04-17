package main

import (
	"flag"
	"fmt"
	"github.com/signmem/viper"
	"github.com/signmem/falcon-plus/modules/status-collectd/g"
	"github.com/signmem/falcon-plus/modules/status-collectd/kfk"
	"github.com/signmem/falcon-plus/modules/status-collectd/proc"
	"github.com/signmem/falcon-plus/modules/status-collectd/selector"
	"log"
	_ "os"
	_ "os/signal"
	"runtime"
	_ "syscall"
	"sync"
	"time"
	myhttp "github.com/signmem/falcon-plus/modules/status-collectd/http"
)

var (
	Config *viper.Viper
	disableSend bool
	enableCnt bool
	RunStat bool
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// modify channel by terry.zeng
// func fetchCounter(ch chan *TsdbItem, wg *sync.WaitGroup, timeout time.Duration) {

func fetchCounter(ch chan *kfk.MItem, wg *sync.WaitGroup, timeout time.Duration) {
	for _, falconType := range [5]string{"transfer", "graph", "kafka_consumer", "trend", "judge"} {

		sKey := fmt.Sprintf("falconServers.%s", falconType)
		for _, server := range Config.GetStringSlice(sKey) {

			c, _ := kfk.ClientFactory(falconType, server, timeout)
			wg.Add(1)
			go func(c kfk.FalconCommon, falconType string, server string) {
				defer wg.Done()
				defer c.Close()

				if selector.Role == "slave" {
					log.Println("[DEBUG] fetchCounter() go routine break.")
					RunStat = false
					runtime.Goexit()
				}

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
					if enableCnt && mType == "qps" {
						ch <- counter.ToKafkaItem4Cnt(falconType, mType, server)
					} else {
						ch <- counter.ToKafkaItem(falconType, mType, server)
					}
				}
			}(c, falconType, server)

		}
	}
	wg.Wait()
}

func InitConfig() {
	v := viper.New()
	v.SetConfigType(g.Config().Env.Type)
	v.SetConfigName(g.Config().Env.Name)
	v.AddConfigPath(g.Config().Env.Path)

	err := v.ReadInConfig()
	if err != nil {
		log.Fatal("error on parsing configuration file", err)
	}
	Config = v
}

// func fetchHealth(ch chan *TsdbItem, wg *sync.WaitGroup, timeout time.Duration) {
// modify channel from terry.zeng
func fetchHealth(ch chan *kfk.MItem, wg *sync.WaitGroup, timeout time.Duration) {
	falconServers := Config.GetStringMapStringSlice("falconServers")

	for falconType, servers := range falconServers {
		for _, server := range servers {

			c, _ := kfk.ClientFactory(falconType, server, timeout)
			wg.Add(1)

			go func(c kfk.FalconCommon, falconType string, server string) {

				defer wg.Done()
				defer c.Close()

				if selector.Role == "slave" {
					log.Println("[DEBUG] fetchHealth() go routine break.")
					RunStat = false
					runtime.Goexit()
				}

				health, err := c.Health()
				ch <- kfk.HealthToKafkaItem(falconType, server, err == nil && health)
			}(c, falconType, server)

		}
	}
}

func fetchLocalHealth(ch chan *kfk.MItem, wg *sync.WaitGroup) {
	var localProc []*proc.SCount
	localProc = append(localProc, proc.SendToKafkaCntTotal)
	localProc = append(localProc, proc.SendToKafkaCntDrop)
	localProc = append(localProc, proc.SendToKafkaCntSuccess)

	if selector.Role == "slave" {
		log.Println("[DEBUG] fetchLocalHealth() go routine break.")
		RunStat = false
		runtime.Goexit()
	}

	for _, name := range localProc {
		wg.Add(1)
		go func(name *proc.SCount) {
			defer wg.Done()
			ch <- kfk.GetLocalInc(name)
		}(name)
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
	InitConfig()
	disableSend = g.Config().DisableSend
	enableCnt  = g.Config().EnableCnt
	kfk.Topic = g.Config().Kafka.Topic
	kfk.KafkaServer = g.Config().Kafka.Servers

}


func main() {

	httpListen := g.Config().HTTP.Listen + ":" + g.Config().HTTP.Port
	go myhttp.Start(httpListen)
	go selector.Start()

	// ch := make(chan *TsdbItem, 20)
	// sender := NewTsdbSender(Config.GetString("tsdbServer"), disableSend)
	// go sender.Start(ch, time.Second)

	for {

		if selector.Role == "master"  {

			ch := make(chan *kfk.MItem, 5)
			MItemList := kfk.MItemList{}
			go MItemList.Start(ch, 60)
			var wg sync.WaitGroup

			ticker := time.NewTicker(time.Minute * 1)

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

			ticker3 := time.NewTicker(time.Minute * 1)
			var wg3 sync.WaitGroup
			go func() {
				
				for ; true; <-ticker3.C {
					fetchLocalHealth(ch, &wg3)
				}
			}()

			/*
			sigCh := make(chan os.Signal)
			signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
			s := <-sigCh

			// wait running collect goroutine
			log.Printf("Get signal: %s, Waiting collect...\n", s)
			*/

			// wait and stop collect counter
			wg.Wait()
			ticker.Stop()

			// wait and stop collect health
			wg2.Wait()
			ticker2.Stop()

			/*
			for i := 0; i < 10; i++ {
				if len(ch) == 0 && sender.IsClear() {
					break
				}
				log.Printf("Wait 200 ms for sender #%d", i)
				time.Sleep(time.Millisecond * 200)
			}
			*/
		}
		time.Sleep( 10 * time.Second )
	}

	select {}
}
