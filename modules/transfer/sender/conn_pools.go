package sender

import (
	"log"

	backend "github.com/signmem/falcon-plus/common/backend_pool"
	"github.com/signmem/falcon-plus/modules/transfer/g"
	nset "github.com/toolkits/container/set"
)

func initConnPools() {
	cfg := g.Config()

	// judge
	judgeInstances := nset.NewStringSet()
	for _, instance := range cfg.Judge.Cluster {
		judgeInstances.Add(instance)
	}
	JudgeConnPools = backend.CreateSafeRpcConnPools(cfg.Judge.MaxConns, cfg.Judge.MaxIdle,
		cfg.Judge.ConnTimeout, cfg.Judge.CallTimeout, judgeInstances.ToSlice())

	// tsdb
	if cfg.Tsdb.Enabled {
		TsdbConnPoolHelper = backend.NewTsdbConnPoolHelper(cfg.Tsdb.Address, cfg.Tsdb.MaxConns, cfg.Tsdb.MaxIdle, cfg.Tsdb.ConnTimeout, cfg.Tsdb.CallTimeout)
	}

	// graph
	graphInstances := nset.NewSafeSet()
	for _, nitem := range cfg.Graph.ClusterList {
		for _, addr := range nitem.Addrs {
			graphInstances.Add(addr)
		}
	}
	GraphConnPools = backend.CreateSafeRpcConnPools(cfg.Graph.MaxConns, cfg.Graph.MaxIdle,
		cfg.Graph.ConnTimeout, cfg.Graph.CallTimeout, graphInstances.ToSlice())

	//added by vincent.zhang for sending to kafka
	if cfg.Kafka.Enabled {
		initKafkaConfig()
		// single kafka producer method
		p, err := NewKafkaProducer(cfg.Kafka.Topic, cfg.Kafka.Address)
		if err == nil {
			kafkaProducer = p
		} else {
			log.Println("new kafka producer fail, error: ", err.Error())
		}
		//end
		// producer pool method
		//KafkaPool = NewKafkaProducerPool(cfg.Kafka.Topic, cfg.Kafka.Address, int32(cfg.Kafka.MaxConcurrent), int32(cfg.Kafka.MaxConcurrent))
		// Added by qimin.xu for kafka filter producer
		if len(cfg.Kafka.Filter) > 0 {
			for f := range cfg.Kafka.Filter {
				fp, err := NewKafkaProducer(cfg.Kafka.Filter[f]["topic"], cfg.Kafka.Address)
				if err == nil {
					kafkaFilterProducer[f] = fp
				} else {
					log.Println("new kafka filter producer fail, error: ", err.Error())
				}
			}
			log.Println("init kafka filter ConnPools success producer: ", kafkaFilterProducer)
		}
		//end
	}

	if cfg.Kafka.LogEnabled {
		initKafkaConfig()
		// single kafka producer method
		p, err := NewKafkaProducer(cfg.Kafka.LogTopic, cfg.Kafka.Address)
		if err == nil {
			kafkaLogProducer = p
		} else {
			log.Println("new kafka Log producer fail, error: ", err.Error())
		}
	}
}

func DestroyConnPools() {
	JudgeConnPools.Destroy()
	GraphConnPools.Destroy()
	TsdbConnPoolHelper.Destroy()
	//added by vincent.zhang for sending to kafka
	kafkaProducer.Close() // single kafka producer method
	kafkaLogProducer.Close()
	//KafkaPool.Destroy()	// producer pool method
}
