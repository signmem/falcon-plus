package host

import (
	"fmt"
	"github.com/open-falcon/falcon-plus/modules/api/config"
	"strconv"

	"github.com/gin-gonic/gin"
	h "github.com/open-falcon/falcon-plus/modules/api/app/helper"
	f "github.com/open-falcon/falcon-plus/modules/api/app/model/falcon_portal"
)

type APICreateTplPlugin struct {
	TplID  int64  `json:"tpl_id" binding:"required"`
	Plugin string `json:"plugin" binding:"required"`
}

func CreateTplPlugin(c *gin.Context) {
	var inputs APICreateTplPlugin
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	count := 0
	if db.Falcon.Table("tpl_plugin").Where("tpl_id = ?", inputs.TplID).Count(&count); count != 0 {
		h.JSONR(c, expecstatus, fmt.Sprintf(`The recoard with TplID %v exists,
			please do not create duplicately!`, inputs.TplID))
		return
	}

	tplplugin := f.TplPlugin{TplID: inputs.TplID, Plugin: inputs.Plugin}
	if dt := db.Falcon.Create(&tplplugin); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf(`Error occurred while insert data to TplPlugin.
		 Data{TplID: %v, Plugin: %v}`, inputs.TplID, inputs.Plugin))
		return
	}
	h.JSONR(c, tplplugin)
	return
}

func UpdateTplPlugin(c *gin.Context) {
	var inputs APICreateTplPlugin
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	tpl_plugin := f.TplPlugin{}
	if dt := db.Falcon.Where("tpl_id = ?", inputs.TplID).Find(&tpl_plugin); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("The recoard with tplID %v not found in table TplPlugin", inputs.TplID))
		return
	}

	tpl_plugin = f.TplPlugin{Plugin: inputs.Plugin}

	if dt := db.Falcon.Model(&tpl_plugin).Where("tpl_id = ?", inputs.TplID).Update(tpl_plugin).Find(&tpl_plugin); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf(`Error occurred while update data to TplPlugin.
		 Data{TplID: %v, Plugin: %v}`, inputs.TplID, inputs.Plugin))
		return
	}
	h.JSONR(c, tpl_plugin)
	return
}

func GetTplRelatedPlugin(c *gin.Context) {
	tplIDtmp := c.Params.ByName("tpl_id")
	if tplIDtmp == "" {
		h.JSONR(c, badstatus, "template id is missing")
		return
	}
	tplID, err := strconv.Atoi(tplIDtmp)
	if err != nil {
		config.Logger.Debugf("template id: %v", tplIDtmp)
		h.JSONR(c, badstatus, err)
		return
	}
	tpl_plugin := f.TplPlugin{}
	if dt := db.Falcon.Where("tpl_id = ?", tplID).Find(&tpl_plugin); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("The recoard with tplID %v not found in table TplPlugin", tplID))
		return
	}

	h.JSONR(c, tpl_plugin)
	return
}

func GetTplRelatedPluginViaName(c *gin.Context) {
	tplName := c.Params.ByName("tpl_name")
	if tplName == "" {
		h.JSONR(c, badstatus, "template name is missing")
		return
	}

	tpl := f.Template{}
	if dt := db.Falcon.Where("tpl_name = ?", tplName).Find(&tpl); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("The recoard with tpl_name %v not found in table Template", tplName))
		return
	}

	tpl_plugin := f.TplPlugin{}
	if dt := db.Falcon.Where("tpl_id = ?", tpl.ID).Find(&tpl_plugin); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("The recoard with tplID %v not found in table TplPlugin", tpl.ID))
		return
	}

	h.JSONR(c, map[string]string{
		"TplName": tplName,
		"Plugin":  tpl_plugin.Plugin,
	})
	return
}

func GetAllTplRelatedPlugin(c *gin.Context) {
	tpl_plugins := []f.TplPlugin{}
	if dt := db.Falcon.Find(&tpl_plugins); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}

	h.JSONR(c, tpl_plugins)
	return
}

func DeleteTplPlugin(c *gin.Context) {
	tplIDtmp := c.Params.ByName("tpl_id")
	if tplIDtmp == "" {
		h.JSONR(c, badstatus, "template id is missing")
		return
	}
	tplID, err := strconv.Atoi(tplIDtmp)
	if err != nil {
		config.Logger.Debugf("tplIDtmp: %v", tplIDtmp)
		h.JSONR(c, badstatus, err)
		return
	}
	tpl_plugin := f.TplPlugin{}
	if dt := db.Falcon.Where("tpl_id = ?", tplID).Find(&tpl_plugin); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("The recoard with tplID %v not found in table TplPlugin", tplID))
		return
	}

	if dt := db.Falcon.Where("tpl_id = ?", tplID).Delete(&tpl_plugin); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	h.JSONR(c, fmt.Sprintf("TplPlugin with tplID %v has been deleted", tplID))
	return
}
