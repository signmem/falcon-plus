package g

import (
	"encoding/json"
	"log"
	"sync"
	"github.com/toolkits/file"
)

type HttpConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type RpcConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type DBConfig struct {
	Dsn           string `json:"dsn"`
	Batch         int    `json:"batch"`
	BatchInterval int    `json:"batchInterval"`
	Concurrent    int    `json:"concurrent"`
	MaxIdle       int    `json:"maxIdle"`
}

type RedisConfig struct {
	MaxIdle		int			`json:"maxidle"`
	MaxActive	int			`json:"maxactive"`
	IdleTimeOut	int			`json:"idletimeout"`
	Server 		string		`json:"server"`
	Port		string		`json:"port"`
}

type GlobalConfig struct {
	Debug			bool 		`json:"debug"`
	DBLog  			bool 		`json:"dblog"`
	LogLevel		string      `json:"log_level"`
	LogMaxAge		int		`json:"logmaxage"`
	LogRotateAge	int		`json:"logrotateage"`
	LogFile			string	`json:"logfile"`
	MetricPort 		string 	`json:"metricport"`
	Gauge    		bool        `json:"gauge"`
	Http     		*HttpConfig `json:"http"`
	Rpc      		*RpcConfig  `json:"rpc"`
	DB       		*DBConfig   `json:"db"`
	Redis   		*RedisConfig  `json:"redis"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	configLock = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	configLock.Lock()
	defer configLock.Unlock()
	config = &c

	log.Println("g.ParseConfig ok, file ", cfg)
}
