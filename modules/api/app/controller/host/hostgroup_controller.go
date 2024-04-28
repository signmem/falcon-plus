package host

import (
	"errors"
	"fmt"
	"github.com/open-falcon/falcon-plus/modules/api/config"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	h "github.com/open-falcon/falcon-plus/modules/api/app/helper"
	f "github.com/open-falcon/falcon-plus/modules/api/app/model/falcon_portal"
	u "github.com/open-falcon/falcon-plus/modules/api/app/utils"
)

func GetHostGroups(c *gin.Context) {
	var (
		limit int
		page  int
		err   error
	)
	pageTmp := c.DefaultQuery("page", "")
	limitTmp := c.DefaultQuery("limit", "")
	q := c.DefaultQuery("q", ".+")
	page, limit, err = h.PageParser(pageTmp, limitTmp)
	if err != nil {
		h.JSONR(c, badstatus, err.Error())
		return
	}
	var hostgroups []f.HostGroup
	var dt *gorm.DB
	if limit != -1 && page != -1 {
		dt = db.Falcon.Raw(fmt.Sprintf("SELECT * from grp  where grp_name regexp '%s' limit %d,%d", q, page, limit)).Scan(&hostgroups)
	} else {
		dt = db.Falcon.Table("grp").Where("grp_name regexp ?", q).Find(&hostgroups)
	}
	if dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	h.JSONR(c, hostgroups)
	return
}

func GetHostGroupsWithTplPlugin(c *gin.Context) {
	hostgroups := []struct {
		ID   int64  `json:"id" gorm:"column:id"`
		Name string `json:"grp_name" gorm:"column:grp_name"`
	}{}

	type template struct {
		Name string `json:"tpl_name" gorm:"column:tpl_name"`
	}

	type plugin struct {
		Name string `json:"dir" gorm:"column:dir"`
	}

	dt := db.Falcon.Table("grp").Select("id, grp_name").Find(&hostgroups)
	if dt.Error != nil {
		h.JSONR(c, expecstatus, "error occurred while fetching hostgroups")
		return
	}

	ret := make([]map[string]interface{}, 0, 2000)

	for _, v := range hostgroups {
		q_tpl := []template{}
		dt = db.Falcon.Raw(fmt.Sprintf("select tpl_name from grp_tpl, tpl where grp_tpl.tpl_id=tpl.id and grp_tpl.grp_id=%d", v.ID)).Scan(&q_tpl)
		if dt.Error != nil {
			h.JSONR(c, expecstatus, "error occurred while fetching tpl")
			return
		}

		q_plugin := []plugin{}
		dt = db.Falcon.Raw(fmt.Sprintf("select DISTINCT dir from grp_tpl, plugin_dir where grp_tpl.grp_id=plugin_dir.grp_id and grp_tpl.grp_id=%d", v.ID)).Scan(&q_plugin)
		if dt.Error != nil {
			h.JSONR(c, expecstatus, "error occurred while fetching plugin")
			return
		}

		ret = append(ret, map[string]interface{}{
			"group":   v.Name,
			"tpls":    q_tpl,
			"plugins": q_plugin,
		})
	}

	h.JSONR(c, ret)
	return
}

type APICrateHostGroup struct {
	Name string `json:"name" binding:"required"`
}

func CrateHostGroup(c *gin.Context) {
	var inputs APICrateHostGroup
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	user, _ := h.GetUser(c)
	hostgroup := f.HostGroup{Name: inputs.Name, CreateUser: user.Name, ComeFrom: 1}
	if dt := db.Falcon.Create(&hostgroup); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	h.JSONR(c, hostgroup)
	return
}

type APIBindHostToHostGroupInput struct {
	Hosts       []string `json:"hosts" binding:"required"`
	HostGroupID int64    `json:"hostgroup_id" binding:"required"`
}

func BindHostToHostGroup(c *gin.Context) {
	var inputs APIBindHostToHostGroupInput
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	user, _ := h.GetUser(c)
	hostgroup := f.HostGroup{ID: inputs.HostGroupID}

	if dt := db.Falcon.Find(&hostgroup); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}

	if !user.IsAdmin() && hostgroup.CreateUser != user.Name {
		h.JSONR(c, badstatus, "You don't have permission.")
		return
	}

	tx := db.Falcon.Begin()
	if dt := tx.Where("grp_id = ?", hostgroup.ID).Delete(&f.GrpHost{}); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete grp_host got error: %v", dt.Error))
		dt.Rollback()
		return
	}
	var ids []int64
	for _, host := range inputs.Hosts {
		ahost := f.Host{Hostname: host}
		var id int64
		var ok bool
		if id, ok = ahost.Existing(); ok {
			ids = append(ids, id)
		} else {
			if dt := tx.Save(&ahost); dt.Error != nil {
				h.JSONR(c, expecstatus, dt.Error)
				tx.Rollback()
				return
			}
			id = ahost.ID
			ids = append(ids, id)
		}
		if dt := tx.Debug().Create(&f.GrpHost{GrpID: hostgroup.ID, HostID: id}); dt.Error != nil {
			h.JSONR(c, expecstatus, fmt.Sprintf("create grphost got error: %s , grp_id: %v, host_id: %v", dt.Error, hostgroup.ID, id))
			tx.Rollback()
			return
		}
	}
	tx.Commit()
	h.JSONR(c, fmt.Sprintf("%v bind to hostgroup: %v", ids, hostgroup.ID))
	return
}

type APIAddHostInput struct {
	Hosts       []string `json:"hosts" binding:"required"`
}

func AddHosts(c *gin.Context) {
	var inputs APIAddHostInput
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	//user, _ := h.GetUser(c)
	tx := db.Falcon.Begin()
	var ids []int64
	for _, host := range inputs.Hosts {
		ahost := f.Host{Hostname: host}
		var id int64
		var ok bool
		if id, ok = ahost.Existing(); ok {
			ids = append(ids, id)
		} else {
			if dt := tx.Save(&ahost); dt.Error != nil {
				h.JSONR(c, expecstatus, dt.Error)
				tx.Rollback()
				return
			}
			id = ahost.ID
			ids = append(ids, id)
		}
	}
	tx.Commit()
	h.JSONR(c, fmt.Sprintf("%v", ids))
	return
}

type APIBindHostToHostGroupInputV2 struct {
	Hosts     []string `json:"hosts" binding:"required"`
	HostGroup string   `json:"hostgroup" binding:"required"`
}

func BindHostToHostGroupV2(c *gin.Context) {
	var inputs APIBindHostToHostGroupInputV2
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	if strings.Contains(strings.ToLower(strings.Join(inputs.Hosts, ",")), "localhost") {
		h.JSONR(c, expecstatus, fmt.Sprintf("The input hosts [%v] contains 'localhost', Please do not use localhost as hostname",
			strings.Join(inputs.Hosts, ",")))
		return
	}

	hostgroup := f.HostGroup{}
	if dt := db.Falcon.Where("grp_name = ?", inputs.HostGroup).Find(&hostgroup); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("Group %v does not exist!", inputs.HostGroup))
		return
	}

	user, _ := h.GetUser(c)

	if !user.IsAdmin() && hostgroup.CreateUser != user.Name {
		h.JSONR(c, badstatus, "You don't have permission.")
		return
	}

	tx := db.Falcon.Begin()
	/*
		if dt := tx.Where("grp_id = ?", hostgroup.ID).Delete(&f.GrpHost{}); dt.Error != nil {
			h.JSONR(c, expecstatus, fmt.Sprintf("delete grp_host got error: %v", dt.Error))
			dt.Rollback()
			return
		}
	*/
	var ids []int64
	for _, host := range inputs.Hosts {
		ahost := f.Host{Hostname: host}
		var id int64
		var ok bool
		if id, ok = ahost.Existing(); ok {
			ids = append(ids, id)
		} else {
			h.JSONR(c, expecstatus, fmt.Sprintf("Host %v does not exist! Rollback Operations!", ahost.Hostname))
			tx.Rollback()
			return
		}

		if dt := tx.Where("grp_id = ? AND host_id = ?", hostgroup.ID, id).Delete(&f.GrpHost{}); dt.Error != nil {
			h.JSONR(c, expecstatus, fmt.Sprintf("delete grp_host got error: %v", dt.Error))
			dt.Rollback()
			return
		}

		if dt := tx.Debug().Create(&f.GrpHost{GrpID: hostgroup.ID, HostID: id}); dt.Error != nil {
			h.JSONR(c, expecstatus, fmt.Sprintf("create grphost got error: %s , grp_id: %v, host_id: %v", dt.Error, hostgroup.ID, id))
			tx.Rollback()
			return
		}
	}
	tx.Commit()
	h.JSONR(c, fmt.Sprintf("Hosts %v with IDs %v bind to hostgroup: %v", inputs.Hosts, ids, hostgroup.Name))
	return
}

type APIUnBindAHostToHostGroup struct {
	HostID      int64 `json:"host_id" binding:"required"`
	HostGroupID int64 `json:"hostgroup_id" binding:"required"`
}

func UnBindAHostToHostGroup(c *gin.Context) {
	var inputs APIUnBindAHostToHostGroup
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	user, _ := h.GetUser(c)
	hostgroup := f.HostGroup{ID: inputs.HostGroupID}

	if !user.IsAdmin() && hostgroup.CreateUser != user.Name {
		h.JSONR(c, badstatus, "You don't have permission.")
		return
	}

	if dt := db.Falcon.Find(&hostgroup); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	if dt := db.Falcon.Where("grp_id = ? AND host_id = ?", inputs.HostGroupID, inputs.HostID).Delete(&f.GrpHost{}); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	h.JSONR(c, fmt.Sprintf("unbind host:%v of hostgroup: %v", inputs.HostID, inputs.HostGroupID))
	return
}

type APIUnBindAHostToHostGroupV2 struct {
	Host      string `json:"host" binding:"required"`
	HostGroup string `json:"hostgroup" binding:"required"`
}

func UnBindAHostToHostGroupV2(c *gin.Context) {
	var inputs APIUnBindAHostToHostGroupV2
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	user, _ := h.GetUser(c)

	host := f.Host{}
	if dt := db.Falcon.Where("hostname = ?", inputs.Host).Find(&host); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("Host %v does not exist!", inputs.Host))
		return
	}

	//hostgroup := f.HostGroup{ID: inputs.HostGroup}
	hostgroup := f.HostGroup{}

	if !user.IsAdmin() && hostgroup.CreateUser != user.Name {
		h.JSONR(c, badstatus, "You don't have permission.")
		return
	}

	if dt := db.Falcon.Where("grp_name = ?", inputs.HostGroup).Find(&hostgroup); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("Group %v does not exist!", inputs.HostGroup))
		return
	}

	if dt := db.Falcon.Where("grp_id = ? AND host_id = ?", hostgroup.ID, host.ID).Delete(&f.GrpHost{}); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	h.JSONR(c, fmt.Sprintf("unbind host: [%v: %v] of hostgroup: [%v: %v]", inputs.Host, host.ID, inputs.HostGroup, hostgroup.ID))
	return
}

func DeleteHostGroup(c *gin.Context) {
	grpIDtmp := c.Params.ByName("host_group")

	if grpIDtmp == "" {
		h.JSONR(c, badstatus, "grp id is missing")
		return
	}

	grpID, err := strconv.Atoi(grpIDtmp)

	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	user, _ := h.GetUser(c)
	hostgroup := f.HostGroup{ID: int64(grpID)}

	if !user.IsAdmin() && hostgroup.CreateUser != user.Name {
		h.JSONR(c, badstatus, "You don't have permission.")
		return
	}

	if dt := db.Falcon.Find(&hostgroup); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}


	tx := db.Falcon.Begin()
	//delete hostgroup referance of grp_host table
	if dt := tx.Where("grp_id = ?", grpID).Delete(&f.GrpHost{}); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete grp_host got error: %v", dt.Error))
		dt.Rollback()
		return
	}
	// delete template of hostgroup
	if dt := tx.Where("grp_id = ?", grpID).Delete(&f.GrpTplV2{}); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete template got error: %v", dt.Error))
		dt.Rollback()
		return
	}
	//delete plugins of hostgroup
	if dt := tx.Where("grp_id = ?", grpID).Delete(&f.Plugin{}); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete plugins got error: %v", dt.Error))
		dt.Rollback()
		return
	}
	//delete aggreators of hostgroup
	if dt := tx.Where("grp_id = ?", grpID).Delete(&f.Cluster{}); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete aggreators got error: %v", dt.Error))
		dt.Rollback()
		return
	}
	//finally delete hostgroup
	//if dt := tx.Delete(&f.HostGroup{ID: int64(grpID)}); dt.Error != nil {
	if dt := tx.Where("id = ?", grpID).Delete(&f.HostGroup{}); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		tx.Rollback()
		return
	}
	tx.Commit()
	h.JSONR(c, fmt.Sprintf("hostgroup:%v has been deleted", grpID))
	return
}

func DeleteHostGroupV2(c *gin.Context) {
	grpName := c.Params.ByName("host_group")
	if grpName == "" {
		h.JSONR(c, badstatus, "grp name is missing")
		return
	}
	user, _ := h.GetUser(c)
	hostgroup := f.HostGroup{}

	if !user.IsAdmin() && hostgroup.CreateUser != user.Name {
		h.JSONR(c, badstatus, "You don't have permission.")
		return
	}

	if dt := db.Falcon.Where("grp_name = ?", grpName).Find(&hostgroup); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("Group %v does not exist!", grpName))
		return
	}

	tx := db.Falcon.Begin()
	if dt := tx.Where("grp_id = ?", hostgroup.ID).Delete(&f.GrpTplV2{}); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete template got error: %v", dt.Error))
		dt.Rollback()
		return
	}
	//delete hostgroup referance of grp_host table
	if dt := tx.Where("grp_id = ?", hostgroup.ID).Delete(&f.GrpHost{}); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete grp_host got error: %v", dt.Error))
		dt.Rollback()
		return
	}
	//delete plugins of hostgroup
	if dt := tx.Where("grp_id = ?", hostgroup.ID).Delete(&f.Plugin{}); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete plugins got error: %v", dt.Error))
		dt.Rollback()
		return
	}
	//delete aggreators of hostgroup
	if dt := tx.Where("grp_id = ?", hostgroup.ID).Delete(&f.Cluster{}); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("delete aggreators got error: %v", dt.Error))
		dt.Rollback()
		return
	}
	//finally delete hostgroup
	//if dt := tx.Delete(&f.HostGroup{ID: hostgroup.ID}); dt.Error != nil {
	if dt := tx.Where("id = ?", hostgroup.ID).Delete(&f.HostGroup{}); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		tx.Rollback()
		return
	}
	tx.Commit()
	h.JSONR(c, fmt.Sprintf("hostgroup:%v with ID %v has been deleted", grpName, hostgroup.ID))
	return
}

func GetHostGroup(c *gin.Context) {
	grpIDtmp := c.Params.ByName("host_group")
	q := c.DefaultQuery("q", ".+")
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
			if ok, err := regexp.MatchString(q, host.Hostname); ok == true && err == nil {
				hosts = append(hosts, host)
			}
		}
	}
	h.JSONR(c, map[string]interface{}{
		"hostgroup": hostgroup,
		"hosts":     hosts,
	})
	return
}

type APIBindTemplateToGroupInputs struct {
	TplID int64 `json:"tpl_id"`
	GrpID int64 `json:"grp_id"`
}

func BindTemplateToGroup(c *gin.Context) {
	var inputs APIBindTemplateToGroupInputs
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	user, _ := h.GetUser(c)
	grpTpl := f.GrpTpl{
		GrpID: inputs.GrpID,
		TplID: inputs.TplID,
	}
	db.Falcon.Where("grp_id = ? and tpl_id = ?", inputs.GrpID, inputs.TplID).Find(&grpTpl)
	if grpTpl.BindUser != "" {
		h.JSONR(c, badstatus, errors.New("this binding already existing, reject!"))
		return
	}
	grpTpl.BindUser = user.Name
	if dt := db.Falcon.Save(&grpTpl); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	h.JSONR(c, grpTpl)
	return
}

type APIBindTemplateToGroupInputsV2 struct {
	TplName string `json:"tpl_name"`
	GrpName string `json:"grp_name"`
}

func BindTemplateToGroupV2(c *gin.Context) {
	var inputs APIBindTemplateToGroupInputsV2
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	template := f.Template{}
	if dt := db.Falcon.Where("tpl_name = ?", inputs.TplName).Find(&template); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("Template %v does not exist!", inputs.TplName))
		return
	}

	hostgroup := f.HostGroup{}
	if dt := db.Falcon.Where("grp_name = ?", inputs.GrpName).Find(&hostgroup); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("Group %v does not exist!", inputs.GrpName))
		return
	}

	user, _ := h.GetUser(c)
	grpTpl := f.GrpTpl{
		GrpID: hostgroup.ID,
		TplID: template.ID,
	}
	grpTplV2 := f.GrpTplV2{
		GrpID:   hostgroup.ID,
		GrpName: inputs.GrpName,
		TplID:   template.ID,
		TplName: inputs.TplName,
	}

	db.Falcon.Where("grp_id = ? and tpl_id = ?", hostgroup.ID, template.ID).Find(&grpTpl)
	if grpTpl.BindUser != "" {
		h.JSONR(c, badstatus, errors.New("this binding already existing, reject!"))
		return
	}
	grpTpl.BindUser = user.Name
	grpTplV2.BindUser = user.Name
	if dt := db.Falcon.Save(&grpTpl); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	h.JSONR(c, grpTplV2)
	return
}

type APIUnBindTemplateToGroupInputs struct {
	TplID int64 `json:"tpl_id"`
	GrpID int64 `json:"grp_id"`
}

func UnBindTemplateToGroup(c *gin.Context) {
	var inputs APIUnBindTemplateToGroupInputs
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	user, _ := h.GetUser(c)
	grpTpl := f.GrpTpl{
		GrpID: inputs.GrpID,
		TplID: inputs.TplID,
	}
	db.Falcon.Where("grp_id = ? and tpl_id = ?", inputs.GrpID, inputs.TplID).Find(&grpTpl)
	switch {
	case !user.IsAdmin() && grpTpl.BindUser != user.Name:
		h.JSONR(c, badstatus, errors.New("You don't have permission can do this."))
		return
	}
	if dt := db.Falcon.Where("grp_id = ? and tpl_id = ?", inputs.GrpID, inputs.TplID).Delete(&grpTpl); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	h.JSONR(c, fmt.Sprintf("template: %v is unbind of HostGroup: %v", inputs.TplID, inputs.GrpID))
	return
}

type APIUnBindTemplateToGroupInputsV2 struct {
	TplName string `json:"tpl_name"`
	GrpName string `json:"grp_name"`
}

func UnBindTemplateToGroupV2(c *gin.Context) {
	var inputs APIUnBindTemplateToGroupInputsV2
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	template := f.Template{}
	if dt := db.Falcon.Where("tpl_name = ?", inputs.TplName).Find(&template); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("Template %v does not exist!", inputs.TplName))
		return
	}

	hostgroup := f.HostGroup{}
	if dt := db.Falcon.Where("grp_name = ?", inputs.GrpName).Find(&hostgroup); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("Group %v does not exist!", inputs.GrpName))
		return
	}

	user, _ := h.GetUser(c)
	grpTpl := f.GrpTpl{
		GrpID: hostgroup.ID,
		TplID: template.ID,
	}
	db.Falcon.Where("grp_id = ? and tpl_id = ?", hostgroup.ID, template.ID).Find(&grpTpl)
	switch {
	case grpTpl.BindUser == "":
		h.JSONR(c, badstatus, fmt.Sprintf("HostGroup %v does not bind to Template %v", inputs.GrpName, inputs.TplName))
		return
	case !user.IsAdmin() && grpTpl.BindUser != user.Name:
		h.JSONR(c, badstatus, errors.New("You don't have permission can do this."))
		return
	}
	if dt := db.Falcon.Where("grp_id = ? and tpl_id = ?", hostgroup.ID, template.ID).Delete(&grpTpl); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	h.JSONR(c, fmt.Sprintf("template: %v is unbind of HostGroup: %v", inputs.TplName, inputs.GrpName))
	return
}

func GetTemplateOfHostGroup(c *gin.Context) {
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
	grpTpls := []f.GrpTpl{}
	Tpls := []f.Template{}
	db.Falcon.Where("grp_id = ?", grpID).Find(&grpTpls)
	if len(grpTpls) != 0 {
		tips := []int64{}
		for _, t := range grpTpls {
			tips = append(tips, t.TplID)
		}
		tipsStr, _ := u.ArrInt64ToString(tips)
		db.Falcon.Where(fmt.Sprintf("id in (%s)", tipsStr)).Find(&Tpls)
	}
	h.JSONR(c, map[string]interface{}{
		"hostgroup": hostgroup,
		"templates": Tpls,
	})
	return
}

func GetTemplateOfHostGroupV2(c *gin.Context) {
	grpName := c.Params.ByName("host_group")
	if grpName == "" {
		h.JSONR(c, badstatus, "grp name is missing")
		return
	}

	hostgroup := f.HostGroup{}
	if dt := db.Falcon.Where("grp_name = ?", grpName).Find(&hostgroup); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("Group %v does not exist!", grpName))
		return
	}
	grpTpls := []f.GrpTpl{}
	Tpls := []f.Template{}
	db.Falcon.Where("grp_id = ?", hostgroup.ID).Find(&grpTpls)
	if len(grpTpls) != 0 {
		tips := []int64{}
		for _, t := range grpTpls {
			tips = append(tips, t.TplID)
		}
		tipsStr, _ := u.ArrInt64ToString(tips)
		db.Falcon.Where(fmt.Sprintf("id in (%s)", tipsStr)).Find(&Tpls)
	}
	h.JSONR(c, map[string]interface{}{
		"hostgroup": hostgroup,
		"templates": Tpls,
	})
	return
}

func GetHostGroupInfo(c *gin.Context) {
	grpName := c.Params.ByName("host_group")
	if grpName == "" {
		h.JSONR(c, badstatus, "grp name is missing")
		return
	}

	hostgroup := f.HostGroup{}
	if dt := db.Falcon.Where("grp_name = ?", grpName).Find(&hostgroup); dt.Error != nil {
		h.JSONR(c, badstatus, fmt.Sprintf("hostgroup %s not exists", grpName))
		return
	}
	hostgroup_id := strconv.FormatInt(hostgroup.ID,10)
	h.JSONR(c, map[string]string{
		"id": hostgroup_id,
		"hostgroup": hostgroup.Name,
	})
	return
}