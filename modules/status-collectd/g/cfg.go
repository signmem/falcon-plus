package g

import (
	"encoding/json"
	"github.com/signmem/file"
	"log"
	"sync"
)


type Env struct {
	Type 		string 		`json:"type"`
	Name		string		`json:"name"`
	Path 		string		`json:"path"`
}

type GlobalConfig struct {
	Debug			bool            `json:"debug"`
	LogMaxAge		int             `json:"logmaxage"`
	LogRotateAge	int				`json:"logrotateage"`
	LogFile			string			`json:"logfile"`
	DisableSend		bool			`json:"disablesend"`
	EnableCnt 		bool 			`json:"enablecnt"`
	Env 			*Env 			`json:"env"`
	HTTP 			*HTTP 			`json:"http"`
	Cluster 		string			`json:"cluster"`
	Kafka 			*KConfig 		`json:"kafka"`
	Cmdb 			*CMDB 			`json:"cmdb"`
	Redis			*RedisConfig	`json:"redis"`
}

type RedisConfig struct {
	Enabled         bool            `json:"enabled"`
	Server          string          `json:"server"`
	Port            string          `json:"port"`
	MaxIdle                 int                             `json:"max_idle"`
	MaxActive               int                             `json:"max_active"`
	IdleTimeOut             int                             `json:"idle_timeout"`
	LockKey                 string                  `json:"lock_key"`
	LockTime                int                             `json:"lock_time"`
	AskLockTime             int                     `json:"ask_locktime"`
}


type HTTP struct {
	Listen 			string 			`json:"listen"`
	Port			string			`json:"port"`
}

type CMDB struct {
	Url 			string			`json:"url"`
	SysName			string			`json:"sysname"`
	Token 			string 			`json:"token"`
}

type KConfig struct {
	Enable 		bool 			`json:"enable"`
	Topic 		string			`json:"topic"`
	Servers 	[]string 		`json:"servers"`
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
