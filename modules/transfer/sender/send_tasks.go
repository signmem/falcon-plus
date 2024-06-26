package sender

import (
	"bytes"

	"github.com/IBM/sarama" // added by vincent.zhang for sending to kafka
	cmodel "github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/transfer/g"
	"github.com/signmem/falcon-plus/modules/transfer/proc"
	nsema "github.com/toolkits/concurrent/semaphore"
	"github.com/toolkits/container/list"
	"log"
	"time"
)

// send
const (
	DefaultSendTaskSleepInterval = time.Millisecond * 50 //默认睡眠间隔为50ms
)

// TODO 添加对发送任务的控制,比如stop等
func startSendTasks() {
	cfg := g.Config()
	// init semaphore
	judgeConcurrent := cfg.Judge.MaxConns
	graphConcurrent := cfg.Graph.MaxConns
	tsdbConcurrent := cfg.Tsdb.MaxConns
	kafkaConurrent := cfg.Kafka.MaxConcurrent //added by vincent.zhang for sending to kafka

	if tsdbConcurrent < 1 {
		tsdbConcurrent = 1
	}

	if judgeConcurrent < 1 {
		judgeConcurrent = 1
	}

	if graphConcurrent < 1 {
		graphConcurrent = 1
	}

	//added by vincent.zhang for sending to kafka
	if kafkaConurrent < 1 {
		kafkaConurrent = 1
	}

	// init send go-routines
	for node := range cfg.Judge.Cluster {
		queue := JudgeQueues[node]
		go forward2JudgeTask(queue, node, judgeConcurrent)
	}

	for node, nitem := range cfg.Graph.ClusterList {
		for _, addr := range nitem.Addrs {
			queue := GraphQueues[node+addr]
			go forward2GraphTask(queue, node, addr, graphConcurrent)
		}
	}

	if cfg.Tsdb.Enabled {
		go forward2TsdbTask(tsdbConcurrent)
	}

	//added by vincent.zhang for sending to kafka
	if cfg.Kafka.Enabled {
		go forward2KafkaTask(kafkaConurrent)
		// added by qimin.xu for sending filter queue data to kafka
		if len(cfg.Kafka.Filter) > 0 {
			for f := range cfg.Kafka.Filter {
				go forward2KafkaFilterTask(kafkaConurrent, f)
			}
			log.Println("load kafka filter form conf, do filter task. tasks num is: ", len(cfg.Kafka.Filter))
		}
	}

	if cfg.Kafka.LogEnabled {
		go forward2KafkaLogTask(kafkaConurrent)
	}
}

// Judge定时任务, 将 Judge发送缓存中的数据 通过rpc连接池 发送到Judge
func forward2JudgeTask(Q *list.SafeListLimited, node string, concurrent int) {
	batch := g.Config().Judge.Batch // 一次发送,最多batch条数据
	addr := g.Config().Judge.Cluster[node]
	sema := nsema.NewSemaphore(concurrent)

	for {
		items := Q.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}

		judgeItems := make([]*cmodel.JudgeItem, count)
		for i := 0; i < count; i++ {
			judgeItems[i] = items[i].(*cmodel.JudgeItem)
		}

		//	同步Call + 有限并发 进行发送
		sema.Acquire()
		go func(addr string, judgeItems []*cmodel.JudgeItem, count int) {
			defer sema.Release()

			resp := &cmodel.SimpleRpcResponse{}
			var err error
			sendOk := false
			for i := 0; i < 3; i++ { //最多重试3次
				err = JudgeConnPools.Call(addr, "Judge.Send", judgeItems, resp)
				if err == nil {
					sendOk = true
					break
				}
				time.Sleep(time.Millisecond * 10)
			}

			// statistics
			if !sendOk {
				log.Printf("send judge %s:%s fail: %v", node, addr, err)
				proc.SendToJudgeFailCnt.IncrBy(int64(count))
			} else {
				proc.SendToJudgeCnt.IncrBy(int64(count))
			}
		}(addr, judgeItems, count)
	}
}

// Graph定时任务, 将 Graph发送缓存中的数据 通过rpc连接池 发送到Graph
func forward2GraphTask(Q *list.SafeListLimited, node string, addr string, concurrent int) {
	batch := g.Config().Graph.Batch // 一次发送,最多batch条数据
	sema := nsema.NewSemaphore(concurrent)

	for {
		items := Q.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}

		graphItems := make([]*cmodel.GraphItem, count)
		for i := 0; i < count; i++ {
			graphItems[i] = items[i].(*cmodel.GraphItem)
		}

		sema.Acquire()
		go func(addr string, graphItems []*cmodel.GraphItem, count int) {
			defer sema.Release()

			resp := &cmodel.SimpleRpcResponse{}
			var err error
			sendOk := false
			for i := 0; i < 3; i++ { //最多重试3次
				err = GraphConnPools.Call(addr, "Graph.Send", graphItems, resp)
				if err == nil {
					sendOk = true
					break
				}
				time.Sleep(time.Millisecond * 10)
			}

			// statistics
			if !sendOk {
				log.Printf("send to graph %s:%s fail: %v", node, addr, err)
				proc.SendToGraphFailCnt.IncrBy(int64(count))
			} else {
				proc.SendToGraphCnt.IncrBy(int64(count))
			}
		}(addr, graphItems, count)
	}
}

// Tsdb定时任务, 将数据通过api发送到tsdb
func forward2TsdbTask(concurrent int) {
	batch := g.Config().Tsdb.Batch // 一次发送,最多batch条数据
	retry := g.Config().Tsdb.MaxRetry
	sema := nsema.NewSemaphore(concurrent)

	for {
		items := TsdbQueue.PopBackBy(batch)
		if len(items) == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}
		//  同步Call + 有限并发 进行发送
		sema.Acquire()
		go func(itemList []interface{}) {
			defer sema.Release()

			var tsdbBuffer bytes.Buffer
			for i := 0; i < len(itemList); i++ {
				tsdbItem := itemList[i].(*cmodel.TsdbItem)
				tsdbBuffer.WriteString(tsdbItem.TsdbString())
				tsdbBuffer.WriteString("\n")
			}

			var err error
			for i := 0; i < retry; i++ {
				err = TsdbConnPoolHelper.Send(tsdbBuffer.Bytes())
				if err == nil {
					proc.SendToTsdbCnt.IncrBy(int64(len(itemList)))
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			if err != nil {
				proc.SendToTsdbFailCnt.IncrBy(int64(len(itemList)))
				log.Println(err)
				return
			}
		}(items)
	}
}

// Kafka定时任务, 将数据通过kafka productor发送到kafka
// added by vincent.zhang for sending to kafka
func forward2KafkaTask(concurrent int) {
	//single kafka producer method is necessary
	if kafkaProducer != nil {
		go kafkaProducer.Run()
	}
	//end

	batch := g.Config().Kafka.Batch // 一次发送,最多batch条数据
	sema := nsema.NewSemaphore(concurrent)
	for {
		items := KafkaQueue.PopBackBy(batch)
		if len(items) == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}
		//  同步Call + 有限并发 进行发送
		sema.Acquire()
		go func(itemList []interface{}) {
			defer sema.Release()
			// single kafka producer method
			if kafkaProducer == nil || kafkaProducer.p == nil {
				proc.SendToKafkaFailCnt.IncrBy(int64(len(itemList)))
				proc.SendToKafkaFailCntTotal.IncrBy(int64(len(itemList)))
				return
			}
			for i := 0; i < len(itemList); i++ {
				kafkaItem := itemList[i].(*KafkaItem)
				kafkaProducer.p.Input() <- &sarama.ProducerMessage{Topic: g.Config().Kafka.Topic, Key: nil, Value: sarama.ByteEncoder(kafkaItem.KafkaString())}
			}
			// single kafka producer end

			// producer pool method
			/*
				producer, err := KafkaPool.Fetch()
				if err != nil {
					//fmt.Printf("get producer fail: err %v. proc: %s\n", err, KafkaPool.Proc())
					proc.SendToKafkaFailCnt.IncrBy(int64(len(itemList)))
					return
				}
				for i := 0; i < len(itemList); i++ {
					kafkaItem := itemList[i].(*KafkaItem)
					producer.p.Input() <- &sarama.ProducerMessage{Topic: g.Config().Kafka.Topic, Key: nil, Value: sarama.ByteEncoder(kafkaItem.KafkaString())}
				}
				KafkaPool.Release(producer)
			*/
			// producer pool end
		}(items)
	}
}

// added by qimin.xu for kafkaFilter producer
func forward2KafkaFilterTask(concurrent int, f string) {
	if kafkaFilterProducer[f] != nil {
		go kafkaFilterProducer[f].Run()
	}

	batch := g.Config().Kafka.Batch
	sema := nsema.NewSemaphore(concurrent)
	for {
		items := KafkaFilterQueues[f].PopBackBy(batch)
		if len(items) == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}
		sema.Acquire()
		go func(itemList []interface{}) {
			defer sema.Release()
			if kafkaFilterProducer[f] == nil || kafkaFilterProducer[f].p == nil {
				//proc.SendToKafkaFailCnt.IncrBy(int64(len(itemList)))
				proc.SendToKafkaFailCntTotal.IncrBy(int64(len(itemList)))
				proc.SendToKafkaFailCntMap[f].IncrBy(int64(len(itemList)))
				return
			}
			for i := 0; i < len(itemList); i++ {
				kafkaItem := itemList[i].(*KafkaItem)
				kafkaFilterProducer[f].p.Input() <- &sarama.ProducerMessage{Topic: g.Config().Kafka.Filter[f]["topic"], Key: nil, Value: sarama.ByteEncoder(kafkaItem.KafkaString())}
			}
		}(items)
	}
}

func forward2KafkaLogTask(concurrent int) {
	//single kafka producer method is necessary
	if kafkaLogProducer != nil {
		go kafkaLogProducer.Run()
	}
	//end

	batch := g.Config().Kafka.Batch // 一次发送,最多batch条数据
	sema := nsema.NewSemaphore(concurrent)
	for {
		items := KafkaLogQueue.PopBackBy(batch)
		if len(items) == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}
		//  同步Call + 有限并发 进行发送
		sema.Acquire()
		go func(itemList []interface{}) {
			defer sema.Release()
			// single kafka producer method
			if kafkaLogProducer == nil || kafkaLogProducer.p == nil {
				proc.SendToKafkaLogFailCnt.IncrBy(int64(len(itemList)))
				return
			}
			for i := 0; i < len(itemList); i++ {
				kafkaItem := itemList[i].(*LogMetricItem)
				kafkaLogProducer.p.Input() <- &sarama.ProducerMessage{Topic: g.Config().Kafka.LogTopic, Key: nil, Value: sarama.ByteEncoder(kafkaItem.KafkaString())}
			}
		}(items)
	}
}
