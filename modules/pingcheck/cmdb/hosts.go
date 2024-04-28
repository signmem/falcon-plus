package cmdb

import (
	"github.com/open-falcon/falcon-plus/common/redisdb"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/g"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/tools"
	"strconv"
	"strings"
	"time"
)

var (
	FistTime = true
	CmdbHostInfoRecord []HostInfo
)

func GetHostInfoFromCMDB(monitorDomains []string) ( monitorHostInfo []HostInfo) {
	// monitorDomains 变量来自 cmdb 中需要检测的应用名 (已过滤 dbs 相应应用)

	//api := "/server/query"
	//query := "app_name="

	api  := "/v3/server/query"
	query := "fields=type,use_type,status,os_type,server_name,app_name,ip,room_name"

        // allHostInfo CMDB 中所有主机信息
        // allHostInfo.AppName 对应当前主机应用名

	allHostInfo, err := CmdbApiQuery(api, query)

	if err != nil {
		g.Logger.Debugf("GetAllMonitorHosts() error: %s", err)
		return
	}


	// debugDomain 只包含不需要检测主机名信息
        // debugStatus 只包含状态不是生产的主机名信息，不是 linux, hypervision, vm 的服务器类型
        // debugOS     只包含不是 linux, windows 主机
        // monitorHostInfo 包含所有合法，需要检测的主机

	var debugDomain []string
	var debugStatus []string
	var debugOS []string


	if len(allHostInfo.Object) > 0 {

		for _, hostInfo := range allHostInfo.Object {

			if tools.SliceContains(hostInfo.AppName, monitorDomains) == false {
				debugDomain = append(debugDomain, hostInfo.ServerName)
				continue
			}

			// type       0：物理机 1：虚拟机 2: 容器 3:特殊设备
			// use_type   0:生产机; 1:预发布机;2:计划任务机;3:备机;4:测试机;5:buffer池；6:冷备'
			// status -1：回收; -2:待报废; 0: 库存; 1：上架; 2: 初始化; 3: 部署中; 4: 生产; 5: 下线; 99:SA维护，100:IDC维护，101:测试;
			// os_type 1:centos 2:windows 3:esxi 4:其他 5:ubuntu 6:suse

			if  hostInfo.Type > 1 || hostInfo.UseType > 1 ||
				hostInfo.Status != 4 || hostInfo.OSType == "" {

					debugStatus = append(debugStatus, hostInfo.ServerName)
					continue
			}

			ostype, err := strconv.Atoi(hostInfo.OSType)

			if err != nil {
				debugOS = append(debugOS, hostInfo.ServerName)
				continue
			}
			if ostype > 2 {
				debugOS = append(debugOS, hostInfo.ServerName)
				continue
			}


			var host HostInfo
			host.HostName = hostInfo.ServerName
			host.DomainName = hostInfo.AppName
			host.IPAddr = hostInfo.Ip

			if strings.Contains(hostInfo.RoomName, "佛山开普勒") {
				host.RoomName = "fskaipule"
			} else	if strings.Contains(hostInfo.RoomName, "佛山五沙") {
				host.RoomName = "fswusha"
			} else 	if strings.Contains(hostInfo.RoomName, "广州南沙") {
				host.RoomName = "gznansha"
			} else {
				host.RoomName = "other"
			}

			monitorHostInfo = append(monitorHostInfo, host)

		}
	}

	if g.Config().Debug {
		g.Logger.Debugf("因为不合法域名而跳过主机名字 total %d", len(debugDomain))
		g.Logger.Debugf("因为不合主机法状态及主机类型而跳过主机名 total %d", len(debugStatus))
		g.Logger.Debugf("因为不合法操作系统而跳过主机名字 total %d", len(debugOS))
		g.Logger.Debugf("合法监控中的主机 total %d", len(monitorHostInfo))
	}

	return
}

func GetAllHostInRedis() (redisHosts []string) {

	service := "agent.alive"
	timeOut := int64(g.Config().AgentExpire)
	normalHost, expireHost, _, err := redisdb.RedisServiceExprieScan(service, timeOut)
	if err != nil {
		g.Logger.Errorf("redisdb.RedisServiceExprieScan error: %s", err)
		return
	}

	redisHosts = append(normalHost, expireHost...)

	return
}

func GetCMDBHostEveryHour() {

	// 每小时从 cmdb 中获取一次主机信息
	// 信息存储至常量  CmdbHostInfoRecord 中
	// 当 FistTime = true 时候，则常量 CmdbHostInfoRecord 为空

	// monitorDomains get domain_names
	// GetExcludeDomains() retuen 1, alldomain 2 allowdomain 3 excludedomain
	// CmdbHostInfoRecord == constant  variable

	// exclude dba tag and deploytype in GetExcludeDomains()

	for {
		_, monitorDomains, _ := GetExcludeDomains()

		CmdbHostInfoRecord = GetHostInfoFromCMDB(monitorDomains)

		FistTime = false

		if g.Config().Debug {
			g.Logger.Debugf("GetCMDBHostEveryHour() total %d hosts in cmdb", len(CmdbHostInfoRecord))
		}

		time.Sleep( 3600 * time.Second)

	}

}






