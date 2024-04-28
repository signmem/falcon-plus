//适合kafka版本小于0.9.0
package consumer

import (
	"errors"
	"time"
	"github.com/Shopify/sarama"
	nsema "github.com/toolkits/concurrent/semaphore"
	"github.com/wvanbergen/kafka/consumergroup"
	"github.com/wvanbergen/kazoo-go"
	"github.com/open-falcon/falcon-plus/modules/kafka_consumer/g"
	"github.com/open-falcon/falcon-plus/modules/kafka_consumer/proc"
	"github.com/open-falcon/falcon-plus/modules/kafka_consumer/sender"
)

var (
	lowGroup       *consumergroup.ConsumerGroup
	zookeeperNodes []string
)

func lowInitGroup() error {
	g.Logger.Info("start to init consumer group of low version kafka")
	groupCfg := consumergroup.NewConfig()
	cfg := g.Config()
	if cfg.Consumer.Group == "" {
		return errors.New("kafka group name in config is empty")
	}
	if cfg.Consumer.Zookeeper == "" {
		return errors.New("kafka zookeeper in config is empty")
	}
	if len(cfg.Consumer.Topics) < 0 {
		return errors.New("kafka topics in config is empty")
	}

	if cfg.Consumer.Offset == "oldest" {
		groupCfg.Offsets.Initial = sarama.OffsetOldest
	} else {
		groupCfg.Offsets.Initial = DEFAULT_OFFSET
	}

	if cfg.Consumer.OffsetTimeout <= 0 {
		groupCfg.Offsets.ProcessingTimeout = DEFAULT_TIMEOUT
	} else {
		groupCfg.Offsets.ProcessingTimeout = time.Duration(cfg.Consumer.OffsetTimeout) * time.Second
	}

	groupCfg.Consumer.Return.Errors = true

	zookeeperNodes, groupCfg.Zookeeper.Chroot = kazoo.ParseConnectionString(cfg.Consumer.Zookeeper)
	var err error
	lowGroup, err = consumergroup.JoinConsumerGroup(cfg.Consumer.Group, cfg.Consumer.Topics, zookeeperNodes, groupCfg)
	if err != nil {
		lowGroup = nil
		return err
	}
	return nil
}

func lowRun() {
	if lowGroup == nil {
		g.Logger.Error("run kafka consumer group, lowGroup object is nil")
	}
	// init semaphore
	concurrent := g.Config().Consumer.Concurrent

	if concurrent < 1 {
		concurrent = 1
	}
	sema := nsema.NewSemaphore(concurrent)
	for message := range lowGroup.Messages() {
		sema.Acquire()
		go func(msg *sarama.ConsumerMessage) {
			defer sema.Release()
			sender.Push2TrendSendQueue(string(msg.Value))
			proc.ConsumeCnt.Incr()
			g.Logger.Debugf("Get message: %s\n", string(msg.Value))
			lowGroup.CommitUpto(msg)
		}(message)
	}
}

func lowErrorPrinter() {
	if lowGroup == nil {
		g.Logger.Error("run kafka consumer group, lowGroup object is nil")
	}
	for err := range lowGroup.Errors() {
		g.Logger.Errorf("low consumer group error: %s", err.Error())
	}
}

func lowStop() {
	if lowGroup != nil {
		if err := lowGroup.Close(); err != nil {
			g.Logger.Errorf("Error closing low consumer group :%s", err.Error())
		}
	}
}
