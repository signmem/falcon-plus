package db

import (
	"fmt"
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/hbs/g"
	"log"
)

func UpdateAgent(agentInfo *model.AgentUpdateInfo) {
	var (
		hostname       string
		ip             string
		agent_version  string
		plugin_version string
	)

	sql := fmt.Sprintf(
		"select hostname, ip, agent_version, plugin_version from host where hostname = '%s'",
		agentInfo.ReportRequest.Hostname,
	)

	rows, err := DB.Query(sql)
	defer rows.Close()
	if err == nil {
		for rows.Next() {
			err = rows.Scan(&hostname, &ip, &agent_version, &plugin_version)
			if err != nil {
				log.Println("ERROR:", err)
				continue
			}
		}
	} else {
		log.Printf("select ERROR:%s, hostname:%s", err.Error(), agentInfo.ReportRequest.Hostname)
	}

	if agentInfo.ReportRequest.Hostname == hostname && agentInfo.ReportRequest.IP == ip && agentInfo.ReportRequest.AgentVersion == agent_version && agentInfo.ReportRequest.PluginVersion == plugin_version {
		return
	}

	sql = ""
	if g.Config().Hosts == "" {
		if hostname == "" && ip == "" && agent_version == "" && plugin_version == "" {
			sql = fmt.Sprintf(
				"insert into host(hostname, ip, agent_version, plugin_version) values ('%s', '%s', '%s', '%s')",
				agentInfo.ReportRequest.Hostname,
				agentInfo.ReportRequest.IP,
				agentInfo.ReportRequest.AgentVersion,
				agentInfo.ReportRequest.PluginVersion,
			)
		} else {
			sql = fmt.Sprintf(
				"update host set ip='%s', agent_version='%s', plugin_version='%s' where hostname='%s'",
				agentInfo.ReportRequest.IP,
				agentInfo.ReportRequest.AgentVersion,
				agentInfo.ReportRequest.PluginVersion,
				agentInfo.ReportRequest.Hostname,
			)
		}
	} else {
		// sync, just update
		sql = fmt.Sprintf(
			"update host set ip='%s', agent_version='%s', plugin_version='%s' where hostname='%s'",
			agentInfo.ReportRequest.IP,
			agentInfo.ReportRequest.AgentVersion,
			agentInfo.ReportRequest.PluginVersion,
			agentInfo.ReportRequest.Hostname,
		)
	}

	_, err = DB.Exec(sql)
	if err != nil {
		log.Println("exec", sql, "fail", err)
	}

}
