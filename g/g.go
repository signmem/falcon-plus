package g

import "path/filepath"

var Modules map[string]bool
var BinOf map[string]string
var cfgOf map[string]string
var ModuleApps map[string]string
var logpathOf map[string]string
var PidOf map[string]string
var AllModulesInOrder []string

func init() {
	Modules = map[string]bool{
		"agent":          true,
		"aggregator":     true,
		"graph":          true,
		"hbs":            true,
		"judge":          true,
		"nodata":         true,
		"transfer":       true,
		"gateway":        true,
		"api":            true,
		"alarm":          true,
		"kafka_consumer": true,
		"trend":          true,
	}

	BinOf = map[string]string{
		"agent":          "./agent/bin/falcon-agent",
		"aggregator":     "./aggregator/bin/falcon-aggregator",
		"graph":          "./graph/bin/falcon-graph",
		"hbs":            "./hbs/bin/falcon-hbs",
		"judge":          "./judge/bin/falcon-judge",
		"nodata":         "./nodata/bin/falcon-nodata",
		"transfer":       "./transfer/bin/falcon-transfer",
		"gateway":        "./gateway/bin/falcon-gateway",
		"api":            "./api/bin/falcon-api",
		"alarm":          "./alarm/bin/falcon-alarm",
		"kafka_consumer": "./kafka_consumer/bin/falcon-kafka_consumer",
		"trend":          "./trend/bin/falcon-trend",
	}

	cfgOf = map[string]string{
		"agent":          "./agent/config/cfg.json",
		"aggregator":     "./aggregator/config/cfg.json",
		"graph":          "./graph/config/cfg.json",
		"hbs":            "./hbs/config/cfg.json",
		"judge":          "./judge/config/cfg.json",
		"nodata":         "./nodata/config/cfg.json",
		"transfer":       "./transfer/config/cfg.json",
		"gateway":        "./gateway/config/cfg.json",
		"api":            "./api/config/cfg.json",
		"alarm":          "./alarm/config/cfg.json",
		"kafka_consumer": "./kafka_consumer/config/cfg.json",
		"trend":          "./trend/config/cfg.json",
	}

	ModuleApps = map[string]string{
		"agent":          "falcon-agent",
		"aggregator":     "falcon-aggregator",
		"graph":          "falcon-graph",
		"hbs":            "falcon-hbs",
		"judge":          "falcon-judge",
		"nodata":         "falcon-nodata",
		"transfer":       "falcon-transfer",
		"gateway":        "falcon-gateway",
		"api":            "falcon-api",
		"alarm":          "falcon-alarm",
		"kafka_consumer": "falcon-kafka_consumer",
		"trend":          "falcon-trend",
	}

	logpathOf = map[string]string{
		"agent":          "./agent/logs/agent.log",
		"aggregator":     "./aggregator/logs/aggregator.log",
		"graph":          "./graph/logs/graph.log",
		"hbs":            "./hbs/logs/hbs.log",
		"judge":          "./judge/logs/judge.log",
		"nodata":         "./nodata/logs/nodata.log",
		"transfer":       "./transfer/logs/transfer.log",
		"gateway":        "./gateway/logs/gateway.log",
		"api":            "./api/logs/api.log",
		"alarm":          "./alarm/logs/alarm.log",
		"kafka_consumer": "./kafka_consumer/logs/kafka_consumer.log",
		"trend":          "./trend/logs/trend.log",
	}

	PidOf = map[string]string{
		"agent":          "<NOT SET>",
		"aggregator":     "<NOT SET>",
		"graph":          "<NOT SET>",
		"hbs":            "<NOT SET>",
		"judge":          "<NOT SET>",
		"nodata":         "<NOT SET>",
		"transfer":       "<NOT SET>",
		"gateway":        "<NOT SET>",
		"api":            "<NOT SET>",
		"alarm":          "<NOT SET>",
		"kafka_consumer": "<NOT SET>",
		"trend":          "<NOT SET>",
	}

	// Modules are deployed in this order
	AllModulesInOrder = []string{
		"graph",
		"hbs",
		"judge",
		"transfer",
		"nodata",
		"aggregator",
		"agent",
		"gateway",
		"api",
		"alarm",
		"kafka_consumer",
		"trend",
	}
}

func Bin(name string) string {
	p, _ := filepath.Abs(BinOf[name])
	return p
}

func Cfg(name string) string {
	p, _ := filepath.Abs(cfgOf[name])
	return p
}

func LogPath(name string) string {
	p, _ := filepath.Abs(logpathOf[name])
	return p
}

func LogDir(name string) string {
	d, _ := filepath.Abs(filepath.Dir(logpathOf[name]))
	return d
}
