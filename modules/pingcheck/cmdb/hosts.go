package cmdb

import (
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"github.com/signmem/falcon-plus/modules/pingcheck/tools"
	"regexp"
	"strings"
	"time"
)

var (
	FistTime = true
	FalconCheckRecord []HostInfo
	PingCheckRecord []HostInfo
	AllowCmdbHostInfo []HostInfo
	ReplicateHostInfo map[string][]string
	FalconNotReport []string
	RePingMetrics  []HostInfo
	AgentAlive []HostInfo
	regDisklessClient = regexp.MustCompile(`DISKLESS-CLIENT`)
)

func getHostInfoFromCMDB(diableMonitorDomains []string) (falconCheck []HostInfo,
	pingCheck []HostInfo, replicationHostWithIPS map[string][]string, falconNotReport []string,
	repingHostFilter []HostInfo, agentAliveInRedis []HostInfo) {

	// return info:
	// falconCheck	 检测 falcon-agent 的服务器列表
	// pingCheck     只检测 ping 的服务器列表
	// replicationHostWithIPS  (重复主机名，用于告警) [ip][hostname ... ]
	// falconNotReport  需要上报 agent.alive 但现在没有在 redis  agent.alive/ 中检测到的主机
	// repingHost  = agent 故障 (redis agent.ping, agent.pingdead) +  dbaHost  + cmdb montype != 4 (排除 agent.alive)


	// pingHostList + pingCheck + dbaHost  need ping vevy minute 2025-12-07
	// pingHostList == agent.ping  agent.dead []string  <--- fix
	// dbaHost == dba host list ( []HostInfo)
	// pingCheck ==  CMDB PING ONLY []HostInfo )

	//
	// repingHost == pingHostList + dbaHost + pingCheck
	// repingHost == ping every minute and update agent.ping metrics
	//
	//
	// if repingHost in pingHostList ( do not send message to pigeon )
	//



	// diableMonitorDomains 变量来自 cmdb 中已过滤 dbs 应用

	//api := "/server/query"
	//query := "app_name="

	var debugHostDisable []string
	var allCmdbHost []HostInfo
	var repingHost []HostInfo
	var dbaHost []HostInfo
	var falconAgentNotReport []HostInfo
	replicationHostWithIPS = make(map[string][]string,0)
	// var agentAliveHost []HostInfo


	// debugHostDisable 忽略，不检测 PING 不检测 AGENT 的主机
	// allCmdbHost 所有 cmdb 中的服务器
	// dbaHost 从 CMDB 中过滤的 DBA 主机
	// falconAgentNotReport 没有成功上报的 主机
	// replicationHostWithIPS  重复主机名
	//  agentAliveHost  == agent.alive == success


	// getRedisSkipPingHost() 获取 redis 中 "falcon.ping" , "falcon.pingdaed 主机 list 
	// ** 这里需要注意 pingHostList 变量 中的主机忽略了每分钟执行 PING 操作 

	pingHostList := getRedisSkipPingHost()
	agentAliveList, _ := getAgentAvlive()


	api  := "/v3/server/query"
	query := "fields=type,use_type,status,os_type,server_name,app_name,ip,room_name,monitor_type"

	// allHostInfo   CMDB 中所有主机信息
	// allHostInfo.AppName 对应当前主机应用名

	allHostInfo, err := CmdbApiQuery(api, query)

	if err != nil {
		g.Logger.Errorf("getHostInfoFromCMDB() error: %s", err)
		return nil, nil, make(map[string][]string), nil, nil, nil
	}

	// debugHostDisable  不检测 falcon-agent 及 ping 的服务器 
	// falconCheck	     检测 falcon-agent 的服务器列表
	// pingCheck     只检测 ping 的服务器列表

	if len(allHostInfo.Object) > 0 {

		for _, hostInfo := range allHostInfo.Object {

			// 需要测试 "10.218.205.21","10.209.205.21","10.130.100.63","10.218.17.5","10.209.51.99","10.223.34.46","10.205.26.13","10.141.165.34" ( GTM, SHTEAM 服务器 )
			// 需要重写下面逻辑，待验证特殊设备用法

			// type       0：物理机 1：虚拟机 2: 容器 3:特殊设备 5:云主机
			// use_type   0:生产机; 1:预发布机; 2:计划任务机;3:备机;4:测试机;5:buffer池；6:冷备'
			// status -1：回收; -2:待报废; 0: 库存; 1：上架; 2: 初始化; 3: 部署中; 4: 生产; 5: 下线; 99:SA维护，100:IDC维护，101:测试;
			// os_type 1:centos 2:windows 3:esxi 4:其他 5:ubuntu 6:suse
			// monitor_type 0 其他 1 falcon 4 不可 ping 

			// 意图: 只获取 物理机，虚拟机, 特殊设备(gtm, stheam)   云主机
			// 只针对生产机器及预发布  <-- include all host fix by 2025-12-07
			// 只对生产状态服务器进行检测
			// 只检查 centos, windows 服务器
			// 只检查监控设定为 falcon 监控的服务器 用于检测是否安装 falcon agent
			// 如果 monitor_type = 4 则不执行 ping 检测

			// debugHostDisable (不 ping 不检测 agent ) 直接放弃服务检测的设备   
			// pingCheck  (只检测 ping, monitor_type != [ 1, 4 ])
			// falconCheck (验证 falcon-agent)

			// 临时假如特殊设备

			var host HostInfo
			hostname := hostInfo.ServerName
			ipaddr := hostInfo.Ip
			host.HostName = hostname
			host.DomainName = hostInfo.AppName
			host.IPAddr = ipaddr


			// 不可 ping (容器, 非生产状态，默认不让 ping)
			// Type = 2 (容器)  status = 4 (生产)   [ MonType 4 (不可 ping) 后续处理 ]
			// 这部分主机不需要执行每分钟，每小时 ping 不生成 agent.ping 指标

			if  hostInfo.Type == 2 || hostInfo.Status != 4 || 
				hostInfo.AppName == "" || hostInfo.Ip == "" || hostInfo.ServerName == ""  {

				debugHostDisable = append(debugHostDisable, hostInfo.ServerName)
				continue
			}

			// allCmdbHost  所有服务器主机
			allCmdbHost = append(allCmdbHost, host)

			if _, ok := replicationHostWithIPS[hostname]; ok {

				// 对于重复主机名处理

				ipList := replicationHostWithIPS[hostname]
				if tools.SliceContains(ipaddr, ipList) == false {

					ipList = append(ipList, ipaddr)
					replicationHostWithIPS[hostname] = ipList
				}

				continue
			} else {
				// 唯一主机列表
				var ipList []string
				ipList = append(ipList, ipaddr)
				replicationHostWithIPS[hostname] = ipList

			}

			// debugHostDisable 不需要执行检测 ping 服务器
			if hostInfo.MonType == 4 {
				debugHostDisable = append(debugHostDisable, hostInfo.ServerName)
				continue
			}


			// pingHostList =  redis 中 "agent.ping"  "agent.pingdead" 中的主机 list 
			if len(pingHostList) > 0 {
				if tools.SliceContains(hostInfo.ServerName, pingHostList) == true {
					repingHost = append(repingHost, host)
					continue
				}
			}


			//  DBA 物理机 diableMonitorDomains (包含了 dba 域)
			if  len(diableMonitorDomains) > 0 {
				if tools.SliceContains(hostInfo.AppName, diableMonitorDomains) == true {
					// 假如 diableMonitorDomains 包含了对应应用
					// 与 dba 确认后，对 dba 所有应用都不需要 agent 检测不需要 ping
					// dbHost need ping and upload agent.ping metrics 2025-12-07
					repingHost = append(repingHost, host)
					dbaHost = append(dbaHost,host)
					continue
				}
			}

			// 匹配 DISKLESS-CLIENT 主机名，并忽略处理
			// r, _ := regexp.Compile("DISKLESS-CLIENT")

			if regDisklessClient.MatchString(hostInfo.ServerName) {

			// if r.MatchString(hostInfo.ServerName) == true {
				debugHostDisable = append(debugHostDisable, hostInfo.ServerName)
				continue
			}

			// montype = 1 falcon check
			// montype != [1, 4], pingcheck
			// falconCheck 定义为 montitor_type = 1 需要 falcon agent 检测
			// replicationHostWithIPS 重复主机名变量 

			if hostInfo.MonType == 1 {

				// agentAliveList == 成功上报 redis 服务器主机名)
				// falconAgentNotReport 没有成功上报的 主机

				falconCheck = append(falconCheck, host)

				if tools.SliceContains(hostInfo.ServerName, agentAliveList) == false {
					falconAgentNotReport = append(falconAgentNotReport, host)
					falconNotReport = append(falconNotReport, hostInfo.ServerName)

					// 增加没有上报服务器每分钟上报 agent.ping 用法
					// add all host without montype = 4 to repingHost 2025-03-18
					repingHost = append(repingHost, host)
				} else {

					// AgentAlive  ==  agent.alive = success
					agentAliveInRedis = append(agentAliveInRedis, host)
				}

			} else {
				// 之前已经忽略所有 MonType == 4 服务器

				pingCheck = append(pingCheck, host)
				// add all host without montype = 4 to repingHost 2025-03-18
				repingHost = append(repingHost, host)
			}
			// add all host without montype = 4 to repingHost 2025-03-18
			// repingHost = append(repingHost, host)

		}
	} else {
		g.Logger.Error("GetHostInfoFromCMDB 无法从 cmdb api 中获取任何主机信息，请排查 cmdb:port/v3/server/query 接口是否正常.")
		return
	}

	for hostname, ipaddress := range replicationHostWithIPS {
		if len(ipaddress) == 1 {
			delete(replicationHostWithIPS, hostname)
		}
	}

	repingHostFilter = repingHost

	// repingHost 这里需要增加 falconAgentNotReport 中的主机 方便执行每分钟 PING 操作增加 agent.ping 监控
	// repingHost -> []HostInfo 
	// falconAgentNotReport
	// terry.zeng 2026-03-10

	if g.Config().Debug {
		g.Logger.Debugf("[CMDB CHECK INDEX] ============ start ")
		g.Logger.Debugf("[CMDB CHECK INDEX] Total 物理机: %d", len(allCmdbHost))
		g.Logger.Debugf("[CMDB CHECK INDEX] 不需要检测的设备数量 total %d", len(debugHostDisable))
		g.Logger.Debugf("[CMDB CHECK INDEX] CMDB 配置 falcon 监控的物理机数量 total %d", len(falconCheck))
		g.Logger.Debugf("[CMDB CHECK INDEX] 正常上报 redis 的 falcon 数量为 %d", len(agentAliveList))
		g.Logger.Debugf("[CMDB CHECK INDEX] CMDB 配置 falcon 但没有上报主机 total %d", len(falconAgentNotReport))
		g.Logger.Debugf("[CMDB CHECK INDEX] CMDB 标记只需要 ping 的物理机数量 total %d", len(pingCheck))
		g.Logger.Debugf("[CMDB CHECK INDEX] 重复主机名 total %d", len(replicationHostWithIPS))
		g.Logger.Debugf("[CMDB CHECK INDEX] 因 agent 故障 ping 检测的物理机数量 total %d", len(pingHostList))
		g.Logger.Debugf("[CMDB CHECK INDEX] 每分钟检测 agent.ping 的物理机数量 total %d", len(repingHostFilter))
		g.Logger.Debugf("[CMDB CHECK INDEX] 忽略的 DBA 域主机 total %d", len(dbaHost))
		g.Logger.Debugf("[CMDB CHECK INDEX] ============ end ")
	}

	return
}

func GetCMDBHostLoop() {

	// 每小时从 cmdb 中获取一次主机信息
	// 信息存储至常量  CmdbHostInfoRecord 中
	// 当 FistTime = true 时候，则常量 CmdbHostInfoRecord 为空

	// monitorDomains 
	// getExcludeDomains() retuen 1, alldomain 2 allowdomain 3 diableMonitorDomains
	// 1. alldomain = 从 cmdb 获取所有的 应用名称 (包含容器包含空等所有)
	// 2. allowdomain = 应用包含了物理机,vm, 设定为 monitor_type = 1 的特殊设备的应用
	// 3. diableMonitorDomains 不监控的应用(例如 dba 专用的服务器)
	// CmdbHostInfoRecord == 常量, 存主机信息

	for {
		_, _, diableMonitorDomains := getExcludeDomains()

		// diableMonitorDomains 过滤的主机，需要执行 agent.ping  2025-12-07

		falconCheckRecord, allowPing, replicateHostName, falconNotReport,
		  repingHost, agentAliveHost := getHostInfoFromCMDB(diableMonitorDomains)

		// 只有当 falconCheckRecord > 0 才需要更新 cmdb 信息
		// 预防 cmdb 故障无法返回数据
		//
		// falconCheckRecord CMDB 中正常所有上报的 agent
		// allowPing == CMDB 标记只需要 PING 并且需要报警
		// replicateHostName == 重复主机名，需要报警
		// falconNotReport  == CMDB 标记需要 FALCON 但不上报， 需要报警
		//
		// hostAndIpaddr == []stings 过滤了所有重复的主机名
		// replicateHostName == map[string][]string 重复的主机名

		if len(falconCheckRecord) > 0 {
			FalconCheckRecord = falconCheckRecord
			PingCheckRecord = allowPing
			ReplicateHostInfo = replicateHostName  // replication hostname alarm !!!! terry.zeng
			FalconNotReport = falconNotReport
			AllowCmdbHostInfo = append(PingCheckRecord, FalconCheckRecord...)
			RePingMetrics = repingHost
			AgentAlive = agentAliveHost
		}

		FistTime = false
		time.Sleep( 5 * time.Minute)

	}

}

func CheckHostInList(hostname string, hostInfo []HostInfo) (bool) {

	temp_name := strings.Split(hostname, "@")
	split_name := temp_name[0]

	if len(hostInfo) > 0 {
		for _, info := range hostInfo {
			if split_name == info.HostName {
				return true
			}
		}
	} else {
		return false
	}
	return false
}

func getHostFromHostinfo(hostMap  []HostInfo) (hostList []string) {
	hostList = make([]string, 0, len(hostMap))
	for _, host := range hostMap {
		hostList = append(hostList, host.HostName)
	}
	return
}

