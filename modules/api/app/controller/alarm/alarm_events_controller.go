package alarm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	h "github.com/open-falcon/falcon-plus/modules/api/app/helper"
	alm "github.com/open-falcon/falcon-plus/modules/api/app/model/alarm"
)

type APIGetAlarmListsInputs struct {
	StartTime     int64  `json:"startTime" form:"startTime"`
	EndTime       int64  `json:"endTime" form:"endTime"`
	Priority      int    `json:"priority" form:"priority"`
	Status        string `json:"status" form:"status"`
	ProcessStatus string `json:"process_status" form:"process_status"`
	Metrics       string `json:"metrics" form:"metrics"`
	//id
	EventId string `json:"event_id" form:"event_id"`
	//number of reacord's limit on each page
	Limit int `json:"limit" form:"limit"`
	//pagging
	Page int `json:"page"`
}

func (input APIGetAlarmListsInputs) checkInputsContain() error {
	if input.StartTime == 0 && input.EndTime == 0 {
		if input.EventId == "" {
			return errors.New("startTime, endTime OR event_id, You have to at least pick one on the request.")
		}
	}
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (s APIGetAlarmListsInputs) collectFilters() string {
	tmp := []string{}
	if s.StartTime != 0 {
		tmp = append(tmp, fmt.Sprintf("timestamp >= FROM_UNIXTIME(%v)", s.StartTime))
	}
	if s.EndTime != 0 {
		tmp = append(tmp, fmt.Sprintf("timestamp <= FROM_UNIXTIME(%v)", s.EndTime))
	}
	if s.Priority != -1 {
		tmp = append(tmp, fmt.Sprintf("priority = %d", s.Priority))
	}
	if s.Status != "" {
		status := ""
		statusTmp := strings.Split(s.Status, ",")
		for indx, n := range statusTmp {
			if indx == 0 {
				status = fmt.Sprintf(" status = '%s' ", n)
			} else {
				status = fmt.Sprintf(" %s OR status = '%s' ", status, n)
			}
		}
		status = fmt.Sprintf("( %s )", status)
		tmp = append(tmp, status)
	}
	if s.ProcessStatus != "" {
		pstatus := ""
		pstatusTmp := strings.Split(s.ProcessStatus, ",")
		for indx, n := range pstatusTmp {
			if indx == 0 {
				pstatus = fmt.Sprintf(" process_status = '%s' ", n)
			} else {
				pstatus = fmt.Sprintf(" %s OR process_status = '%s' ", pstatus, n)
			}
		}
		pstatus = fmt.Sprintf("( %s )", pstatus)
		tmp = append(tmp, pstatus)
	}
	if s.Metrics != "" {
		tmp = append(tmp, fmt.Sprintf("metrics regexp '%s'", s.Metrics))
	}
	if s.EventId != "" {
		tmp = append(tmp, fmt.Sprintf("id = '%s'", s.EventId))
	}
	filterStrTmp := strings.Join(tmp, " AND ")
	if filterStrTmp != "" {
		filterStrTmp = fmt.Sprintf("WHERE %s", filterStrTmp)
	}
	return filterStrTmp
}

func AlarmLists(c *gin.Context) {
	var inputs APIGetAlarmListsInputs
	//set default
	inputs.Page = -1
	inputs.Priority = -1
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	if err := inputs.checkInputsContain(); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	filterCollector := inputs.collectFilters()
	//for get correct table name
	f := alm.EventCases{}
	cevens := []alm.EventCases{}
	perparedSql := ""
	//if no specific, will give return first 2000 records
	if inputs.Page == -1 {
		if inputs.Limit >= 2000 || inputs.Limit == 0 {
			inputs.Limit = 2000
		}
		perparedSql = fmt.Sprintf("select * from %s %s order by timestamp DESC limit %d", f.TableName(), filterCollector, inputs.Limit)
	} else {
		//set the max limit of each page
		if inputs.Limit >= 50 {
			inputs.Limit = 50
		}
		perparedSql = fmt.Sprintf("select * from %s %s  order by timestamp DESC limit %d,%d", f.TableName(), filterCollector, inputs.Page, inputs.Limit)
	}
	db.Alarm.Raw(perparedSql).Find(&cevens)
	h.JSONR(c, cevens)
}

type APIEventsGetInputs struct {
	StartTime int64 `json:"startTime" form:"startTime"`
	EndTime   int64 `json:"endTime" form:"endTime"`
	Status    int   `json:"status" form:"status" binding:"gte=-1,lte=1"`
	//event_caseId
	EventId string `json:"event_id" form:"event_id" binding:"required"`
	//number of reacord's limit on each page
	Limit int `json:"limit" form:"limit"`
	//pagging
	Page int `json:"page" form:"page"`
}

func (s APIEventsGetInputs) collectFilters() string {
	tmp := []string{}
	filterStrTmp := ""
	if s.StartTime != 0 {
		tmp = append(tmp, fmt.Sprintf("timestamp >= FROM_UNIXTIME(%v)", s.StartTime))
	}
	if s.EndTime != 0 {
		tmp = append(tmp, fmt.Sprintf("timestamp <= FROM_UNIXTIME(%v)", s.EndTime))
	}
	if s.EventId != "" {
		tmp = append(tmp, fmt.Sprintf("event_caseId = '%s'", s.EventId))
	}
	if s.Status == 0 || s.Status == 1 {
		tmp = append(tmp, fmt.Sprintf("status = %d", s.Status))
	}
	if len(tmp) != 0 {
		filterStrTmp = strings.Join(tmp, " AND ")
		filterStrTmp = fmt.Sprintf("WHERE %s", filterStrTmp)
	}
	return filterStrTmp
}

func EventsGet(c *gin.Context) {
	var inputs APIEventsGetInputs
	inputs.Status = -1
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	filterCollector := inputs.collectFilters()
	//for get correct table name
	f := alm.Events{}
	evens := []alm.Events{}
	if inputs.Limit == 0 || inputs.Limit >= 50 {
		inputs.Limit = 50
	}
	perparedSql := fmt.Sprintf("select id, event_caseId, cond, status, timestamp from %s %s order by timestamp DESC limit %d,%d", f.TableName(), filterCollector, inputs.Page, inputs.Limit)
	db.Alarm.Raw(perparedSql).Scan(&evens)
	h.JSONR(c, evens)
}

type APIEventCasesGetInputs struct {
	StartTime     string `json:"startTime" form:"startTime"`
	EndTime       string `json:"endTime" form:"endTime"`
	Priority      string `json:"priority" form:"priority"`
	Status        string `json:"status" form:"status"`
	ProcessStatus string `json:"process_status" form:"process_status"`
	Metric        string `json:"metric" form:"metric"`

	BussName string `json:"buss_name" form:"buss_name"`
	GrpName  string `json:"grp_name" form:"grp_name"`
	Endpoint string `json:"endpoint" form:"endpoint"`
	IP       string `json:"ip" form:"ip"`
	//number of reacord's limit on each page
	//Limit int `json:"limit" form:"limit"`
}

func GetEventCases(c *gin.Context) {
	var inputs APIEventCasesGetInputs
	inputs.BussName = c.DefaultQuery("buss_name", "")
	inputs.GrpName = c.DefaultQuery("grp_name", "")
	inputs.Endpoint = c.DefaultQuery("endpoint", "")
	inputs.IP = c.DefaultQuery("ip", "")
	inputs.Status = c.DefaultQuery("status", "")
	inputs.ProcessStatus = c.DefaultQuery("process_status", "unresolved")
	inputs.Priority = c.DefaultQuery("priority", "-1")
	inputs.StartTime = c.DefaultQuery("startTime", strconv.FormatInt(time.Now().Unix()-3600, 10))
	inputs.EndTime = c.DefaultQuery("endTime", strconv.FormatInt(time.Now().Unix(), 10))
	inputs.Metric = c.DefaultQuery("metric", "all")

	//for get correct table name
	//f := alm.EventCases{}
	cevens := []alm.EventCases{}

	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	validProcessStatus := []string{"ignored", "unresolved", "resolved", `in progress`, "comment"}
	if !stringInSlice(inputs.ProcessStatus, validProcessStatus) {
		h.JSONR(c, badstatus, fmt.Sprintf(`The valid value of ProcessStatus is one of %v`, validProcessStatus))
		return
	}

	filterBussName := func(d *gorm.DB) *gorm.DB {
		if inputs.BussName != "" {
			return d.Where("buss_name = ?", inputs.BussName)
		} else {
			return d
		}
	}

	filterGroup := func(d *gorm.DB) *gorm.DB {
		if inputs.GrpName != "" {
			s_group := strings.SplitN(inputs.GrpName, ",", -1)
			return d.Where("grp_name in (?)", s_group)
		} else {
			return d
		}
	}

	filterEndpoint := func(d *gorm.DB) *gorm.DB {
		if inputs.Endpoint != "" {
			s_endpoint := strings.SplitN(inputs.Endpoint, ",", -1)
			return d.Where("endpoint in (?)", s_endpoint)
		} else {
			return d
		}
	}

	filterIP := func(d *gorm.DB) *gorm.DB {
		if inputs.IP != "" {
			s_ip := strings.SplitN(inputs.IP, ",", -1)
			return d.Where("ip in (?)", s_ip)
		} else {
			return d
		}
	}

	filterStatus := func(d *gorm.DB) *gorm.DB {
		if inputs.Status != "" {
			return d.Where("status = ?", inputs.Status)
		} else {
			return d
		}
	}

	filterMetric := func(d *gorm.DB) *gorm.DB {
		if inputs.Metric == "hardware" {
			return d.Where("metric like 'hardware.%'")
		} else if inputs.Metric == "non-hardware" {
			return d.Where("metric not like 'hardware.%'")
		} else if inputs.Metric == "all" {
			return d
		} else {
			if inputs.Metric != "" {
				m_value := inputs.Metric + "%"
				return d.Where("metric like ?", m_value)
			} else {
				return d
			}
		}
	}

	filterProcessStatus := func(d *gorm.DB) *gorm.DB {
		if inputs.ProcessStatus != "" {
			return d.Where("process_status = ?", inputs.ProcessStatus)
		} else {
			return d
		}
	}

	filterPriority := func(d *gorm.DB) *gorm.DB {
		if inputs.Priority != "-1" {
			return d.Where("priority = ?", inputs.Priority)
		} else {
			return d
		}
	}

	filterTime := func(d *gorm.DB) *gorm.DB {
		cond := fmt.Sprintf(`update_at >= FROM_UNIXTIME(%v) AND
		 update_at <= FROM_UNIXTIME(%v)`, inputs.StartTime, inputs.EndTime)
		return d.Where(cond)
	}
	db.Alarm.Scopes(filterBussName, filterGroup, filterEndpoint, filterIP, filterStatus, filterMetric, filterProcessStatus, filterPriority, filterTime).Find(&cevens)
	h.JSONR(c, cevens)
}

// Add FId
type EventCasesWFid struct {
	ID            string     `json:"id" gorm:"column:id"`
	HostId        int64      `json:"host_id" gorm:"column:host_id"`
	Endpoint      string     `json:"endpoint" grom:"column:endpoint"`
	GrpId         int64      `json:"grp_id" gorm:"column:grp_id"`
	Grp_name      string     `json:"grp_name" grom:"column:grp_name"`
	Metric        string     `json:"metric" grom:"metric"`
	Func          string     `json:"func" grom:"func"`
	Cond          string     `json:"cond" grom:"cond"`
	Note          string     `json:"note" grom:"note"`
	MaxStep       int        `json:"step" grom:"step"`
	CurrentStep   int        `json:"current_step" grom:"current_step"`
	Priority      int        `json:"priority" grom:"priority"`
	Status        string     `json:"status" grom:"status"`
	Timestamp     *time.Time `json:"timestamp" grom:"timestamp"`
	UpdateAt      *time.Time `json:"update_at" grom:"update_at"`
	ClosedAt      *time.Time `json:"closed_at" grom:"closed_at"`
	ClosedNote    string     `json:"closed_note" grom:"closed_note"`
	UserModified  int64      `json:"user_modified" grom:"user_modified"`
	TplCreator    string     `json:"tpl_creator" grom:"tpl_creator"`
	ExpressionId  int64      `json:"expression_id" grom:"expression_id"`
	StrategyId    int64      `json:"strategy_id" grom:"strategy_id"`
	TemplateId    int64      `json:"template_id" grom:"template_id"`
	ProcessNote   int64      `json:"process_note" grom:"process_note"`
	ProcessStatus string     `json:"process_status" grom:"process_status"`
	FId           int64      `json:"fid" gorm:"column:fid"`
}

// GetEventCases Version 2 which including fid returning, Add by campbell.tang @2018.9.8
func GetEventCasesV2(c *gin.Context) {
	var inputs APIEventCasesGetInputs
	inputs.GrpName = c.DefaultQuery("grp_name", "")
	inputs.Status = c.DefaultQuery("status", "")
	inputs.ProcessStatus = c.DefaultQuery("process_status", "unresolved")
	inputs.Metric = c.DefaultQuery("metric", "all")
	inputs.Priority = c.DefaultQuery("priority", "-1")
	inputs.StartTime = c.DefaultQuery("startTime", strconv.FormatInt(time.Now().Unix()-3600, 10))
	inputs.EndTime = c.DefaultQuery("endTime", strconv.FormatInt(time.Now().Unix(), 10))

	//for get correct table name
	//f := alm.EventCases{}
	cevens := []EventCasesWFid{}

	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	validProcessStatus := []string{"ignored", "unresolved", "resolved", `in progress`, "comment"}
	if !stringInSlice(inputs.ProcessStatus, validProcessStatus) {
		h.JSONR(c, badstatus, fmt.Sprintf(`The valid value of ProcessStatus is one of %v`, validProcessStatus))
		return
	}

	filterGroup := func(d *gorm.DB) *gorm.DB {
		if inputs.GrpName != "" {
			s_group := strings.SplitN(inputs.GrpName, ",", -1)
			return d.Where("grp_name in (?)", s_group)
		} else {
			return d
		}
	}

	filterStatus := func(d *gorm.DB) *gorm.DB {
		if inputs.Status != "" {
			return d.Where("status = ?", inputs.Status)
		} else {
			return d
		}
	}

	filterMetric := func(d *gorm.DB) *gorm.DB {
		if inputs.Metric == "hardware" {
			return d.Where("metric like 'hardware.%'")
		} else if inputs.Metric == "non-hardware" {
			return d.Where("metric not like 'hardware.%'")
		} else {
			return d
		}
	}

	filterProcessStatus := func(d *gorm.DB) *gorm.DB {
		if inputs.ProcessStatus != "" {
			return d.Where("process_status = ?", inputs.ProcessStatus)
		} else {
			return d
		}
	}

	// filterPriority := func(d *gorm.DB) *gorm.DB {
	// 	if inputs.Priority != "-1" {
	// 		return d.Where("priority = ?", inputs.Priority)
	// 	} else {
	// 		return d
	// 	}
	// }

	filterPriority := func(d *gorm.DB) *gorm.DB {
		if inputs.Priority != "-1" {
			s_priority := strings.SplitN(inputs.Priority, ",", -1)
			return d.Where("priority in (?)", s_priority)
		} else {
			return d
		}
	}

	filterTime := func(d *gorm.DB) *gorm.DB {
		cond := fmt.Sprintf(`update_at >= FROM_UNIXTIME(%v) AND
		 update_at <= FROM_UNIXTIME(%v)`, inputs.StartTime, inputs.EndTime)
		return d.Where(cond)
	}
	db.Alarm.Table("event_cases").Scopes(filterGroup, filterStatus, filterMetric, filterProcessStatus, filterPriority, filterTime).Find(&cevens)

	//var strategy falcon_portal.Strategy
	//var expression falcon_portal.Expression
	tmp_fid := struct {
		FId int64 `json:"fid" gorm:"column:fid"`
	}{}

	for i, event := range cevens {
		if event.StrategyId > 0 && event.ExpressionId == 0 {
			dt := db.Falcon.Table("strategy").Select("fid").Where("id = ?", event.StrategyId).Find(&tmp_fid)
			if dt.Error != nil {
				//h.JSONR(c, expecstatus, "error occurred while fetching strategy.fid")
				//return
				continue
			}
		} else if event.StrategyId == 0 && event.ExpressionId > 0 {
			dt := db.Falcon.Table("expression").Select("fid").Where("id = ?", event.ExpressionId).Find(&tmp_fid)
			if dt.Error != nil {
				//h.JSONR(c, expecstatus, "error occurred while fetching expression.fid")
				//return
				continue
			}
		}
		cevens[i].FId = tmp_fid.FId
	}

	h.JSONR(c, cevens)
}

type APITotalOfEventCasesGetInputs struct {
	StartTime     string `json:"startTime" form:"startTime"`
	EndTime       string `json:"endTime" form:"endTime"`
	Priority      string `json:"priority" form:"priority"`
	Metric        string `json:"metric" form:"metric"`
	Status        string `json:"status" form:"status"`
	ProcessStatus string `json:"process_status" form:"process_status"`

	GrpName string `json:"grp_name" form:"grp_name"`
	//number of reacord's limit on each page
	Top  string `json:"top" form:"top"`
	Sort string `json:"sort" form:"sort"`
}

func GetTotalOfEventCases(c *gin.Context) {
	var inputs APITotalOfEventCasesGetInputs
	inputs.GrpName = c.DefaultQuery("grp_name", "")
	inputs.Status = c.DefaultQuery("status", "")
	inputs.ProcessStatus = c.DefaultQuery("process_status", "unresolved")
	inputs.Sort = c.DefaultQuery("sort", "desc")
	inputs.Top = c.DefaultQuery("top", "")
	inputs.Metric = c.DefaultQuery("metric", "all")
	inputs.Priority = c.DefaultQuery("priority", "-1")
	inputs.StartTime = c.DefaultQuery("startTime", strconv.FormatInt(time.Now().Unix()-3600, 10))
	inputs.EndTime = c.DefaultQuery("endTime", strconv.FormatInt(time.Now().Unix(), 10))

	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	validProcessStatus := []string{"ignored", "unresolved", "resolved", `in progress`, "comment"}
	if !stringInSlice(inputs.ProcessStatus, validProcessStatus) {
		h.JSONR(c, badstatus, fmt.Sprintf(`The valid value of ProcessStatus is one of %v`, validProcessStatus))
		return
	}

	filterGroup := func(d *gorm.DB) *gorm.DB {
		if inputs.GrpName != "" {
			s_group := strings.SplitN(inputs.GrpName, ",", -1)
			return d.Where("grp_name in (?)", s_group)
		} else {
			return d
		}
	}

	filterStatus := func(d *gorm.DB) *gorm.DB {
		if inputs.Status != "" {
			return d.Where("status = ?", inputs.Status)
		} else {
			return d
		}
	}

	filterProcessStatus := func(d *gorm.DB) *gorm.DB {
		if inputs.ProcessStatus != "" {
			return d.Where("process_status = ?", inputs.ProcessStatus)
		} else {
			return d
		}
	}

	filterPriority := func(d *gorm.DB) *gorm.DB {
		if inputs.Priority != "-1" {
			s_priority := strings.SplitN(inputs.Priority, ",", -1)
			return d.Where("priority in (?)", s_priority)
		} else {
			return d
		}
	}

	filterMetric := func(d *gorm.DB) *gorm.DB {
		if inputs.Metric == "hardware" {
			return d.Where("metric like 'hardware.%'")
		} else if inputs.Metric == "non-hardware" {
			return d.Where("metric not like 'hardware.%'")
		} else {
			return d
		}
	}

	filterTime := func(d *gorm.DB) *gorm.DB {
		cond := fmt.Sprintf(`update_at >= FROM_UNIXTIME(%v) AND
		 update_at <= FROM_UNIXTIME(%v)`, inputs.StartTime, inputs.EndTime)
		return d.Where(cond)
	}

	orderBy := func(d *gorm.DB) *gorm.DB {
		if strings.ToUpper(inputs.Sort) == "ASC" {
			return d.Order("COUNT(grp_id) ASC")
		} else {
			return d.Order("COUNT(grp_id) DESC")
		}
	}

	recordLimit := func(d *gorm.DB) *gorm.DB {
		if inputs.Top != "" {
			return d.Limit(inputs.Top)
		} else {
			return d
		}
	}

	type Ret struct {
		Grp_name string `json:"grp_name"`
		Count    int    `json:"count"`
	}

	r := []Ret{}

	//for get correct table name
	table := alm.EventCases{}.TableName()

	db.Alarm.Table(table).Select("grp_name, COUNT(grp_id) as count").Group("grp_name").Scopes(filterGroup,
		filterStatus, filterProcessStatus, filterPriority, filterMetric, filterTime, orderBy, recordLimit).Scan(&r)

	h.JSONR(c, r)
}

//added by vincent.zhang for alarms panel, ignored maintained host
type APIGetEventCasesInputs struct {
	StartTime string `json:"startTime" form:"startTime"`
	EndTime   string `json:"endTime" form:"endTime"`
	Priority  string `json:"priority" form:"priority"`
	Metric    string `json:"metric" form:"metric"`
	GrpName   string `json:"grp_name" form:"grp_name"`
	//number of reacord's limit on each page
	Sort string `json:"sort" form:"sort"`
	Top  string `json:"top" form:"top"`
}

func GetEventCasesTotal(c *gin.Context) {
	var inputs APIGetEventCasesInputs
	inputs.GrpName = c.DefaultQuery("grp_name", "")
	inputs.Top = c.DefaultQuery("top", "")
	inputs.Sort = c.DefaultQuery("sort", "")
	inputs.Metric = c.DefaultQuery("metric", "all")
	inputs.Priority = c.DefaultQuery("priority", "-1")
	inputs.StartTime = c.DefaultQuery("startTime", strconv.FormatInt(time.Now().Unix()-3600, 10))
	inputs.EndTime = c.DefaultQuery("endTime", strconv.FormatInt(time.Now().Unix(), 10))

	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	type Ret struct {
		Grp_name string `json:"grp_name"`
		Count    int    `json:"count"`
	}

	group_count := []Ret{}

	now := time.Now().Unix()

	filterHost := func(d *gorm.DB) *gorm.DB {
		return d.Where(`h.hostname is null or h.maintain_begin > ? or h.maintain_end < ?`, now, now)
	}

	filterStatus := func(d *gorm.DB) *gorm.DB {
		return d.Where(`e.status='PROBLEM' and e.process_status='unresolved'`)
	}

	filterPriority := func(d *gorm.DB) *gorm.DB {
		if inputs.Priority != "-1" {
			s_priority := strings.SplitN(inputs.Priority, ",", -1)
			return d.Where("e.priority in (?)", s_priority)
		} else {
			return d
		}
	}

	filterGroup := func(d *gorm.DB) *gorm.DB {
		if inputs.GrpName != "" {
			s_group := strings.SplitN(inputs.GrpName, ",", -1)
			return d.Where("e.grp_name in (?)", s_group)
		} else {
			return d
		}
	}

	filterMetric := func(d *gorm.DB) *gorm.DB {
		if inputs.Metric == "hardware" {
			return d.Where("e.metric like 'hardware.%'")
		} else if inputs.Metric == "non-hardware" {
			return d.Where("e.metric not like 'hardware.%'")
		} else {
			return d
		}
	}

	filterTime := func(d *gorm.DB) *gorm.DB {
		cond := fmt.Sprintf(`e.update_at >= FROM_UNIXTIME(%v) AND e.update_at <= FROM_UNIXTIME(%v)`, inputs.StartTime, inputs.EndTime)
		return d.Where(cond)
	}

	orderBy := func(d *gorm.DB) *gorm.DB {
		str := strings.ToUpper(inputs.Sort)
		if str == "ASC" {
			return d.Order("COUNT(grp_id) ASC")
		} else if str == "DESC" {
			return d.Order("COUNT(grp_id) DESC")
		} else {
			return d
		}
	}

	recordLimit := func(d *gorm.DB) *gorm.DB {
		if inputs.Top != "" {
			return d.Limit(inputs.Top)
		} else {
			return d
		}
	}

	dt := db.Alarm.Table("event_cases e").Select(`e.grp_name, COUNT(e.id) as count`).Joins(`left join falcon_portal.host h on e.endpoint=h.hostname`).Scopes(filterHost, filterStatus,
		filterPriority, filterGroup, filterMetric, filterTime, orderBy, recordLimit).Group(`grp_name`).Scan(&group_count)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	h.JSONR(c, group_count)

}

func GetEventCasesDetail(c *gin.Context) {
	var inputs APIGetEventCasesInputs
	inputs.GrpName = c.DefaultQuery("grp_name", "")
	inputs.Top = c.DefaultQuery("top", "")
	inputs.Sort = c.DefaultQuery("sort", "")
	inputs.Metric = c.DefaultQuery("metric", "all")
	inputs.Priority = c.DefaultQuery("priority", "-1")
	inputs.StartTime = c.DefaultQuery("startTime", strconv.FormatInt(time.Now().Unix()-3600, 10))
	inputs.EndTime = c.DefaultQuery("endTime", strconv.FormatInt(time.Now().Unix(), 10))

	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	eCases := []alm.EventCases{}
	now := time.Now().Unix()

	filterHost := func(d *gorm.DB) *gorm.DB {
		return d.Where(`h.hostname is null or h.maintain_begin > ? or h.maintain_end < ?`, now, now)
	}

	filterStatus := func(d *gorm.DB) *gorm.DB {
		return d.Where(`e.status='PROBLEM' and e.process_status='unresolved'`)
	}

	filterPriority := func(d *gorm.DB) *gorm.DB {
		if inputs.Priority != "-1" {
			s_priority := strings.SplitN(inputs.Priority, ",", -1)
			return d.Where("e.priority in (?)", s_priority)
		} else {
			return d
		}
	}

	filterGroup := func(d *gorm.DB) *gorm.DB {
		if inputs.GrpName != "" {
			s_group := strings.SplitN(inputs.GrpName, ",", -1)
			return d.Where("e.grp_name in (?)", s_group)
		} else {
			return d
		}
	}

	filterMetric := func(d *gorm.DB) *gorm.DB {
		if inputs.Metric == "hardware" {
			return d.Where("e.metric like 'hardware.%'")
		} else if inputs.Metric == "non-hardware" {
			return d.Where("e.metric not like 'hardware.%'")
		} else {
			return d
		}
	}

	filterTime := func(d *gorm.DB) *gorm.DB {
		cond := fmt.Sprintf(`e.update_at >= FROM_UNIXTIME(%v) AND e.update_at <= FROM_UNIXTIME(%v)`, inputs.StartTime, inputs.EndTime)
		return d.Where(cond)
	}

	orderBy := func(d *gorm.DB) *gorm.DB {
		str := strings.ToUpper(inputs.Sort)
		if str == "ASC" {
			return d.Order("e.update_at ASC")
		} else if str == "DESC" {
			return d.Order("e.update_at DESC")
		} else {
			return d
		}
	}

	recordLimit := func(d *gorm.DB) *gorm.DB {
		if inputs.Top != "" {
			return d.Limit(inputs.Top)
		} else {
			return d
		}
	}

	dt := db.Alarm.Table("event_cases e").Select("e.*").Joins(`left join falcon_portal.host h on e.endpoint=h.hostname`).Scopes(filterHost, filterStatus,
		filterPriority, filterGroup, filterMetric, filterTime, orderBy, recordLimit).Scan(&eCases)

	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	h.JSONR(c, eCases)

}
