//added by vincent.zhang for sending to kafka 2017.09.25
package sender

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/signmem/falcon-plus/modules/transfer/g"
	"github.com/signmem/falcon-plus/modules/transfer/proc"

	"github.com/IBM/sarama"
)

var (
	errMaxProducers  = fmt.Errorf("maximum producers number reached")
	kafkaProducerCfg *sarama.Config //sarama kafka config
)

type KafkaProducer struct {
	p    sarama.AsyncProducer
	name string
}

func (this *KafkaProducer) Name() string {
	return this.name
}

func (this *KafkaProducer) Run() {
	cfg := g.Config()
	if this.p == nil {
		return
	}
	for {
		select {
		case err := <-this.p.Errors():
			log.Println("Failed to produce message to kafka", err)
			proc.SendToKafkaFailCntTotal.Incr()
			// Added by qimin.xu for kafka filter producer
			kafkaProductName := this.Name()
			// vipfalcon  主 topic 队列计数器
			if kafkaProductName == cfg.Kafka.Topic {
				proc.SendToKafkaFailCnt.Incr()
			} else if kafkaProductName == cfg.Kafka.LogTopic {
				proc.SendToKafkaLogFailCnt.Incr()
			} else {
				// filter kafka 队列计数器
				if len(cfg.Kafka.Filter) > 0 {
					for f := range cfg.Kafka.Filter {
						if kafkaProductName == cfg.Kafka.Filter[f]["topic"] {
							proc.SendToKafkaFailCntMap[f].Incr()
						}
					}
				}
			}
		case <-this.p.Successes():
			proc.SendToKafkaCntTotal.Incr()
			// Added by qimin.xu for kafka filter producer
			kafkaProductName := this.Name()
			// vipfalcon  主 topic 队列计数器
			if kafkaProductName == cfg.Kafka.Topic {
				proc.SendToKafkaCnt.Incr()
			} else if kafkaProductName == cfg.Kafka.LogTopic {
				proc.SendToKafkaLogCnt.Incr()
			} else {
				// filter kafka 队列计数器
				if len(cfg.Kafka.Filter) > 0 {
					for f := range cfg.Kafka.Filter {
						if kafkaProductName == cfg.Kafka.Filter[f]["topic"] {
							proc.SendToKafkaCntMap[f].Incr()
						}
					}
				}
			}
		}
	}
}

//added by vincent.zhang for sending string log to kafka
func (this *KafkaProducer) LogRun() {
	cfg := g.Config()
	if this.p == nil {
		return
	}
	for {
		select {
		case err := <-this.p.Errors():
			log.Println("Failed to produce message to kafka", err)
			proc.SendToKafkaFailCntTotal.Incr()
			// Added by qimin.xu for kafka filter producer
			kafkaProductName := this.Name()
			// vipfalcon  主 topic 队列计数器
			if kafkaProductName == cfg.Kafka.Topic {
				proc.SendToKafkaFailCnt.Incr()
			}
			// filter kafka 队列计数器
			if len(cfg.Kafka.Filter) > 0 {
				for f := range cfg.Kafka.Filter {
					if kafkaProductName == cfg.Kafka.Filter[f]["topic"] {
						proc.SendToKafkaFailCntMap[f].Incr()
					}
				}
			}
		case <-this.p.Successes():
			proc.SendToKafkaCntTotal.Incr()
			// Added by qimin.xu for kafka filter producer
			kafkaProductName := this.Name()
			// vipfalcon  主 topic 队列计数器
			if kafkaProductName == cfg.Kafka.Topic {
				proc.SendToKafkaCnt.Incr()
			}
			// filter kafka 队列计数器
			if len(cfg.Kafka.Filter) > 0 {
				for f := range cfg.Kafka.Filter {
					if kafkaProductName == cfg.Kafka.Filter[f]["topic"] {
						proc.SendToKafkaCntMap[f].Incr()
					}
				}
			}
		}
	}
}

func (this *KafkaProducer) AsyncClose() {
	if this.p != nil {
		this.p.AsyncClose()
	}
}

func (this *KafkaProducer) Close() error {
	if this.p != nil {
		return this.p.Close()
	}
	return nil
}

func NewKafkaProducer(name string, address []string) (*KafkaProducer, error) {
	if len(address) <= 0 {
		return nil, fmt.Errorf("no producer address when new producer")
	}
	p, err := sarama.NewAsyncProducer(address, kafkaProducerCfg)
	if err != nil {
		return nil, err
	} else {
		return &KafkaProducer{p, name}, nil
	}
}

//Kafka_Producer_Pool
type KafkaProducerPool struct {
	sync.RWMutex

	Name         string
	Address      []string
	MaxProducers int32
	MaxIdle      int32
	Cnt          int64

	New func(name string, address []string) (*KafkaProducer, error)

	active int32
	free   []*KafkaProducer
	all    map[string]*KafkaProducer
}

func initKafkaConfig() {
	if kafkaProducerCfg == nil {
		kafkaProducerCfg = sarama.NewConfig()
	}
	cfg := g.Config()
	kafkaProducerCfg.Producer.Return.Successes = true
	kafkaProducerCfg.Producer.Retry.Max = cfg.Kafka.MaxRetry
	//kafkaProducerCfg.Producer.Timeout = 1 * time.Second
	kafkaProducerCfg.Net.DialTimeout = time.Duration(cfg.Kafka.ConnTimeout) * time.Millisecond
	kafkaProducerCfg.Net.WriteTimeout = time.Duration(cfg.Kafka.CallTimeout) * time.Millisecond
}

func NewKafkaProducerPool(name string, address []string, maxProducers int32, maxIdle int32) *KafkaProducerPool {
	pool := KafkaProducerPool{Name: name, Address: address, MaxProducers: maxProducers, MaxIdle: maxIdle, Cnt: 0, all: make(map[string]*KafkaProducer)}
	pool.New = func(name string, address []string) (*KafkaProducer, error) {
		return NewKafkaProducer(name, address)
	}
	return &pool
}

func (this *KafkaProducerPool) Proc() string {
	this.RLock()
	defer this.RUnlock()

	return fmt.Sprintf("Name:%s,Cnt:%d,active:%d,all:%d,free:%d",
		this.Name, this.Cnt, this.active, len(this.all), len(this.free))
}

func (this *KafkaProducerPool) Fetch() (*KafkaProducer, error) {
	this.Lock()
	defer this.Unlock()

	// get from free
	producer := this.fetchFree()
	if producer != nil {
		return producer, nil
	}

	if this.overMax() {
		return nil, errMaxProducers
	}

	// create new producer
	producer, err := this.newProducer()
	if err != nil {
		return nil, err
	}
	err = this.runCheckResult(producer)
	if err != nil {
		return nil, err
	}
	this.increActive()
	return producer, nil
}

func (this *KafkaProducerPool) Release(producer *KafkaProducer) {
	this.Lock()
	defer this.Unlock()

	if this.overMaxIdle() {
		this.deleteProducer(producer)
		this.decreActive()
	} else {
		this.addFree(producer)
	}
}

func (this *KafkaProducerPool) ForceClose(producer *KafkaProducer) {
	this.Lock()
	defer this.Unlock()

	this.deleteProducer(producer)
	this.decreActive()
}

func (this *KafkaProducerPool) Destroy() {
	this.Lock()
	defer this.Unlock()

	for _, producer := range this.free {
		if producer != nil {
			producer.AsyncClose()
		}
	}

	for _, producer := range this.all {
		if producer != nil {
			producer.AsyncClose()
		}
	}

	this.active = 0
	this.free = []*KafkaProducer{}
	this.all = map[string]*KafkaProducer{}
}

// internal, concurrently unsafe
func (this *KafkaProducerPool) newProducer() (*KafkaProducer, error) {
	name := fmt.Sprintf("%s_%d_%d", this.Name, this.Cnt, time.Now().Unix())
	producer, err := this.New(name, this.Address)
	if err != nil {
		if producer != nil {
			producer.Close()
		}
		return nil, err
	}

	this.Cnt++
	this.all[producer.Name()] = producer
	return producer, nil
}

func (this *KafkaProducerPool) runCheckResult(producer *KafkaProducer) error {
	if producer == nil || producer.p == nil {
		return fmt.Errorf("run check result fail")
	}
	go producer.Run()
	return nil
}

func (this *KafkaProducerPool) deleteProducer(producer *KafkaProducer) {
	if producer != nil {
		producer.AsyncClose()
	}
	delete(this.all, producer.Name())
}

func (this *KafkaProducerPool) addFree(producer *KafkaProducer) {
	this.free = append(this.free, producer)
}

func (this *KafkaProducerPool) fetchFree() *KafkaProducer {
	if len(this.free) == 0 {
		return nil
	}

	producer := this.free[0]
	this.free = this.free[1:]
	return producer
}

func (this *KafkaProducerPool) increActive() {
	this.active += 1
}

func (this *KafkaProducerPool) decreActive() {
	this.active -= 1
}

func (this *KafkaProducerPool) overMax() bool {
	return this.active >= this.MaxProducers
}

func (this *KafkaProducerPool) overMaxIdle() bool {
	return int32(len(this.free)) >= this.MaxIdle
}
