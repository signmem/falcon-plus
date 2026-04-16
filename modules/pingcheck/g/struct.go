package g

import (
	"fmt"
	"sync"
)

type GlobalConfig struct {
	Debug			bool				`json:"debug"`
	AgentCheck		bool				`json:"agent_check"`   // 测试用检测 agent 安装
	AlarmEnable		bool				`json:"alarm_enable"` // 允许告警
	LogMaxAge		int					`json:"log_max_age"`  // 日志最大时间
	LogRotateAge		int				`json:"log_rotate_age"` // 回滚日志时间
	AgentExpire  		int64 			`json:"agent_expire"`    // agent 过期告警
	PingDeadExpire 		int64			`json:"ping_dead_expire"` // ping 过期告警
	FalconRePingConcurrent	int			`json:"falcon_reping_concurrent"`  // reping concurrent
	FalconPingConcurrent	int			`json:"falcon_ping_concurrent"`   // falcon.ping concurrent
	AgentPriority		string				`json:"agent_priority"`   // 告警级别
	CheckInterval		int 				`json:"check_interval"`   // 检测间隔时间
	CheckPingInterval	int					`json:"check_ping_interval"`
	LogFile				string				`json:"log_file"`
	AlarmFile 			string				`json:"alarm_file"`
	ExcludeBusGroup		[]string			`json:"exclude_bus_group"`
	ExcludeAppName 		[]string			`json:"exclude_app_name"`
	Http				*HttpConfig			`json:"http"`
	Redis				*RedisConfig		`json:"redis"`
	Falcon 				*Falcon 			`json:"falcon"`
	FalconAgent			*FalconAgent		`json:"falcon_agent"`
	Cmdb 				*Cmdb 				`json:"cmdb"`
	Pigeon				*Pigeon				`json:"pigeon"`
	Transfer 			*TransferConfig		`json:"transfer"`
	Degrade             *Degrade            `json:"degrade"`
	Proxy				*Proxy				`json:"proxy"`
}

type Degrade struct {
	Enabled		bool 	`json:"enabled"`		// 降级控制`
	Period 		int		`json:"period"`			// 时间窗口
	AlarmLimit 	int		`json:"alarm_limit"`	// 时间窗口内允许发生的最大报警
	FrozenTime	int		`json:"frozen_time"`	// 降级时间 (分钟)
}


type TransferConfig struct {
	Interval 	int 		`json:"interval"`
	Servers 	[]string	`json:"servers"`
	Addrs		[]string	`json:"addrs"`
	Timeout		int			`json:"timeout"`
}

type HttpPingRequest struct {
	Ipaddr		string		`json:"ipaddr"`
}

type HttpPingResponse struct {
	PingStatus		bool		`json:"pingcheck"`
	Ipaddr			string		`json:"ipaddr"`
}

type Cmdb struct {
	Url 		string 			`json:"url"`
	SysName 	string 			`json:"sysname"`
	Token 		string 			`json:"token"`
}

type HttpConfig struct {
	Enabled  bool   `json:"enabled"`
	Listen   string `json:"listen"`
}

type Falcon struct {
	FalconAuthName 		string 				`json:"falconauth"`
	FalconAuthSig 		string 				`json:"falconsign"`
	Url 				string 				`json:"url"`
}

type FalconAgent struct {
	Address 			string				`json:"address"`
	Port 				string				`json:"port"`
	Step 				int64 				`json:"step"`
}

type Proxy struct {
	Servers 			[]string	`json:"servers"`
}

type Pigeon struct {
	PigeonSource 		string 		`json:"source"`
	PigeonKey			string 		`json:"key"`
	M3dbUrl 			string 		`json:"m3dburl"`
	PigeonUrl 			string		`json:"pigeonurl"`
}


type RedisConfig struct {
	Enabled         bool            `json:"enabled"`
	Server          string          `json:"server"`
	Port            string          `json:"port"`
	MaxIdle			int				`json:"max_idle"`
	MaxActive		int				`json:"max_active"`
	IdleTimeOut		int				`json:"idle_timeout"`
	LockKey			string 			`json:"lock_key"`
	LockTime 		int				`json:"lock_time"`
	AskLockTime		int 			`json:"ask_locktime"`
}


type LruCache struct {
	Timestamp 		int64			`json:"timestamp"`
	HostList		[]string		`json:"hostlist"`
}

func (l LruCache) String() string {
	return fmt.Sprintf("time: %d, count: %d, list: %v", l.Timestamp,
		len(l.HostList), l.HostList )
}

func (l LruCache) Len() int {
	return len(l.HostList)
}

func (l LruCache) HostDetail() string {
	var detail string
	i := 1
	for _, info := range l.HostList {
		if i == 1  {
			detail = info
			i += 1
		} else {
			detail = detail + "," + info
		}
	}
	return fmt.Sprintf("%s", detail)
}


type Lru struct {
	Mu      sync.Mutex
	TotalLru map[string][]string
}


type SCount struct {
	sync.RWMutex
	Name    string          `json:"name"`
	Cnt     int64           `json:"cnt"`
	Time    int64           `json:"time"`
}