package falcon

import (
	"fmt"
	"github.com/signmem/falcon-plus/modules/pingcheck/cmdb"
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
	"github.com/signmem/falcon-plus/modules/pingcheck/net"
	"github.com/signmem/falcon-plus/modules/pingcheck/pigeon"
	"os"
	"strconv"
	"strings"
)

func SendAlarm(host string, appDomain string, checkDone bool, agentCheck string) () {

	// 用于发送告警至 pigeon
	// host = 告警主机名
	// appDomain 假如不为空，则以 appDomain 为告警域名为准,否则从 cmdb 获取
	// 当 checkDone 为 true 则认为之前已经执行过 pingcheck 并且 ping 结果为不可 ping 

	// var falconIpaddr string
	var cmdbIpaddr string
	var cmdbDomain string
	var status int
	var pigeonInfo pigeon.Alarm
	var admin string
	var adminName string

	config := g.Config()
	cmdbApi := "/server/query"
	cmdbQuery := "server_name=" + host
	cmdbInfo, err := cmdb.CmdbApiQuery(cmdbApi, cmdbQuery)

	if err != nil {
		g.Logger.Errorf("[SendAlarm] CmdbApiQuery() get host %s info error.", host)

		// do delete host from falcon api
		// DeleteFlaonHost(host)
		g.FalconAgentAliveAlarmFailure.Incr()
		return
	}

	if len(cmdbInfo.Object) > 0 {
		cmdbIpaddr = cmdbInfo.Object[0].Ip

		if appDomain != "" {
			cmdbDomain = appDomain
		} else {
			cmdbDomain = cmdbInfo.Object[0].AppName
		}

		status = cmdbInfo.Object[0].Status
		userType := cmdbInfo.Object[0].UseType
		cmdbType := cmdbInfo.Object[0].Type


		if status != 4  || cmdbType == 2 || userType > 1 || cmdbIpaddr ==  "" ||  cmdbDomain == ""{
			// 非生产状态主机，无需告警

			g.Logger.Warningf("[Falcon Agent Alarm CMDB Skip] app:%s, host:%s ip:%s, status:%d," +
				"type:%d, usertype:%d", cmdbDomain, host, cmdbIpaddr, status, cmdbType, userType)
			return
		}

		admin = cmdbInfo.Object[0].SysAdmin
		adminName = cmdbInfo.Object[0].SysAdminName

	} else {
		g.Logger.Errorf("[CMDB error] get host %s length is zero!!", host)
		g.FalconAgentAliveAlarmFailure.Incr()
		return
	}

	// pingStatus, err := net.CheckPing(cmdbIpaddr)
	// net.CheckPing  must running under by root

	var pingStatus bool

	if checkDone == false {
		pingStatus, _ = net.PingFromProxy(cmdbIpaddr)
	} else {
		pingStatus = false
	}

	pigeonInfo.Domain = cmdbDomain
	pigeonInfo.Ip = cmdbIpaddr
	pigeonInfo.Metric = "agent.alive"
	pigeonInfo.Hostname = host
	pigeonInfo.Value = "-1"
	pigeonInfo.Priority = config.AgentPriority

	if pingStatus == false {

		pigeonInfo.Metric = "agent.ping"

		if cmdb.CheckHostInList(host, cmdb.AllowCmdbHostInfo) == false {
			// pingStatus = false 服务器不可以 ping 
			// cmdb.AllowCmdbHostInfo = CMDB 中 除了不可 ping 的服务器，可报警范围
			g.Logger.Infof("[ALARM ALARM SKIP] 主机: %s  ip: %s   %s 故障，但不在合法 agent 检测 list 中", host, cmdbIpaddr, pigeonInfo.Metric)
			return
		}

		pigeonInfo.Fid = 66043
		pigeonInfo.AlarmCode = "122-000"
		// g.Logger.Infof("[ALARM] PingStatus() ping %s false.", cmdbIpaddr)
		pigeonInfo.Subject = fmt.Sprintf("[PING] 主机名: %s IP: %s ping critical", host, cmdbIpaddr)
		pigeonInfo.Message = fmt.Sprintf("[PING] 主机名: %s IP: %s 无法被 ping 通，常见为物理机故障。", host, cmdbIpaddr)
	} else {

		pigeonInfo.Metric = "agent.alive"

		if cmdb.CheckHostInList(host, cmdb.FalconCheckRecord) == false {
			// agentCheck = redis == 安装 falcon-agent 并上报 REDIS 服务器
			// 只有 cmdb.FalconCheckRecord 中的服务器允许上报 falcon-agent 故障信息
			g.Logger.Infof("[ALARM ALARM SKIP] 主机: %s  ip: %s   %s 故障，但不在合法 agent 检测 list 中", host, cmdbIpaddr, pigeonInfo.Metric)
			return
		}

		pigeonInfo.Fid = 14237
		pigeonInfo.AlarmCode = "122-000"
		// g.Logger.Infof("[ALARM] PingStatus() falcon agent %s false.", cmdbIpaddr)
		pigeonInfo.Subject = fmt.Sprintf("[FALCON] 主机名: %s IP: %s falcon agent 故障", host, cmdbIpaddr)
		pigeonInfo.Message = fmt.Sprintf("[FALCON] 主机名: %s IP: %s falcon agent 故障, 参考 WIKI https://wiki.corp.vipshop.com/pages/viewpage.action?pageId=2236291330 进行 falcon-agent 故障排查", host, cmdbIpaddr)
	}

	// g.FalconAgentLru  -> use to count Lru every host.
	// 统一把 falcon.agent 故障及 agent 故障且 ping 不通服务器放入  g.FalconAgentLru 计数器

	hostinfo := cmdbDomain + "@" + host + "@" + cmdbIpaddr

	if agentCheck == "redis" {


		g.FalconAgentLru.Append(hostinfo)

		if config.Debug {
			g.Alarmer.Debugf("[FALCON AGENT 告警记录] %s",	pigeonInfo.Subject)
		}
	}

	if agentCheck == "cmdb" {


		if config.Debug == true {
			g.Logger.Debugf("[CMDB AGENT CHECK 告警记录] 应用: %s 主机名: %s IP: %s " +
				"admin: %s 管理员: %s metrics: %s", cmdbDomain, host, cmdbIpaddr, admin,
				adminName, pigeonInfo.Metric)

			g.Alarmer.Debugf("[CMDB AGENT CHECK 告警记录] %s", pigeonInfo.Subject)
		}

		if config.AgentCheck == false {
			// terry.zeng (测试用, 直接 return 不发送报警)
			return
		}
	}

	if config.AlarmEnable  == true {
		g.FalconAgentAliveAlarmSuccess.Incr()
		_ = pigeon.SendPigeonAlarm(pigeonInfo)
	}
	return
}


func SendInternalAlarm() {

	// 用于发送告警至 pigeon
	// host = 告警主机名
	// appDomain 假如不为空，则以 appDomain 为告警域名为准,否则从 cmdb 获取

	config := g.Config()
	var pigeonInfo pigeon.Alarm

	pigeonInfo.Fid = 72129
	pigeonInfo.AlarmCode = "232-000"
	pigeonInfo.Value =  strconv.Itoa(g.FalconAgentLru.GetLength(0))
	pigeonInfo.Priority = config.AgentPriority
	pigeonInfo.Domain = "falcon-pingcheck.vip.vip.com"
	pigeonInfo.Ip = g.GetIP()
	pigeonInfo.Metric = "falcon.agentcheck.degrade"
	pigeonInfo.Hostname, _ = os.Hostname()

	pigeonInfo.Subject = fmt.Sprintf("[FALCON AGENT CHECK 降级通知] IP: %s 最近 %d 分钟产生了 %d 个 " +
		"falcon agent check 告警", pigeonInfo.Ip, config.Degrade.Period,
		g.FalconAgentLru.GetLength(0))

	/// terry.zeng !!!!!
	pigeonInfo.Message = fmt.Sprintf("[FALCON AGENT CHECK 降级通知] 最近 %d 分钟产生了 %d 个" +
		"falcon agent check 故障告警，详细主机信息: %s", config.Degrade.Period,
		g.FalconAgentLru.GetLength(0), g.FalconAgentLru.GetLruDetail(2))

	if config.Debug == true {
		g.Alarmer.Debugf("[FALCON AGENT CHECK 降级通知] %s",
			g.FalconAgentLru.GetLruDetail(0))
	}

	if config.AlarmEnable == true {
		_ = pigeon.SendPigeonAlarm(pigeonInfo)
	} 

}


func SendPingDegardAlarm() {

	// 用于发送告警至 pigeon
	// host = 告警主机名
	// appDomain 假如不为空，则以 appDomain 为告警域名为准,否则从 cmdb 获取

	config := g.Config()
	var pigeonInfo pigeon.Alarm

	pigeonInfo.Fid = 67300
	pigeonInfo.AlarmCode = "232-000"
	pigeonInfo.Value =  strconv.Itoa(g.FalconPingLru.GetLength(0))
	pigeonInfo.Priority = config.AgentPriority
	pigeonInfo.Domain = "falcon-pingcheck.vip.vip.com"
	pigeonInfo.Ip = g.GetIP()
	pigeonInfo.Metric = "falcon.pingcheck.degrade"
	pigeonInfo.Hostname, _ = os.Hostname()

	pigeonInfo.Subject = fmt.Sprintf("[FALCON PING CHECK 降级通知] IP: %s 最近 %d 分钟产生了 %d 个 " +
		"falcon-pingcheck 告警", pigeonInfo.Ip, config.Degrade.Period,
		g.FalconPingLru.GetLength(0))

	/// terry.zeng !!!!!
	pigeonInfo.Message = fmt.Sprintf("[FALCON PING CHECK 降级通知] 最近 %d 分钟产生了 %d 个" +
		"falcon-pingcheck 故障告警，最近 2 分钟的详细主机信息: %s", config.Degrade.Period,
		g.FalconPingLru.GetLength(2), g.FalconPingLru.GetLruDetail(2))

	if config.Debug == true {
		g.Alarmer.Debugf("[FALCON PING CHECK 降级通知] %s", g.FalconPingLru.GetLruDetail(0))
	}

	if config.AlarmEnable == true {
		_ = pigeon.SendPigeonAlarm(pigeonInfo)
	}

}

func SendRedisAlarm(alarmInfo string) {

	// 用于发送告警至 pigeon
	// string = alarm info message
	config := g.Config()
	var pigeonInfo pigeon.Alarm

	pigeonInfo.Fid = 71425
	pigeonInfo.AlarmCode = "232-000"
	pigeonInfo.Priority = config.AgentPriority
	pigeonInfo.Domain = "falcon-pingcheck.vip.vip.com"
	pigeonInfo.Ip = g.GetIP()
	pigeonInfo.Metric = "falcon.pingcheck.redis_connect"
	pigeonInfo.Hostname, _ = os.Hostname()
	pigeonInfo.Value =  "0"

	// g.Logger.Warningf("[REDIS ERROR] redis 连接错误，信息: %s", alarmInfo)

	pigeonInfo.Subject = fmt.Sprintf("[FALCON PINGCHECK 连接 redis 错误] 故障 IP : %s", pigeonInfo.Ip)
	pigeonInfo.Message = fmt.Sprintf("[FALCON PINGCHECK 连接 redis 错误] 故障信息: %s", alarmInfo)

	g.Alarmer.Debugf("[SendRedisAlarm 告警记录] %s", pigeonInfo.Subject)

	if config.AlarmEnable == true {
		_ = pigeon.SendPigeonAlarm(pigeonInfo)
	}

}


func SendPingDieHardAlarm(domain string, hostList []string) {

	var hostAllow []string

	for _, host := range hostList {
		if 	cmdb.CheckHostInList(host, cmdb.AllowCmdbHostInfo) == true {
			// skip alarm when host not in ping allow and falcon agent allow
			hostAllow = append(hostAllow, host)
		} else {
			g.Logger.Debugf("[PING DIE DEBUG] 当前主机 %s 因为 CMDB " +
				"设定原因，跳过 pingdie 告警", host)
		}
	}


	if len(hostAllow) == 0 {
		return
	}

	// 用于发送告警至 pigeon
	// string = alarm info message
	config := g.Config()
	var pigeonInfo pigeon.Alarm

	detail := strings.Split(hostList[0], "@")
	hostname := detail[0]
	ipaddress := detail[1]
	pigeonInfo.Fid = 71557
	pigeonInfo.AlarmCode = "232-000"
	pigeonInfo.Priority = config.AgentPriority
	pigeonInfo.Domain = domain
	pigeonInfo.Ip = ipaddress
	pigeonInfo.Metric = "falcon.pingcheck.pingdie"
	pigeonInfo.Hostname = hostname
	pigeonInfo.Value =  "0"

	g.Logger.Infof("[PING ALARMS] 应用 %s 无法 ping 服务器数量 %d, 包含: %s", domain, len(hostList), strings.Join(hostList[:],","))

	pigeonInfo.Subject = fmt.Sprintf("[PING] 应用 %s 发生 %d 台服务器不可 ping 故障", domain, len(hostList))
	pigeonInfo.Message = fmt.Sprintf("[PING] 应用 %s 发生 %d 台服务器不可 ping 故障，包含下面服务器: %s", domain, len(hostList), strings.Join(hostList[:],","))

	// use to put hostlist to  FalconPingLru

	for _, host := range hostList {
		hostDetail := domain + "@" + host
		g.FalconPingLru.Append(hostDetail)
	}

	if config.Debug {
		g.Alarmer.Debugf("[SendPingDieHardAlarm 告警记录] %s", pigeonInfo.Message)
	}

	if config.AgentCheck == false {
		return
	}

	if config.AlarmEnable == true  {
		g.FalconPingAlarmSuccess.Incr()
		_ = pigeon.SendPigeonAlarm(pigeonInfo)
	}

}


func sendMultiHostnameAlarm(hostname string, ipList []string) {

	// 用于发送告警至 pigeon
	// hostname = 告警主机名
	// ipList == multi ipaddr with same hostname

	config := g.Config()

	cmdbApi := "/server/query"

	for _, ipAddr := range ipList {

		cmdbQuery := "ip=" + ipAddr
		cmdbInfo, err := cmdb.CmdbApiQuery(cmdbApi, cmdbQuery)

		if err != nil {
			g.FalconCMDBHostNameAlarmFailure.Incr()
			g.Logger.Errorf("[sendMultiHostnameAlarm] CmdbApiQuery() get ip %s info error.", ipAddr)
			continue
		}

		var cmdbDomain string

		if len(cmdbInfo.Object) > 0 {
			cmdbDomain = cmdbInfo.Object[0].AppName
		}

		var pigeonInfo pigeon.Alarm
		pigeonInfo.Fid = 72559
		pigeonInfo.AlarmCode = "232-000"
		pigeonInfo.Value =  strconv.Itoa(g.FalconPingLru.GetLength(0))
		pigeonInfo.Priority = config.AgentPriority
		pigeonInfo.Domain = cmdbDomain
		pigeonInfo.Ip = ipAddr
		pigeonInfo.Metric = "falcon.hostname.duplicate"
		pigeonInfo.Hostname = hostname

		pigeonInfo.Subject = fmt.Sprintf("[CMDB 主机名重复告警] 主机名 %s IP: %s " +
			"发生了主机名重复问题", hostname, ipAddr)

		/// terry.zeng !!!!!
		pigeonInfo.Message = fmt.Sprintf("[CMDB 主机名重复告警] 主机名 %s IP: %s " +
			"发生了主机名重复问题, 与下面IP 地址主机名一致 %v", hostname, ipAddr, ipList)

		if config.Debug == true {
			g.Alarmer.Debugf("[sendMultiHostnameAlarm 告警记录] %s", pigeonInfo.Message )
		}

		if config.AlarmEnable == true  {
			g.FalconCMDBHostNameAlarmSuccess.Incr()
			_ = pigeon.SendPigeonAlarm(pigeonInfo)
		}
	}
}
