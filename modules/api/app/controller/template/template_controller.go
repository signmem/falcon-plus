package template

import (
	"fmt"
	"github.com/signmem/falcon-plus/modules/api/config"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	h "github.com/signmem/falcon-plus/modules/api/app/helper"
	f "github.com/signmem/falcon-plus/modules/api/app/model/falcon_portal"
)

type APIGetTemplatesOutput struct {
	Templates []CTemplate `json:"templates"`
}
type CTemplate struct {
	Template   f.Template `json:"template"`
	ParentName string     `json:"parent_name"`
}

func GetTemplates(c *gin.Context) {
	var (
		limit int
		page  int
		err   error
	)
	pageTmp := c.DefaultQuery("page", "")
	limitTmp := c.DefaultQuery("limit", "")
	page, limit, err = h.PageParser(pageTmp, limitTmp)
	if err != nil {
		h.JSONR(c, badstatus, err.Error())
		return
	}
	var dt *gorm.DB
	var templates []f.Template
	q := c.DefaultQuery("q", ".+")
	if limit != -1 && page != -1 {
		dt = db.Falcon.Raw(
			fmt.Sprintf("SELECT * from tpl WHERE tpl_name regexp %s limit %d,%d", q, page, limit)).Scan(&templates)
	} else {
		dt = db.Falcon.Where("tpl_name regexp ?", q).Find(&templates)
	}
	if dt.Error != nil {
		config.Logger.Errorf("%s", dt.Error.Error())
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	output := APIGetTemplatesOutput{}
	output.Templates = []CTemplate{}
	for _, t := range templates {
		var pname string
		pname, err := t.FindParentName()
		if err != nil {
			h.JSONR(c, badstatus, err)
			return
		}
		output.Templates = append(output.Templates, CTemplate{
			Template:   t,
			ParentName: pname,
		})
	}
	h.JSONR(c, output)
	return
}

func GetTemplatesSimple(c *gin.Context) {
	var dt *gorm.DB
	templates := []f.Template{}
	q := c.DefaultQuery("q", ".+")
	dt = db.Falcon.Select("id, tpl_name").Where("tpl_name regexp ?", q).Find(&templates)
	if dt.Error != nil {
		config.Logger.Errorf("%s", dt.Error.Error())
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	h.JSONR(c, templates)
	return
}

func GetOneTemplateNameByID(c *gin.Context) {
	var dt *gorm.DB
	tplID := c.Params.ByName("tpl_id")
	if tplID == "" {
		h.JSONR(c, badstatus, "template id is missing")
		return
	}

	tplidint, err := strconv.Atoi(tplID)
	if err != nil {
		h.JSONR(c, badstatus, "template id error")
	}

	template := f.Template{}
	if dt = db.Falcon.Where("id = ?", tplidint).Find(&template); dt.Error != nil {
		h.JSONR(c, badstatus, fmt.Sprintf("template id %s does not exist!", tplID))
		return
	}

	template_id := strconv.FormatInt(template.ID,10)
	h.JSONR(c, map[string]string{
		"tpl_name": template.Name,
		"id":       template_id,
	})
	return
}

func GetATemplate(c *gin.Context) {
	tplidtmp := c.Params.ByName("tpl_id")
	if tplidtmp == "" {
		h.JSONR(c, badstatus, "tpl_id is missing")
		return
	}
	tplId, err := strconv.Atoi(tplidtmp)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	var tpl f.Template
	if dt := db.Falcon.Find(&tpl, tplId); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	var stratges []f.Strategy
	dt := db.Falcon.Where("tpl_id = ?", tplId).Find(&stratges)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	action := f.Action{}
	if tpl.ActionID != 0 {
		if dt = db.Falcon.Find(&action, tpl.ActionID); dt.Error != nil {
			h.JSONR(c, badstatus, dt.Error)
			return
		}
	}
	pname, _ := tpl.FindParentName()
	h.JSONR(c, map[string]interface{}{
		"template":    tpl,
		"stratges":    stratges,
		"action":      action,
		"parent_name": pname,
	})
	return
}

type UpdateOwnerInput struct {
	TplID    int64  `json:"tpl_id" binding:"required"`
	Owner    string `json:"tpl_owner" binding:"required"`
}

func UpdateTemplateOwner(c *gin.Context){
	var inputs UpdateOwnerInput
	err := c.Bind(&inputs)
	config.Logger.Debugf("UpdateTemplateOwner input: %v", inputs)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	user, err := h.GetUser(c)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	} else if inputs.Owner == "" {
		h.JSONR(c, badstatus, "input template_owner empty, please check it")
		return
	} else if inputs.TplID == 0 {
		h.JSONR(c, badstatus, "input TemplateID empty, please check it")
		return
	}

	if ! user.IsAdmin() {
		h.JSONR(c, badstatus, "you dont have permission.")
		return
	}

	updateTpl := map[string]interface{} {
		"CreateUser": inputs.Owner,
	}

	var tpl f.Template

	if dt :=  db.Falcon.Model(&tpl).Where("id = ?", inputs.TplID).Update(updateTpl); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	h.JSONR(c, "template ownere changed")
	return
}

type APICreateTemplateInput struct {
	Name     string `json:"name" binding:"required"`
	ParentID int64  `json:"parent_id" binding:"exists"`
	ActionID int64  `json:"action_id"`
}

func CreateTemplate(c *gin.Context) {
	var inputs APICreateTemplateInput
	err := c.Bind(&inputs)
	config.Logger.Debugf("CreateTemplate input: %v", inputs)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	user, err := h.GetUser(c)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	} else if inputs.Name == "" {
		h.JSONR(c, badstatus, "input name is empty, please check it")
		return
	}
	template := f.Template{
		Name:       inputs.Name,
		ParentID:   inputs.ParentID,
		ActionID:   inputs.ActionID,
		CreateUser: user.Name,
	}
	dt := db.Falcon.Table("tpl").Save(&template)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	h.JSONR(c, "template created")
	return
}

type APIUpdateTemplateInput struct {
	Name     string `json:"name" binding:"required"`
	ParentID int64  `json:"parent_id" binding:"exists"`
	TplID    int64  `json:"tpl_id" binding:"required"`
}

func UpdateTemplate(c *gin.Context) {
	var inputs APIUpdateTemplateInput
	err := c.Bind(&inputs)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	user, err := h.GetUser(c)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	var tpl f.Template
	if dt := db.Falcon.Find(&tpl, inputs.TplID); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	if tpl.CreateUser != user.Name && !user.IsAdmin() {
		h.JSONR(c, badstatus, "You don't have permission!")
		return
	}

	utpl := map[string]interface{}{
		"Name":     inputs.Name,
		"ParentID": inputs.ParentID,
	}
	if dt := db.Falcon.Model(&tpl).Where("id = ?", inputs.TplID).Update(utpl); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}
	h.JSONR(c, "template updated")
	return
}

func DeleteTemplate(c *gin.Context) {
	tidTmp, _ := c.Params.Get("tpl_id")
	if tidTmp == "" {
		h.JSONR(c, badstatus, "tpl_id is missing")
		return
	}
	tplId, err := strconv.Atoi(tidTmp)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	tx := db.Falcon.Begin()
	var tpl f.Template
	if dt := tx.Find(&tpl, tplId); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		tx.Rollback()
		return
	}
	actionId := tpl.ActionID
	if dt := tx.Delete(&tpl); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		tx.Rollback()
		return
	}
	if actionId != 0 {
		if dt := tx.Delete(&f.Action{}, actionId); dt.Error != nil {
			h.JSONR(c, badstatus, dt.Error)
			tx.Rollback()
			return
		}
	}
	if dt := tx.Where("tpl_id = ?", tplId).Delete(&f.Strategy{}); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		tx.Rollback()
		return
	}
	tx.Commit()
	h.JSONR(c, fmt.Sprintf("template %d has been deleted", tplId))
	return
}

type APICreateActionToTmplateInput struct {
	UIC                string `json:"uic" binding:"exists"`
	URL                string `json:"url" binding:"exists"`
	Callback           int    `json:"callback" binding:"exists"`
	BeforeCallbackSMS  int    `json:"before_callback_sms" binding:"exists"`
	AfterCallbackSMS   int    `json:"after_callback_sms" binding:"exists"`
	BeforeCallbackMail int    `json:"before_callback_mail" binding:"exists"`
	AfterCallbackMail  int    `json:"after_callback_mail" binding:"exists"`
	TplId              int64  `json:"tpl_id" binding:"required"`
}

func CreateActionToTmplate(c *gin.Context) {
	var inputs APICreateActionToTmplateInput
	err := c.Bind(&inputs)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	action := f.Action{
		UIC:                inputs.UIC,
		URL:                inputs.URL,
		Callback:           inputs.Callback,
		BeforeCallbackSMS:  inputs.BeforeCallbackSMS,
		BeforeCallbackMail: inputs.BeforeCallbackMail,
		AfterCallbackMail:  inputs.AfterCallbackMail,
		AfterCallbackSMS:   inputs.AfterCallbackSMS,
	}
	tx := db.Falcon.Begin()
	if dt := tx.Table("action").Save(&action); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		tx.Rollback()
		return
	}
	var lid []int
	tx.Raw("select LAST_INSERT_ID() as id").Pluck("id", &lid)
	aid := lid[0]
	var tpl f.Template
	if dt := tx.Find(&tpl, inputs.TplId); dt.Error != nil {
		h.JSONR(c, badstatus, fmt.Sprintf("template: %d ; %s", inputs.TplId, dt.Error.Error()))
		tx.Rollback()
		return
	}

	dt := tx.Model(&tpl).UpdateColumns(f.Template{ActionID: int64(aid)})
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		tx.Rollback()
		return
	}
	tx.Commit()
	// db.Falcon.Commit()
	h.JSONR(c, fmt.Sprintf("action is created and bind to template: %d", inputs.TplId))
	return
}

type APIUpdateActionToTmplateInput struct {
	ID                 int64  `json:"id" validate:"required"`
	UIC                string `json:"uic" binding:"exists"`
	URL                string `json:"url" binding:"exists"`
	Callback           int    `json:"callback" binding:"exists"`
	BeforeCallbackSMS  int    `json:"before_callback_sms" binding:"exists"`
	AfterCallbackSMS   int    `json:"after_callback_sms" binding:"exists"`
	BeforeCallbackMail int    `json:"before_callback_mail" binding:"exists"`
	AfterCallbackMail  int    `json:"after_callback_mail" binding:"exists"`
}

func UpdateActionToTmplate(c *gin.Context) {
	var inputs APIUpdateActionToTmplateInput
	err := c.BindJSON(&inputs)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	var action f.Action
	tx := db.Falcon.Begin()
	if dt := tx.Find(&action, inputs.ID); dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		tx.Rollback()
		return
	}

	uaction := map[string]interface{}{
		"UIC":                inputs.UIC,
		"URL":                inputs.URL,
		"Callback":           inputs.Callback,
		"BeforeCallbackSMS":  inputs.BeforeCallbackSMS,
		"BeforeCallbackMail": inputs.BeforeCallbackMail,
		"AfterCallbackMail":  inputs.AfterCallbackMail,
		"AfterCallbackSMS":   inputs.AfterCallbackSMS,
	}
	dt := tx.Model(&action).Where("id = ?", inputs.ID).Update(uaction)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		tx.Rollback()
		return
	}
	tx.Commit()
	h.JSONR(c, fmt.Sprintf("action is updated, row affected: %d", dt.RowsAffected))
	return
}

func GetActionByID(c *gin.Context) {
	aid := c.Param("act_id")
	act_id, err := strconv.Atoi(aid)
	if err != nil {
		h.JSONR(c, badstatus, "invalid action id")
		return
	}

	act := f.Action{}
	dt := db.Falcon.Table("action").Where("id = ?", act_id).First(&act)
	if dt.Error != nil {
		h.JSONR(c, badstatus, dt.Error)
		return
	}

	h.JSONR(c, act)
}
