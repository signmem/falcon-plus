package g

import (
	"encoding/json"
	"github.com/toolkits/file"
	"log"
	"strings"
	"sync"
)

type HttpConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type RpcConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type SocketConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
	Timeout int    `json:"timeout"`
}

type RedisConfig struct {
	MaxIdle		int			`json:"maxidle"`
	MaxActive	int			`json:"maxactive"`
	IdleTimeOut	int			`json:"idletimeout"`
	Server 		string		`json:"server"`
	Port		string		`json:"port"`
}

type JudgeConfig struct {
	Enabled     bool                    `json:"enabled"`
	Batch       int                     `json:"batch"`
	ConnTimeout int                     `json:"connTimeout"`
	CallTimeout int                     `json:"callTimeout"`
	MaxConns    int                     `json:"maxConns"`
	MaxIdle     int                     `json:"maxIdle"`
	Replicas    int                     `json:"replicas"`
	Cluster     map[string]string       `json:"cluster"`
	ClusterList map[string]*ClusterNode `json:"clusterList"`
}

type GraphConfig struct {
	Enabled     bool                    `json:"enabled"`
	Batch       int                     `json:"batch"`
	ConnTimeout int                     `json:"connTimeout"`
	CallTimeout int                     `json:"callTimeout"`
	MaxConns    int                     `json:"maxConns"`
	MaxIdle     int                     `json:"maxIdle"`
	Replicas    int                     `json:"replicas"`
	Cluster     map[string]string       `json:"cluster"`
	ClusterList map[string]*ClusterNode `json:"clusterList"`
}

type TsdbConfig struct {
	Enabled     bool   `json:"enabled"`
	Batch       int    `json:"batch"`
	ConnTimeout int    `json:"connTimeout"`
	CallTimeout int    `json:"callTimeout"`
	MaxConns    int    `json:"maxConns"`
	MaxIdle     int    `json:"maxIdle"`
	MaxRetry    int    `json:"retry"`
	Address     string `json:"address"`
}

//added by vincent.zhang for sending to kafka
type KafkaConfig struct {
	Enabled       bool                         `json:"enabled"`
	LogEnabled    bool                         `json:"logEnabled"`
	Batch         int                          `json:"batch"`
	ConnTimeout   int                          `json:"connTimeout"`
	CallTimeout   int                          `json:"callTimeout"`
	MaxConcurrent int                          `json:"maxConcurrent"`
	MaxRetry      int                          `json:"retry"`
	Address       []string                     `json:"address"`
	Topic         string                       `json:"topic"`
	LogTopic      string                       `json:"logTopic"`
	Filter        map[string]map[string]string `json:"filter"` // added by qimin.xu for filter tag
}

type GlobalConfig struct {
	Debug   bool          `json:"debug"`
	IllegalChar   []string   `json:"illegalchar"`
	MinStep int           `json:"minStep"` //最小周期,单位sec
	Http    *HttpConfig   `json:"http"`
	Rpc     *RpcConfig    `json:"rpc"`
	Socket  *SocketConfig `json:"socket"`
	Judge   *JudgeConfig  `json:"judge"`
	Graph   *GraphConfig  `json:"graph"`
	Tsdb    *TsdbConfig   `json:"tsdb"`
	Kafka   *KafkaConfig  `json:"kafka"` //added by vincent.zhang for sending to kafka
	Redis   *RedisConfig  `json:"redis"`
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

	// split cluster config
	c.Judge.ClusterList = formatClusterItems(c.Judge.Cluster)
	c.Graph.ClusterList = formatClusterItems(c.Graph.Cluster)

	configLock.Lock()
	defer configLock.Unlock()
	config = &c

	log.Println("g.ParseConfig ok, file ", cfg)
}

// CLUSTER NODE
type ClusterNode struct {
	Addrs []string `json:"addrs"`
}

func NewClusterNode(addrs []string) *ClusterNode {
	return &ClusterNode{addrs}
}

// map["node"]="host1,host2" --> map["node"]=["host1", "host2"]
func formatClusterItems(cluster map[string]string) map[string]*ClusterNode {
	ret := make(map[string]*ClusterNode)
	for node, clusterStr := range cluster {
		items := strings.Split(clusterStr, ",")
		nitems := make([]string, 0)
		for _, item := range items {
			nitems = append(nitems, strings.TrimSpace(item))
		}
		ret[node] = NewClusterNode(nitems)
	}

	return ret
}
