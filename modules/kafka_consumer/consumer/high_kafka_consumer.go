package consumer

import (
	"errors"

	"github.com/IBM/sarama"
	nsema "github.com/toolkits/concurrent/semaphore"

	"github.com/signmem/falcon-plus/modules/kafka_consumer/g"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/proc"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/sender"
)

var (
	highGroup *cluster.Consumer
)

func highInitGroup() error {
	g.Logger.Info("start to init consumer group of high version kafka")
	groupCfg := cluster.NewConfig()
	cfg := g.Config()
	if cfg.Consumer.Group == "" {
		return errors.New("kafka group name in config is empty")
	}
	if len(cfg.Consumer.Topics) < 0 {
		return errors.New("kafka topics in config is empty")
	}
	if len(cfg.Consumer.Kafka) == 0 {
		return errors.New("kafka address in config is empty")
	}

	if cfg.Consumer.Offset == "oldest" {
		groupCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	} else {
		groupCfg.Consumer.Offsets.Initial = DEFAULT_OFFSET
	}

	groupCfg.Consumer.Return.Errors = true

	var err error
	highGroup, err = cluster.NewConsumer(cfg.Consumer.Kafka, cfg.Consumer.Group, cfg.Consumer.Topics, groupCfg)
	if err != nil {
		highGroup = nil
		return err
	}
	return nil
}

func highRun() {
	if highGroup == nil {
		g.Logger.Error("run kafka consumer group, highGroup object is nil")
	}
	// init semaphore
	concurrent := g.Config().Consumer.Concurrent

	if concurrent < 1 {
		concurrent = 1
	}
	sema := nsema.NewSemaphore(concurrent)
	for message := range highGroup.Messages() {
		sema.Acquire()
		go func(msg *sarama.ConsumerMessage) {
			defer sema.Release()
			sender.Push2TrendSendQueue(string(msg.Value))
			proc.ConsumeCnt.Incr()
			g.Logger.Debugf("Get message: %s\n", string(msg.Value))
			highGroup.MarkOffset(msg, "")
		}(message)
	}
}

func highErrorPrinter() {
	if highGroup == nil {
		g.Logger.Error("run kafka consumer group, highGroup object is nil")
	}
	for err := range highGroup.Errors() {
		g.Logger.Errorf("high consumer group error: %s", err.Error())
	}
}

func highStop() {
	if highGroup != nil {
		if err := highGroup.Close(); err != nil {
			g.Logger.Errorf("Error closing high consumer group :%s", err.Error())
		}
	}
}
