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

type RedisConfig struct {
	Addr            string   `json:"addr"`
	AddrCMDBCache   string   `json:"addrCMDBCache"`
	MaxIdle         int      `json:"maxIdle"`
	HighQueues      []string `json:"highQueues"`
	LowQueues       []string `json:"lowQueues"`
	UserIMQueue     string   `json:"userIMQueue"`
	UserSmsQueue    string   `json:"userSmsQueue"`
	UserMailQueue   string   `json:"userMailQueue"`
	PigeonHighQueue string   `json:"pigeonHighQueue"` //add by vincent.zhang for combining high level alarms
	PigeonLowQueue  string   `json:"pigeonLowQueue"`  //add by vincent.zhang for combining high low alarms
}

type ApiConfig struct {
	Sms          string `json:"sms"`
	Mail         string `json:"mail"`
	Dashboard    string `json:"dashboard"`
	CMDB         string `json:"cmdb"`
	PlusApi      string `json:"plus_api"`
	PlusApiToken string `json:"plus_api_token"`
	IM           string `json:"im"`
}

//add by vincent.zhang for pigeon
type CombinerConfig struct {
	Levels   []int `json:"levels"`
	Intervel int   `json:"interval"`
}

type PigeonConfig struct {
	AlarmAddr    string          `json:"alarm_addr"`
	OKAddr       string          `json:"ok_addr"`
	FidAddr      string          `json:"fid_addr"`
	HighCombiner *CombinerConfig `json:"high_combiner"`
	LowCombiner  *CombinerConfig `json:"low_combiner"`
}

//add end

type FalconPortalConfig struct {
	Addr string `json:"addr"`
	Idle int    `json:"idle"`
	Max  int    `json:"max"`
}

type WorkerConfig struct {
	IM     int `json:"im"`
	Sms    int `json:"sms"`
	Mail   int `json:"mail"`
	Pigeon int `json:"pigeon"` //add pigeon by vincent.zhang
}

type HousekeeperConfig struct {
	EventRetentionDays int `json:"event_retention_days"`
	EventDeleteBatch   int `json:"event_delete_batch"`
}

type GlobalConfig struct {
	LogLevel     string              `json:"log_level"`
	FalconPortal *FalconPortalConfig `json:"falcon_portal"`
	Http         *HttpConfig         `json:"http"`
	Redis        *RedisConfig        `json:"redis"`
	SendOK       bool                `json:"send_ok"`       //add by vincent.zhang for ok event
	ChangeIgnore bool                `json:"change_ignore"` //add by vincent.zhang for ignore status, 已被ignore的告警再来problem event是否改变为unresolved，并发送
	SendMoreMax  bool                `json:"send_more_max"` //add by vincent.zhang for ignore status, 已达到最大告警次数的告警再来problem event是否发送
	Pigeon       *PigeonConfig       `json:"pigeon"`        //add by vincent.zhang for pigeon
	Api          *ApiConfig          `json:"api"`
	Worker       *WorkerConfig       `json:"worker"`
	Housekeeper  *HousekeeperConfig  `json:"Housekeeper"`
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
		log.Fatalln("config file:", cfg, "is not existent")
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
	log.Println("read config file:", cfg, "successfully")
}
