package http

type AgentAlarm struct {
	TimeStamp string	`json:"time"`
	HostList  []string	`json:"hostlist"`
}

