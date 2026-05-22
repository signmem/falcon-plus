package kfk

import (
	"github.com/signmem/sarama"
	"github.com/signmem/falcon-plus/modules/status-collectd/g"
	"github.com/signmem/falcon-plus/modules/status-collectd/proc"
	"time"
	"sync"
)

var (
	Topic		string
	KafkaServer	[]string
	kafkaProducer	sarama.SyncProducer
	producerOnce	sync.Once
	producerErr	error
	// 异步生产者全局单例
	asyncProducer	sarama.AsyncProducer
	asyncOnce	sync.Once
	asyncErr	error
)


// 初始化 Kafka 生产者 (全局单例)
func InitKafkaProducer() {
	producerOnce.Do(func() {
		config := sarama.NewConfig()
		config.Producer.RequiredAcks = sarama.WaitForAll
		config.Producer.Partitioner = sarama.NewRandomPartitioner
		config.Producer.Return.Successes = true

		kafkaProducer, producerErr = sarama.NewSyncProducer(KafkaServer, config)
		if producerErr != nil {
			g.Logger.Fatalf("init kafka producer failed: %v", producerErr)
		}
		g.Logger.Info("init kafka producer success")
	})
}


// 关闭生产者 (程序退出时调用)
func CloseKafkaProducer() {
	if kafkaProducer != nil {
		_ = kafkaProducer.Close()
	}
}

// ====================== 异步生产者 =====================
func InitAsyncProducer() {
	asyncOnce.Do(func() {
		config := sarama.NewConfig()
		config.Producer.RequiredAcks = sarama.WaitForAll
		config.Producer.Partitioner = sarama.NewRandomPartitioner
		config.Producer.Return.Successes = true
		config.Producer.Timeout = 5 * time.Second

		// 生产优化参数
		config.Producer.Flush.Frequency = 100 * time.Millisecond
		config.Producer.Flush.MaxMessages = 200

		p, err := sarama.NewAsyncProducer(KafkaServer, config)
		if err != nil {
			asyncErr = err
			g.Logger.Fatalf("init async producer failed: %v", err)
			return
		}
		asyncProducer = p

		// 只启动一个 goroutine 处理结果
		go handleAsyncResult(p)
		g.Logger.Info("init async kafka producer success")
	})
}

// 处理成功/失败 (全局只运行一个)
func handleAsyncResult(p sarama.AsyncProducer) {
	for {
		select {
		case err := <-p.Errors():
			if err != nil {
				g.Logger.Errorf("async send error: %v", err)
				proc.SendToKafkaCntDrop.Incr()
			}
		case <-p.Successes():
			proc.SendToKafkaCntSuccess.Incr()
		}
	}
}

func CloseAsyncProducer() {
	if asyncProducer != nil {
		asyncProducer.AsyncClose()
	}
}


// kafka produce
func Produce(m []*MItem) {

	// initial kafka
	if kafkaProducer == nil {
		InitKafkaProducer()
	}

	for _, item := range m {

		proc.SendToKafkaCntTotal.Incr()

		// 构造消息
		msg := &sarama.ProducerMessage{
			Topic: Topic,
			Value: sarama.StringEncoder(item.String()),
		}

		// 发送（复用全局生产者）
		_, _, err := kafkaProducer.SendMessage(msg)
		if err != nil {
			g.Logger.Errorf("send kafka msg failed: %v", err)
			proc.SendToKafkaCntDrop.Incr()
			continue
		}

		proc.SendToKafkaCntSuccess.Incr()

	}
}


// 异步发送 (安全、无泄漏、高速)
func AsyncProducer(m []*MItem) {
	if asyncProducer == nil {
		InitAsyncProducer()
	}

	if asyncErr != nil || asyncProducer == nil {
		g.Logger.Error("async producer not ready")
		return
	}

	for _, item := range m {
		// 正确统计总数
		proc.SendToKafkaCntTotal.Incr()

		if g.Config().Debug {
			g.Logger.Debugf("%s", item.String())
		}

		msg := &sarama.ProducerMessage{
			Topic: Topic,
			Value: sarama.ByteEncoder(item.String()),
		}

		// 发送带防阻塞保护
		select {
		case asyncProducer.Input() <- msg:
		case <-time.After(1 * time.Second):
			g.Logger.Error("async producer channel full")
			proc.SendToKafkaCntDrop.Incr()
		}
	}
}
