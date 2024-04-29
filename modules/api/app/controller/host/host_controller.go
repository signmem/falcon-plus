package host

import (
	"fmt"
	"github.com/signmem/falcon-plus/modules/api/config"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	cmodel "github.com/signmem/falcon-plus/common/model"
	h "github.com/signmem/falcon-plus/modules/api/app/helper"
	alm "github.com/signmem/falcon-plus/modules/api/app/model/alarm"
	f "github.com/signmem/falcon-plus/modules/api/app/model/falcon_portal"
	m "github.com/signmem/falcon-plus/modules/api/app/model/graph"
	u "github.com/signmem/falcon-plus/modules/api/app/utils"
	grh "github.com/signmem/falcon-plus/modules/api/graph"
)

type APIMaintainInputs struct {
	HostNames []string `json:"host_names"`
	Duration  int64    `json:"duration"`
}

func GetHostBindToWhichHostGroup(c *gin.Context) {
	HostIdTmp := c.Params.ByName("host_id")
	if HostIdTmp == "" {
		h.JSONR(c, badstatus, "host id is missing")
		return
	}
	//hostID, err := strconv.Atoi(HostIdTmp)
	hostID, err := strconv.ParseInt(HostIdTmp, 10, 64)
	if err != nil {
		config.Logger.Debugf("HostId: %v", HostIdTmp)
		h.JSONR(c, badstatus, err)
		return
	}
	grpHostMap := []f.GrpHost{}
	db.Falcon.Select("grp_id").Where("host_id = ?", hostID).Find(&grpHostMap)
	grpIds := []int64{}
	for _, g := range grpHostMap {
		grpIds = append(grpIds, g.GrpID)
	}
	hostgroups := []f.HostGroup{}
	if len(grpIds) != 0 {
		grpIdsStr, _ := u.ArrInt64ToString(grpIds)
		db.Falcon.Where(fmt.Sprintf("id in (%s)", grpIdsStr)).Find(&hostgroups)
	}
	h.JSONR(c, hostgroups)
	return
}

func GetHostGroupWithTemplate(c *gin.Context) {
	grpIDtmp := c.Params.ByName("host_group")
	if grpIDtmp == "" {
		h.JSONR(c, badstatus, "grp id is missing")
		return
	}
	grpID, err := strconv.Atoi(grpIDtmp)
	if err != nil {
		config.Logger.Debugf("grpIDtmp: %v", grpIDtmp)
		h.JSONR(c, badstatus, err)
		return
	}
	hostgroup := f.HostGroup{ID: int64(grpID)}
	if dt := db.Falcon.Find(&hostgroup); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	hosts := []f.Host{}
	grpHosts := []f.GrpHost{}
	if dt := db.Falcon.Where("grp_id = ?", grpID).Find(&grpHosts); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	for _, grph := range grpHosts {
		var host f.Host
		db.Falcon.Find(&host, grph.HostID)
		if host.ID != 0 {
			hosts = append(hosts, host)
		}
	}
	h.JSONR(c, map[string]interface{}{
		"hostgroup": hostgroup,
		"hosts":     hosts,
	})
	return
}

func GetGrpsRelatedHost(c *gin.Context) {
	hostIDtmp := c.Params.ByName("host_id")
	if hostIDtmp == "" {
		h.JSONR(c, badstatus, "host id is missing")
		return
	}
	//hostID, err := strconv.Atoi(hostIDtmp)
	hostID, err := strconv.ParseInt(hostIDtmp, 10, 64)
	if err != nil {
		config.Logger.Debugf("host id: %v", hostIDtmp)
		h.JSONR(c, badstatus, err)
		return
	}

	//host := f.Host{ID: int64(hostID)}
	host := f.Host{ID: hostID}
	if dt := db.Falcon.Find(&host); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	grps := host.RelatedGrp()
	h.JSONR(c, grps)
	return
}

func GetHostIP(c *gin.Context) {
	hostname := c.Params.ByName("hostname")
	if hostname == "" {
		h.JSONR(c, badstatus, "host name is missing")
		return
	}

	host := f.Host{}
	if dt := db.Falcon.Where("hostname = ?", hostname).Find(&host); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("Host %v does not exist!", hostname))
		return
	}

	hostID := strconv.FormatInt(host.ID,10)
	h.JSONR(c, map[string]string{
		"hostname": host.Hostname,
		"ip":       host.Ip,
		"id":       hostID,
	})
	return
}

func GetIPList(c *gin.Context) {
	Hosts := []f.Host{}
	if dt := db.Falcon.Table("host").Select("ip").Find(&Hosts); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	ips := []string{}
	if len(Hosts) != 0 {
		for _, h := range Hosts {
			if len(h.Ip) != 0 {
				ips = append(ips, h.Ip)
			}
		}
	}

	h.JSONR(c, ips)
	return
}

func GetHostList(c *gin.Context) {
	Hosts := []f.Host{}
	if dt := db.Falcon.Table("host").Select("hostname").Find(&Hosts); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	hosts := []string{}
	if len(Hosts) != 0 {
		for _, h := range Hosts {
			hosts = append(hosts, h.Hostname)
		}
	}

	h.JSONR(c, hosts)
	return
}

func GetHosts(c *gin.Context) {
	Hosts := []f.Host{}
	if dt := db.Falcon.Table("host").Find(&Hosts); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	/*
		hosts := []string{}
		if len(Hosts) != 0 {
			for _, h := range Hosts {
				hosts = append(hosts, h.Hostname)
			}
		}
	*/
	h.JSONR(c, Hosts)
	return
}

func GetTplsRelatedHost(c *gin.Context) {
	hostIDtmp := c.Params.ByName("host_id")
	if hostIDtmp == "" {
		h.JSONR(c, badstatus, "host id is missing")
		return
	}
	hostID, err := strconv.Atoi(hostIDtmp)
	if err != nil {
		config.Logger.Debugf("host id: %v", hostIDtmp)
		h.JSONR(c, badstatus, err)
		return
	}
	host := f.Host{ID: int64(hostID)}
	if dt := db.Falcon.Find(&host); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	tpls := host.RelatedTpl()
	h.JSONR(c, tpls)
	return
}

func SetMaintain(c *gin.Context) {
	var inputs APIMaintainInputs
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	if inputs.HostNames == nil {
		h.JSONR(c, badstatus, "host name is missing")
		return
	}

	if inputs.Duration == 0 {
		inputs.Duration = 60
	}

	Hosts := []string{}
	for _, v := range inputs.HostNames {
		host := f.Host{}
		dt := db.Falcon.Where("hostname = ?", v).Find(&host)
		if dt.Error != nil {
			h.JSONR(c, expecstatus, fmt.Sprintf("Host %v not found!", v))
		} else {
			Hosts = append(Hosts, v)
		}
	}

	if len(Hosts) == 0 {
		h.JSONR(c, badstatus, "None of hosts found!")
		return
	}

	maintainBegin := time.Now().Unix()
	//maintainEnd := time.Now().Unix() + 315360000 //10 years later
	maintainEnd := time.Now().Unix() + inputs.Duration*60

	dt := db.Falcon.Table("host").Where("hostname in (?)",
		Hosts).Updates(map[string]interface{}{"maintain_begin": maintainBegin, "maintain_end": maintainEnd})
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	h.JSONR(c, fmt.Sprintf("Set the hosts %v to maintain OK!", Hosts))
}

func UnsetMaintain(c *gin.Context) {
	var inputs APIMaintainInputs
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	if inputs.HostNames == nil {
		h.JSONR(c, badstatus, "host name is missing")
		return
	}

	Hosts := []string{}
	for _, v := range inputs.HostNames {
		host := f.Host{}
		dt := db.Falcon.Where("hostname = ?", v).Find(&host)
		if dt.Error != nil {
			h.JSONR(c, expecstatus, fmt.Sprintf("Host %v not found!", v))
		} else {
			Hosts = append(Hosts, v)
		}
	}

	if len(Hosts) == 0 {
		h.JSONR(c, badstatus, "None of hosts found!")
		return
	}

	maintainBegin := 0
	maintainEnd := 0

	dt := db.Falcon.Table("host").Where("hostname in (?)",
		Hosts).Updates(map[string]interface{}{"maintain_begin": maintainBegin, "maintain_end": maintainEnd})
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	h.JSONR(c, fmt.Sprintf("Unset maintain for the hosts %v OK!", Hosts))
}

func delGraphEndpoint(inputs []string) {
	type DBRows struct {
		Endpoint  string
		CounterId int
		Counter   string
		Type      string
		Step      int
	}
	rows := []DBRows{}
	dt := db.Graph.Raw(
		`select a.endpoint, b.id AS counter_id, b.counter, b.type, b.step from endpoint as a, endpoint_counter as b
		where b.endpoint_id = a.id
		AND a.endpoint in (?)`, inputs).Scan(&rows)
	if dt.Error != nil {
		fmt.Println(dt.Error)
		return
	}

	var affected_counter int64 = 0
	var affected_endpoint int64 = 0

	if len(rows) > 0 {
		var params []*cmodel.GraphDeleteParam = []*cmodel.GraphDeleteParam{}
		for _, row := range rows {
			param := &cmodel.GraphDeleteParam{
				Endpoint: row.Endpoint,
				DsType:   row.Type,
				Step:     row.Step,
			}
			fields := strings.SplitN(row.Counter, "/", 2)
			if len(fields) == 1 {
				param.Metric = fields[0]
			} else if len(fields) == 2 {
				param.Metric = fields[0]
				param.Tags = fields[1]
			} else {
				config.Logger.Errorf("invalid counter", row.Counter)
				continue
			}
			params = append(params, param)
		}
		grh.Delete(params)
	}

	tx := db.Graph.Begin()

	if len(rows) > 0 {
		var cids []int = make([]int, len(rows))
		for i, row := range rows {
			cids[i] = row.CounterId
		}

		dt = tx.Table("endpoint_counter").Where("id in (?)", cids).Delete(&m.EndpointCounter{})
		if dt.Error != nil {
			fmt.Println(dt.Error)
			tx.Rollback()
			return
		}
		affected_counter = dt.RowsAffected

		dt = tx.Raw(`delete from tag_endpoint where endpoint_id in 
			(select id from endpoint where endpoint in (?))`, inputs)
		if dt.Error != nil {
			fmt.Println(dt.Error)
			tx.Rollback()
			return
		}
	}

	dt = tx.Table("endpoint").Where("endpoint in (?)", inputs).Delete(&m.Endpoint{})
	if dt.Error != nil {
		fmt.Println(dt.Error)
		tx.Rollback()
		return
	}
	affected_endpoint = dt.RowsAffected
	tx.Commit()

	ret := map[string]int64{
		"affected_endpoint": affected_endpoint,
		"affected_counter":  affected_counter,
	}

	fmt.Printf("%#v\n", ret)
}

//delete endpoint alarm notes & event cases
func delEndpointAlarm(endpoints []string) {
	//for get EventCases table name
	ec := alm.EventCases{}
	//for get EventNote table name
	//en := alm.EventNote{}

	if dt := db.Alarm.Where("endpoint in (?)", endpoints).Delete(&ec); dt.Error != nil {
		config.Logger.Errorf("Error occurred while deleting event cases for endpoint %v", endpoints)
		return
	}

	fmt.Printf("eventcases of endpoints %v have been deleted", endpoints)
}

//delete host & related grp_host, graphs, alarms
func DeleteHost(c *gin.Context) {
	hostname := c.Params.ByName("hostname")
	if hostname == "" {
		h.JSONR(c, badstatus, "host name is missing")
		return
	}

	host := f.Host{}
	if dt := db.Falcon.Where("hostname = ?", hostname).Find(&host); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("Host %v does not exist!", hostname))
		return
	}

	tx := db.Falcon.Begin()
	// UnBind Host To HostGroup
	if dt := tx.Where("host_id = ?", host.ID).Delete(&f.GrpHost{}); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete grp_host got error: %v", dt.Error))
		dt.Rollback()
		return
	}

	// Delete Host
	if dt := tx.Where("hostname = ?", hostname).Delete(&host); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete host got error: %v", dt.Error))
		dt.Rollback()
		return
	}
	tx.Commit()

	// Delete Endpoint Alarms
	delEndpointAlarm([]string{hostname})

	// Delete Endpoint Graph
	delGraphEndpoint([]string{hostname})

	h.JSONR(c, fmt.Sprintf("host:%v and related grp_host, graphs, alarms have been deleted", hostname))
	return
}

type APIDeleteHostsInputs struct {
	HostNames []string `json:"host_names"`
}

func DeleteHosts(c *gin.Context) {
	var inputs APIDeleteHostsInputs
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	if inputs.HostNames == nil {
		h.JSONR(c, badstatus, "host name is missing")
		return
	}

	Hosts := []string{}
	HostIDs := []int64{}
	for _, v := range inputs.HostNames {
		host := f.Host{}
		dt := db.Falcon.Where("hostname = ?", v).Find(&host)
		if dt.Error != nil {
			h.JSONR(c, expecstatus, fmt.Sprintf("Host %v not found!", v))
		} else {
			Hosts = append(Hosts, v)
			HostIDs = append(HostIDs, host.ID)
		}
	}

	/*
		host := f.Host{}
		if dt := db.Falcon.Where("hostname = ?", hostname).Find(&host); dt.Error != nil {
			h.JSONR(c, expecstatus, fmt.Sprintf("Host %v does not exist!", hostname))
			return
		}
	*/

	tx := db.Falcon.Begin()
	// UnBind Host To HostGroup
	if dt := tx.Where("host_id in (?)", HostIDs).Delete(&f.GrpHost{}); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete grp_host got error: %v", dt.Error))
		dt.Rollback()
		return
	}

	// Delete Host
	if dt := tx.Where("hostname in (?)", Hosts).Delete(&f.Host{}); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete hosts got error: %v", dt.Error))
		dt.Rollback()
		return
	}
	tx.Commit()

	// Delete Endpoint Alarms
	delEndpointAlarm(Hosts)

	// Delete Endpoint Graph
	delGraphEndpoint(Hosts)

	h.JSONR(c, fmt.Sprintf("hosts:%v and related grp_host, graphs, alarms have been deleted", Hosts))
	return
}

type HostBindHostGroupStruct struct {
	HostID 		int		`json:"host_id"`
	GrpID 		int		`json:"grp_id"`
}

func CheckHostBindHostGroup(c *gin.Context) {
	// use to check if host bind hostgroup already

	var inputs HostBindHostGroupStruct
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	grpHost := f.GrpHost{}

	if dt := db.Falcon.Where("grp_id = ? AND host_id = ?", int64(inputs.GrpID), int64(inputs.HostID)).Find(&grpHost); dt.Error != nil {
		h.JSONR(c, fmt.Sprintf("false"))
		return
	}

	h.JSONR(c, fmt.Sprintf("true"))
	return
}