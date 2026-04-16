package g

const (
	Version = "3.1.2"
)

var (
	TransferCheck  = false
	RedisNormalHost []string
	SkipAgentAlarm = false
	SkipPingAlarm = false
	FalconAgentLru Lru
	FalconPingLru Lru

	FalconAgentAliveReportSuccess = NewSCount("falconagent.alive.report.success")
	FalconAgentAliveReportTimeOut = NewSCount("falconagent.alive.report.timeout")

	FalconAgentAliveAlarmSuccess = NewSCount("falconagent.alive.alarm.success")
	FalconAgentAliveAlarmFailure = NewSCount("falconagent.alive.alarm.failure")

	FalconAgentAliveAlarmCounter = NewSCount("falconagent.alive.alarm.counter")
	FalconAgentAliveAlarmDegarded = NewSCount("falconagent.alive.alarm.degarded")


	FalconAgentCMDBReportFailure = NewSCount("falconagent.cmdb.report.failure")

	FalconCMDBHostNameReplicate = NewSCount("falconagent.cmdb.hostname.replicate")

	FalconCMDBHostNameAlarmFailure  = NewSCount("falconagent.cmdb.hostname.alarm.failure")
	FalconCMDBHostNameAlarmSuccess = NewSCount("falconagent.cmdb.hostname.alarm.success")

	FalconPingTotal   = NewSCount("falconagent.ping.total")
	FalconPingSuccess = NewSCount("falconagent.ping.success")
	FalconPingFailuer = NewSCount("falconagent.ping.failure")
	FalconPingDie     = NewSCount("falconagent.ping.die")

	FalconPingAlarmSuccess = NewSCount("falconagent.ping.alarm.success")

	FalconPingCheckTotal    = NewSCount("falcon.ping.check.total")
	FalconPingCheckSuccess  = NewSCount("falcon.ping.check.success")
	FalconPingCheckFailuer  = NewSCount("falcon.ping.check.failure")

)

// falcon.GetRedisHostsExpire  maintain  RedisNormalHost
// FalconAgentLru  -> /falcon.agent  Lru
// FalconPingLru   -> /falcon.pingdead Lru
