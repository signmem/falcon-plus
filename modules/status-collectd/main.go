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
	"os"
	"context"
	"os/signal"
	_ "runtime"
	"syscall"
	"sync"
	"time"
	myhttp "github.com/signmem/falcon-plus/modules/status-collectd/http"
)

var (
	Config        *viper.Viper
	disableSend   bool
	enableCnt     bool
	RunStat       bool
	roleLock      sync.RWMutex
	runStatLock   sync.Mutex
)

// 一次性初始化全局资源
var (
	ch          = make(chan *kfk.MItem, 200)
	ticker      = time.NewTicker(1 * time.Minute)
	ticker2     = time.NewTicker(1 * time.Minute)
	ticker3     = time.NewTicker(1 * time.Minute)
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// ------------- fetchCounter 已修复 -------------
func fetchCounter(ch chan *kfk.MItem, wg *sync.WaitGroup, timeout time.Duration) {
	falconTypes := []string{"transfer", "graph", "kafka_consumer", "trend", "judge"}

	for _, falconType := range falconTypes {
		sKey := fmt.Sprintf("falconServers.%s", falconType)
		for _, server := range Config.GetStringSlice(sKey) {
			c, err := kfk.ClientFactory(falconType, server, timeout)
			if err != nil {
				log.Printf("[ERROR] create client failed: %s %s %v", falconType, server, err)
				continue
			}

			wg.Add(1)
			go func(c kfk.FalconCommon, ft string, srv string) {
				defer func() {
					if c != nil {
						c.Close()
					}
					wg.Done()
				}()

				roleLock.RLock()
				isSlave := selector.Role == "slave"
				roleLock.RUnlock()

				if isSlave {
					return
				}

				counters, err := c.Counter()
				if err != nil {
					log.Printf("[ERROR] counter failed: %s %s %v", ft, srv, err)
					return
				}

				for _, counter := range counters {
					mKey := fmt.Sprintf("falconCounterMetrics.%s.%s", ft, counter.Name)
					mType := Config.GetString(mKey)
					if mType == "" {
						continue
					}

					var item *kfk.MItem
					if enableCnt && mType == "qps" {
						item = counter.ToKafkaItem4Cnt(ft, mType, srv)
					} else {
						item = counter.ToKafkaItem(ft, mType, srv)
					}

					select {
					case ch <- item:
					case <-time.After(3 * time.Second):
						log.Printf("[WARN] channel full, drop item: %s %s", ft, srv)
					}
				}
			}(c, falconType, server)
		}
	}
}

// ------------- fetchHealth 修复 -------------
func fetchHealth(ch chan *kfk.MItem, wg *sync.WaitGroup, timeout time.Duration) {
	falconServers := Config.GetStringMapStringSlice("falconServers")

	for ft, servers := range falconServers {
		for _, srv := range servers {
			c, err := kfk.ClientFactory(ft, srv, timeout)
			if err != nil {
				continue
			}

			wg.Add(1)
			go func(c kfk.FalconCommon, ft string, srv string) {
				defer func() {
					if c != nil {
						c.Close()
					}
					wg.Done()
				}()

				roleLock.RLock()
				isSlave := selector.Role == "slave"
				roleLock.RUnlock()
				if isSlave {
					return
				}

				health, err := c.Health()
				item := kfk.HealthToKafkaItem(ft, srv, err == nil && health)

				select {
				case ch <- item:
				case <-time.After(3 * time.Second):
				}
			}(c, ft, srv)
		}
	}
}

// ------------- fetchLocalHealth 修复 -------------
func fetchLocalHealth(ch chan *kfk.MItem, wg *sync.WaitGroup) {
	roleLock.RLock()
	isSlave := selector.Role == "slave"
	roleLock.RUnlock()
	if isSlave {
		return
	}

	localProc := []*proc.SCount{
		proc.SendToKafkaCntTotal,
		proc.SendToKafkaCntDrop,
		proc.SendToKafkaCntSuccess,
	}

	for _, name := range localProc {
		wg.Add(1)
		go func(name *proc.SCount) {
			defer wg.Done()
			item := kfk.GetLocalInc(name)
			select {
			case ch <- item:
			case <-time.After(3 * time.Second):
			}
		}(name)
	}
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

func init() {
	cfg := flag.String("c", "cfg.json", "specify config file")
	flag.Parse()

	g.ParseConfig(*cfg)
	g.Logger = g.InitLog()
	InitConfig()

	disableSend = g.Config().DisableSend
	enableCnt = g.Config().EnableCnt
	kfk.Topic = g.Config().Kafka.Topic
	kfk.KafkaServer = g.Config().Kafka.Servers

	// 初始化 Kafka
	kfk.InitKafkaProducer()
	kfk.InitAsyncProducer()
}

// ------------- 消费 ch 并发送到 Kafka -------------
func consumeAndSend(ctx context.Context) {
	var itemList []*kfk.MItem
	t := time.NewTicker(3 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case item := <-ch:
			if item != nil {
				itemList = append(itemList, item)
			}
			// 批量发送
			if len(itemList) >= 60 {
				kfk.Produce(itemList)
				itemList = nil
			}
		case <-t.C:
			if len(itemList) > 0 {
				kfk.Produce(itemList)
				itemList = nil
			}
		}
	}
}

func main() {
	httpListen := g.Config().HTTP.Listen + ":" + g.Config().HTTP.Port
	go myhttp.Start(httpListen)
	go selector.Start()

	// 启动消费发送
	ctx, cancel := context.WithCancel(context.Background())
	go consumeAndSend(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	// 只启动一次定时任务
	go taskLoop(ctx, fetchCounter, ch, 5*time.Second)
	go taskLoop(ctx, fetchHealth, ch, 5*time.Second)
	go taskLoopNoTimeout(ctx, fetchLocalHealth, ch)

	// 主循环
	for {
		select {
		case <-sigCh:
			log.Println("exit signal received")
			cancel()
			goto EXIT
		default:
		}

		roleLock.RLock()
		isMaster := selector.Role == "master"
		roleLock.RUnlock()

		if isMaster {
			time.Sleep(2 * time.Second)
		} else {
			time.Sleep(10 * time.Second)
		}
	}

EXIT:
	log.Println("waiting exit...")
	time.Sleep(2 * time.Second)
	log.Println("exited")
}

func taskLoop(ctx context.Context, f func(chan *kfk.MItem, *sync.WaitGroup, time.Duration), ch chan *kfk.MItem, timeout time.Duration) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			f(ch, &wg, timeout)
		}
	}
}

func taskLoopNoTimeout(ctx context.Context, f func(chan *kfk.MItem, *sync.WaitGroup), ch chan *kfk.MItem) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			f(ch, &wg)
		}
	}
}
