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

type ConsumerConfig struct {
	KafkaVersion  string   `json:kafkaVersion`
	Topics        []string `json:topics`
	Group         string   `json:group`
	Offset        string   `json:offset`
	OffsetTimeout int      `json:offsetTimeout`
	Concurrent    int      `json:concurrent`
	Zookeeper     string   `json:zookeeper`
	Kafka         []string `json:kafka`
}

type TrendConfig struct {
	Enabled     bool                    `json:"enabled"`
	Batch       int                     `json:"batch"`
	ConnTimeout int                     `json:"connTimeout"`
	CallTimeout int                     `json:"callTimeout"`
	MaxConns    int                     `json:"maxConns"`
	MaxIdle     int                     `json:"maxIdle"`
	Replicas    int                     `json:"replicas"`
	Cluster     map[string]string       `json:"cluster"`
	// ClusterList map[string]*ClusterNode `json:"clusterList"`
}

type TransferConfig struct {
	Enabled     bool              `json:"enabled"`
	Batch       int               `json:"batch"`
	Retry       int               `json:"retry"`
	ConnTimeout int               `json:"connTimeout"`
	CallTimeout int               `json:"callTimeout"`
	MaxConns    int               `json:"maxConns"`
	MaxIdle     int               `json:"maxIdle"`
	Cluster     map[string]string `json:"cluster"`
}

type GlobalConfig struct {
	Debug 		 bool 			`json:"debug"`
	DebugTrend	 bool			`json:"debugtrend"`
	DebugTraffer bool 			`json:"debugtraffer"`
	LogMaxAge	 int             `json:"logmaxage"`
	LogRotateAge int             `json:"logrotateage"`
	LogFile      string  		`json:"logfile"`
	Http         *HttpConfig     `json:"http"`
	Consumer     *ConsumerConfig `json:"consumer"`
	Trend        *TrendConfig    `json:"trend"`
	Transfer     *TransferConfig `json:"transfer"`
	PercentCheck map[string]bool `json:"percent_check"`
	IgnoreHost   map[string]bool `json:"ignore_host"`
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

	/*

	 rewrite by terrytsang
	// split cluster config
	c.Trend.ClusterList = formatClusterItems(c.Trend.Cluster)
	*/

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


/*  rewrite by terrytsang

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

*/