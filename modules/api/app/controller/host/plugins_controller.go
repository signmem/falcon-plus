package host

import (
	"errors"
	"fmt"
	"github.com/signmem/falcon-plus/modules/api/config"
	"strconv"

	"github.com/gin-gonic/gin"
	h "github.com/signmem/falcon-plus/modules/api/app/helper"
	f "github.com/signmem/falcon-plus/modules/api/app/model/falcon_portal"
)

type APICreatePluginInput struct {
	GrpId   int64  `json:"hostgroup_id" binding:"required"`
	DirPaht string `json:"dir_path" binding:"required"`
}

type APICreatePluginInputV2 struct {
	GrpName string `json:"hostgroup" binding:"required"`
	DirPath string `json:"dir_path" binding:"required"`
}

func CreatePlugin(c *gin.Context) {
	var inputs APICreatePluginInput
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	user, _ := h.GetUser(c)
	if !user.IsAdmin() {
		hostgroup := f.HostGroup{ID: inputs.GrpId}
		if dt := db.Falcon.Find(&hostgroup); dt.Error != nil {
			h.JSONR(c, expecstatus, dt.Error)
			return
		}
		if hostgroup.CreateUser != user.Name {
			h.JSONR(c, badstatus, "You don't have permission!")
			return
		}
	}

	if inputs.GrpId <= 0 {
		h.JSONR(c, badstatus, "Invalid group id")
		return
	}
	plugin := f.Plugin{Dir: inputs.DirPaht, GrpId: inputs.GrpId, CreateUser: user.Name}
	if dt := db.Falcon.Save(&plugin); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	h.JSONR(c, plugin)
	return
}

func CreatePluginV2(c *gin.Context) {
	// bind plugin to hostgroup
	var inputs APICreatePluginInputV2
	hostgroup := f.HostGroup{}
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	user, _ := h.GetUser(c)
	if !user.IsAdmin() {
		if dt := db.Falcon.Where("grp_name = ?", inputs.GrpName).Find(&hostgroup); dt.Error != nil {
			h.JSONR(c, expecstatus, dt.Error)
			return
		}
		if hostgroup.CreateUser != user.Name {
			h.JSONR(c, badstatus, "You don't have permission!")
			return
		}
	} else {
		if dt := db.Falcon.Where("grp_name = ?", inputs.GrpName).Find(&hostgroup); dt.Error != nil {
			h.JSONR(c, expecstatus, dt.Error)
			return
		}
	}

	if hostgroup.ID <= 0 {
		h.JSONR(c, badstatus, "Invalid group id")
		return
	}

	plugin := f.Plugin{Dir: inputs.DirPath, GrpId: hostgroup.ID, CreateUser: user.Name}
	if dt := db.Falcon.Save(&plugin); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	h.JSONR(c, plugin)
	return
}

type APIUnBindPluginToGroupInputs struct {
	PluginDir string `json:"plugin_dir"`
	GrpName   string `json:"grp_name"`
}

func UnBindPluginToGrp(c *gin.Context) {
	var inputs APIUnBindPluginToGroupInputs
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}


	hostgroup := f.HostGroup{}
	if dt := db.Falcon.Where("grp_name = ?", inputs.GrpName).Find(&hostgroup); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("Group %v does not exist!", inputs.GrpName))
		return
	}

	plugin := f.Plugin{}
	if dt := db.Falcon.Where("dir = ? and grp_id = ?", inputs.PluginDir, hostgroup.ID).Find(&plugin); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("grp_id %v , Plugin %v does not exist!", hostgroup.ID, inputs.PluginDir))
		return
	}

	user, _ := h.GetUser(c)
	// db.Falcon.Where("grp_id = ? and id = ?", hostgroup.ID, plugin.ID).Find(&plugin)
	switch {
		case !user.IsAdmin() && plugin.CreateUser != user.Name:
			h.JSONR(c, badstatus, errors.New("You don't have permission can do this."))
			return
	}

	if dt := db.Falcon.Where("grp_id = ? and id = ?", hostgroup.ID, plugin.ID).Delete(&plugin); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	h.JSONR(c, fmt.Sprintf("Plugin: %v is unbind of HostGroup: %v", inputs.PluginDir, inputs.GrpName))
	return
}

func GetPluginOfGrp(c *gin.Context) {
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
	plugins := []f.Plugin{}
	if dt := db.Falcon.Where("grp_id = ?", grpID).Find(&plugins); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	h.JSONR(c, plugins)
	return
}

func GetPlugins(c *gin.Context) {
	plugins := []f.Plugin{}
	if dt := db.Falcon.Select("DISTINCT dir").Find(&plugins); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	plugin_name := []string{}
	for _, v := range plugins {
		plugin_name = append(plugin_name, v.Dir)
	}

	h.JSONR(c, plugin_name)
	return
}

func DeletePlugin(c *gin.Context) {
	pluginIDtmp := c.Params.ByName("id")
	if pluginIDtmp == "" {
		h.JSONR(c, badstatus, "plugin id is missing")
		return
	}
	pluginID, err := strconv.Atoi(pluginIDtmp)
	if err != nil {
		config.Logger.Debugf("pluginIDtmp: %v", pluginIDtmp)
		h.JSONR(c, badstatus, err)
		return
	}
	plugin := f.Plugin{ID: int64(pluginID)}
	if dt := db.Falcon.Find(&plugin); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	user, _ := h.GetUser(c)
	if !user.IsAdmin() {
		hostgroup := f.HostGroup{ID: plugin.GrpId}
		if dt := db.Falcon.Find(&hostgroup); dt.Error != nil {
			h.JSONR(c, expecstatus, dt.Error)
			return
		}
		if hostgroup.CreateUser != user.Name && plugin.CreateUser != user.Name {
			h.JSONR(c, badstatus, "You don't have permission!")
			return
		}
	}

	if dt := db.Falcon.Delete(&plugin); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	h.JSONR(c, fmt.Sprintf("plugin:%v has been deleted", pluginID))
	return
}
