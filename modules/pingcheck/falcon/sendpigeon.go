package falcon

import (
	"fmt"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/cmdb"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/g"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/net"
	"github.com/open-falcon/falcon-plus/modules/pingcheck/pigeon"
	"os"
	"strconv"
)

func SendAlarm (host string, appDomain string) () {

	// 用于发送告警至 pigeon
	// host = 告警主机名
	// appDomain 假如不为空，则以 appDomain 为告警域名为准,否则从 cmdb 获取

	var falconIpaddr string
	var cmdbIpaddr string
	var cmdbDomain string
	var status int
	var pigeonInfo pigeon.Alarm

	falconIpaddr, err := GetFalconHost(host)

	if err != nil {
		g.Logger.Errorf("GetFalconHost() get host %s error", host)
		return
	}

	cmdbApi := "/server/query"
	cmdbQuery := "server_name=" + host
	cmdbInfo, err := cmdb.CmdbApiQuery(cmdbApi, cmdbQuery)

	if err != nil {
		g.Logger.Errorf("CmdbApiQuery() get host %s info error.", host)

		// do delete host from falcon api
		err := DeleteFlaonHost(host)
		if err != nil {
			g.Logger.Errorf("DeleteFlaonHost() host: %s error.", host)
			return
		}
		g.Logger.Infof("DeleteFlaonHost() host: %s success.", host)

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

		// if g.Config().Debug{
		//	g.Logger.Infof("SendAlarm() ip: %s, domain: %s, status: %d", cmdbIpaddr, cmdbDomain, status)
		// }

		if status != 4 {
			// 非生产状态主机，无需告警
			return
		}

	} else {
		g.Logger.Errorf("CmdbApiQuery() get host %s length is zero!!", host)
		return
	}

	if cmdbIpaddr != falconIpaddr && falconIpaddr != "" && cmdbIpaddr != "" {
		g.Logger.Errorf("falcon ip: %s, cmdb ip: %s",
			falconIpaddr, cmdbIpaddr)
		// do something
	}

	// pingStatus, err := net.CheckPing(cmdbIpaddr)
	// net.CheckPing  must running under by root

	pingStatus := net.PingFromProxy(cmdbIpaddr)

	pigeonInfo.Domain = cmdbDomain
	pigeonInfo.Ip = cmdbIpaddr
	pigeonInfo.Metric = "agent.alive"
	pigeonInfo.Hostname = host
	pigeonInfo.Value = "-1"
	pigeonInfo.Status = status

	if pingStatus == false {
		pigeonInfo.Metric = "agent.ping"
		g.Logger.Infof("[ALARM] PingStatus() ping %s false.", cmdbIpaddr)
		pigeonInfo.Event = fmt.Sprintf("[PING] 主机名: %s IP: %s ping critical", host, cmdbIpaddr)
		pigeonInfo.Detail = fmt.Sprintf("[PING] 主机名: %s IP: %s ping critical", host, cmdbIpaddr)
		pigeonInfo.Message = fmt.Sprintf("[PING] 主机名: %s IP: %s 无法被 ping 通，常见为物理机故障。", host, cmdbIpaddr)
		pigeonInfo.Priority = g.Config().AgentPriority
	} else {
		pigeonInfo.Metric = "agent.alive"
		g.Logger.Infof("[ALARM] PingStatus() falcon agent %s false.", cmdbIpaddr)
		pigeonInfo.Event = fmt.Sprintf("[FALCON] 主机名: %s IP: %s falcon agent 故障", host, cmdbIpaddr)
		pigeonInfo.Detail = fmt.Sprintf("[FALCON] 主机名: %s IP: %s falcon agent 故障", host, cmdbIpaddr)
		pigeonInfo.Message = fmt.Sprintf("[FALCON] 主机名: %s IP: %s falcon agent 故障, 参考 WIKI https://wiki.corp.vipshop.com/pages/viewpage.action?pageId=2236291330 进行 falcon-agent 故障排查", host, cmdbIpaddr)
		pigeonInfo.Priority = g.Config().AgentPriority
	}

	if g.Config().Debug {
		g.Logger.Debugf("[pingcheck 告警记录] 主机: %s, IP: %s, metrics: %s", host,
			cmdbIpaddr, pigeonInfo.Metric)
	}

	if g.Config().AlarmEnable {
		_ = pigeon.SendPigeonAlarm(pigeonInfo)
	}
}


func SendInternalAlarm(hostInfo []string) {

	// 用于发送告警至 pigeon
	// host = 告警主机名
	// appDomain 假如不为空，则以 appDomain 为告警域名为准,否则从 cmdb 获取


	var status int
	var pigeonInfo pigeon.Alarm

	pigeonInfo.Domain = "falcon-pingcheck.vip.vip.com"
	pigeonInfo.Ip = g.GetIP()
	pigeonInfo.Metric = "falcon.pingcheck.degrade"
	pigeonInfo.Hostname, _ = os.Hostname()
	pigeonInfo.Value =  strconv.Itoa(len(hostInfo))
	pigeonInfo.Status = status


	g.Logger.Infof("[INTERNAL] falcon pingcheck 降级， 包含下面主机信息: %v", hostInfo)

	pigeonInfo.Event = fmt.Sprintf("[falcon pingcheck 降级通知] 最近 %d 分钟产生了 %d 个 " +
		"falcon-pingcheck 告警", (g.Config().Degrade.Period + 1) , len(hostInfo))

	pigeonInfo.Detail = fmt.Sprintf("[falcon pingcheck 降级通知] 最近 %d 分钟产生了 %d " +
		"个 falcon-pingcheck 故障告警，详细主机信息包含了 %v", (g.Config().Degrade.Period + 1),
		len(hostInfo), hostInfo)

	pigeonInfo.Message = fmt.Sprintf("[falcon pingcheck 降级通知] 最近 %d 分钟产生了 %d 个" +
		" falcon-pingcheck 故障告警，详细主机信息包含了 %v", (g.Config().Degrade.Period + 1),
		len(hostInfo), hostInfo)

	pigeonInfo.Priority = g.Config().AgentPriority

	if g.Config().Debug {
		g.Logger.Debugf("[告警降级记录] %s", pigeonInfo.Message)
	}

	if g.Config().AlarmEnable {
		_ = pigeon.SendPigeonAlarm(pigeonInfo)
	} 

}


func SendRedisAlarm(alarmInfo string) {

	// 用于发送告警至 pigeon
	// string = alarm info message

	var status int
	var pigeonInfo pigeon.Alarm

	pigeonInfo.Domain = "falcon-pingcheck.vip.vip.com"
	pigeonInfo.Ip = g.GetIP()
	pigeonInfo.Metric = "falcon.pingcheck.redis_connect"
	pigeonInfo.Hostname, _ = os.Hostname()
	pigeonInfo.Value =  "0"
	pigeonInfo.Status = status


	g.Logger.Warningf("[INTERNAL] redis 连接错误，信息: %s", alarmInfo)

	pigeonInfo.Event = fmt.Sprintf("[falcon pingcheck 连接 redis 错误] 故障信息: %s", alarmInfo)

	pigeonInfo.Detail = fmt.Sprintf("[falcon pingcheck 连接 redis 错误] 故障信息: %s", alarmInfo)

	pigeonInfo.Message = fmt.Sprintf("[falcon pingcheck 连接 redis 错误] 故障信息: %s", alarmInfo)

	pigeonInfo.Priority = g.Config().AgentPriority


	if g.Config().AlarmEnable {
		_ = pigeon.SendPigeonAlarm(pigeonInfo)
	}

}
