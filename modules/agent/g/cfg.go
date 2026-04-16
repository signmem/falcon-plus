package g

import (
	"encoding/json"
	"github.com/signmem/file"
	"log"
	"os"
	"sync"
)

type PluginConfig struct {
	Enabled bool   `json:"enabled"`
	Dir     string `json:"dir"`
	Git     string `json:"git"`
	LogDir  string `json:"logs"`
}

type HeartbeatConfig struct {
	Enabled  bool   `json:"enabled"`
	Addr     string `json:"addr"`
	Interval int    `json:"interval"`
	Timeout  int    `json:"timeout"`
}

type TransferConfig struct {
	Enabled  bool     `json:"enabled"`
	Addrs    []string `json:"addrs"`
	Interval int      `json:"interval"`
	Timeout  int      `json:"timeout"`
}

type HttpConfig struct {
	Enabled  bool   `json:"enabled"`
	Listen   string `json:"listen"`
	Backdoor bool   `json:"backdoor"`
}

type CollectorConfig struct {
	IfacePrefix []string `json:"ifacePrefix"`
	MountPoint  []string `json:"mountPoint"`
}

type VMconfig struct {
	SkipSwapMonitor		bool	`json:"skipswapmonitor"`
	Virtual 			bool	`json:"virtual"`
}

type GlobalConfig struct {
	Debug			bool				`json:"debug"`
	IllegalChar     []string            `json:"illegalchar"`
	LogFile  		string				`json:"logfile"`
	RsyncUser		string				`json:"rsyncuser"`
	RsyncPwd		string				`json:"rsyncpwd"`
	Hostname		string				`json:"hostname"`
	IP				string				`json:"ip"`
	Plugin			*PluginConfig     `json:"plugin"`
	Heartbeat		*HeartbeatConfig  `json:"heartbeat"`
	Transfer		*TransferConfig   `json:"transfer"`
	RsyncAccess		string            `json:"rsyncaccess"`
	Http			*HttpConfig       `json:"http"`
	Collector		*CollectorConfig  `json:"collector"`
	DefaultTags		map[string]string `json:"default_tags"`
	IgnoreMetrics	map[string]bool   `json:"ignore"`
	Pidfile			string            `json:"pidfile"`
	CmdbFile		string				`json:"cmdbfile"`
	VMConfig 		string 				`json:"vmconfig"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	lock       = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

func Hostname() () {
	setHostname := Config().Hostname
	if setHostname != "" {
		HostName = setHostname
	}

	setHostname, err := os.Hostname()
	if err != nil {
		log.Printf("ERROR: os.Hostname() fail %s", err)
		HostName = "not-known"
	}

	HostName = setHostname
}

func IP() string {
	ip := Config().IP
	if ip != "" {
		// use ip in configuration
		return ip
	}

	if len(LocalIp) > 0 {
		ip = LocalIp
	}

	return ip
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

	lock.Lock()
	defer lock.Unlock()

	config = &c

	log.Println("read config file:", cfg, "successfully")
}

type PromConfig struct {
	SslEnable			bool			`json:"ssl_enable"`
	ServerConfig 		*PromServer		`json:"prometheus"`
	TLS					*SSLConfig		`json:"tls"`
	ValidMetricFile		string			`json:"validmetricfile"`
	CalMetricFile		string			`json:"calmetricfile"`
	SumMetricFile 		string			`json:"summetricfile"`
	Step 				int64			`json:"step"`
	Debug 				bool 			`json:"debug"`
}

type PromServer struct {
	Server 			string	`json:"server"`
	Port 			string	`json:"port"`
	MetricAPI		string	`json:"metric_api"`
}

type SSLConfig struct {
	CaFile 			string		`json:"cafile"`
	CertFile 		string		`json:"certfile"`
	KeyFile 		string		`json:"keyfile"`
}

type MetricCalType struct {
	MetricSum       string          `json:"metricsum"`
	MetricCount     string          `json:"metriccount"`
	MetricName      string          `json:"metricname"`
}


type SnmpServer struct {
	Debug 			bool 			`json:"debug"`
	SnmpInfo		*SnmpDetail		`json:"snmpdetail"`
	SnmpFile 		string 			`json:"snmpfile"`
}

type SnmpDetail struct {
	Ipaddr 			string 			`json:"ipaddr"`
	Port 			uint16 			`json:"port"`
	Community 		string			`json:"community"`
	Version 		int				`json:"version"`
	Timeout 		int64 			`json:"timeout"`
	Retry 			int 			`json:"retry"`
	Step 			int64			`json:"step"`
	HostName 		string			`json:"hostname"`
}


type SnmpOID struct {
	Oids                    []OIDMAP        `json:"oids"`
	OidWalks                []OidWalk       `json:"oidwalks"`
}

type OIDMAP struct {
	OID     string                  `json:"oid"`
	Alias   string                  `json:"alias"`
	Type    string                  `json:"type"`
}

type OidWalk struct {
	TagName                         string          `json:"tagname"`
	TagOid                          string          `json:"tagoid"`
	Check                           []OIDMAP        `json:"check"`
}

type OidStruct struct {
        WalkNum                 string                  `json:"walknum"`
        WalkReturn              string                  `json:"walkreturn"`
}

