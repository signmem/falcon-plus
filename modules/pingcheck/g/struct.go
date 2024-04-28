package g

import (
	"fmt"
)

type GlobalConfig struct {
	Debug				bool				`json:"debug"`
	ForceCheck			bool				`json:"forcecheck"`
	AlarmEnable			bool				`json:"alarmenable"`
	LogMaxAge			int					`json:"logmaxage"`
	LogRotateAge		int					`json:"logrotateage"`
	AgentExpire  		int64 				`json:"agentexpire"`
	AgentPriority		string				`json:"agentpriority"`
	CheckInterval		int 				`json:"checkinterval"`
	LogFile				string				`json:"logfile"`
	Http				*HttpConfig			`json:"http"`
	Redis				*RedisConfig		`json:"redis"`
	FlushKeyInterval 	int 				`json:"flushkeyinterval"`
	Falcon 				*Falcon 			`json:"falcon"`
	Cmdb 				*Cmdb 				`json:"cmdb"`
	Pigeon				*Pigeon				`json:"pigeon"`
	ExcludeDomains		[]string			`json:"excludedomains"`
	ExcludeDBATags      string 				`json:"excludedbatags"`
	ExcludeDBADeploy	[]int				`json:"excludedbadeploytype"`
	Transfer 			*TransferConfig		`json:"transfer"`
	PingClient			*[]ClientConfig 	`json:"pingclient"`
	Degrade             *Degrade            `json:"degrade"`
	Proxy				*Proxy				`json:"proxy"`
}

type Degrade struct {
	Enabled			bool 			`json:"enabled"`
	Period 			int				`json:"period"`
	AlarmLimit 		int				`json:"alarmlimit"`
	FrozenTime		int				`json:"frozentime"`
}

type ClientConfig struct {
	RoomName		string			`json:"roomname"`
	ShotName		string			`json:"shotname"`
	CheckUrl 		string			`json:"checkurl"`
}

type TransferConfig struct {
	Interval 	int 		`json:"interval"`
	Servers 	[]string	`json:"servers"`
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
	MaxIdle			int				`json:"maxidle"`
	MaxActive		int				`json:"maxactive"`
	IdleTimeOut		int				`json:"idletimeout"`
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

