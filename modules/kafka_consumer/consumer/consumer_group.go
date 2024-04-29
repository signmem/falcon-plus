package consumer

import (
	"time"
	"github.com/IBM/sarama"
	"github.com/signmem/falcon-plus/modules/kafka_consumer/g"
)

const (
	DEFAULT_OFFSET  = sarama.OffsetNewest
	DEFAULT_TIMEOUT = 5 * time.Second
)

func initGroup(isHigh bool) error {
	if isHigh {
		return highInitGroup()
	} else {
		return lowInitGroup()
	}
}

func Start() {
	var isHigh bool
	if g.Config().Consumer.KafkaVersion == "high" {
		isHigh = true
	} else {
		isHigh = false
	}
	err := initGroup(isHigh)
	if err != nil {
		g.Logger.Errorf("init consumer group error: %s", err.Error())
		return
	}
	run(isHigh)
	errorPrinter(isHigh)
}

func run(isHigh bool) {
	if isHigh {
		go highRun()
	} else {
		go lowRun()
	}
}

func Stop() {
	if g.Config().Consumer.KafkaVersion == "high" {
		highStop()
	} else {
		lowStop()
	}
}

func errorPrinter(isHigh bool) {
	if isHigh {
		go highErrorPrinter()
	} else {
		go lowErrorPrinter()
	}
}
