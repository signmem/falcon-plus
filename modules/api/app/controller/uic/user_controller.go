package uic

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	h "github.com/signmem/falcon-plus/modules/api/app/helper"
	"github.com/signmem/falcon-plus/modules/api/app/model/uic"
	"github.com/signmem/falcon-plus/modules/api/app/utils"
)

type APIUserInput struct {
	Name   string `json:"name" binding:"required"`
	Cnname string `json:"cnname" binding:"required"`
	Passwd string `json:"password" binding:"required"`
	Email  string `json:"email" binding:"required"`
	Phone  string `json:"phone"`
	IM     string `json:"im"`
	QQ     string `json:"qq"`
}

func CreateUser(c *gin.Context) {
	var inputs APIUserInput
	err := c.Bind(&inputs)
	switch {
	case err != nil:
		h.JSONR(c, http.StatusBadRequest, err)
		return
	case utils.HasDangerousCharacters(inputs.Cnname):
		h.JSONR(c, http.StatusBadRequest, "name pattern is invalid")
		return
	}
	var user uic.User
	db.Uic.Table("user").Where("name = ?", inputs.Name).Scan(&user)
	if user.ID != 0 {
		h.JSONR(c, http.StatusBadRequest, "name is already existing")
		return
	}
	password := utils.HashIt(inputs.Passwd)
	user = uic.User{
		Name:   inputs.Name,
		Passwd: password,
		Cnname: inputs.Cnname,
		Email:  inputs.Email,
		Phone:  inputs.Phone,
		IM:     inputs.IM,
		QQ:     inputs.QQ,
	}

	//for create a root user during the first time
	if inputs.Name == "root" {
		user.Role = 2
	}

	dt := db.Uic.Table("user").Create(&user)
	if dt.Error != nil {
		h.JSONR(c, http.StatusBadRequest, dt.Error)
		return
	}

	var session uic.Session
	response := map[string]string{}
	s := db.Uic.Table("session").Where("uid = ?", user.ID).Scan(&session)
	if s.Error != nil && s.Error.Error() != "record not found" {
		h.JSONR(c, http.StatusBadRequest, s.Error)
		return
	} else if session.ID == 0 {
		session.Sig = utils.GenerateUUID()
		session.Expired = int(time.Now().Unix()) + 3600*24*30
		session.Uid = user.ID
		db.Uic.Create(&session)
	}

	response["sig"] = session.Sig
	response["name"] = user.Name
	h.JSONR(c, http.StatusOK, response)
	return
}

type APIUserUpdateInput struct {
	Cnname string `json:"cnname" binding:"required"`
	Email  string `json:"email" binding:"required"`
	Phone  string `json:"phone"`
	IM     string `json:"im"`
	QQ     string `json:"qq"`
}

//update current user profile
func UpdateCurrentUser(c *gin.Context) {
	var inputs APIUserUpdateInput
	err := c.BindJSON(&inputs)
	switch {
	case err != nil:
		h.JSONR(c, http.StatusExpectationFailed, err)
		return
	case utils.HasDangerousCharacters(inputs.Cnname):
		h.JSONR(c, http.StatusBadRequest, "name pattern is invalid")
		return
	}
	websession, _ := h.GetSession(c)
	user := uic.User{}
	db.Uic.Table("user").Where("name = ?", websession.Name).Scan(&user)
	if user.ID == 0 {
		h.JSONR(c, http.StatusBadRequest, "name is not existing")
		return
	}
	uid := user.ID
	uuser := map[string]interface{}{
		"Cnname": inputs.Cnname,
		"Email":  inputs.Email,
		"Phone":  inputs.Phone,
		"IM":     inputs.IM,
		"QQ":     inputs.QQ,
	}
	dt := db.Uic.Model(&user).Where("id = ?", uid).Update(uuser)
	if dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}
	h.JSONR(c, "user info updated")
	return
}

type APICgPassedInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func ChangePassword(c *gin.Context) {
	var inputs APICgPassedInput
	err := c.Bind(&inputs)
	if err != nil {
		h.JSONR(c, http.StatusBadRequest, err)
	}
	websession, _ := h.GetSession(c)
	user := uic.User{Name: websession.Name}

	dt := db.Uic.Where(&user).Find(&user)
	switch {
	case dt.Error != nil:
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	case user.Passwd != utils.HashIt(inputs.OldPassword):
		h.JSONR(c, http.StatusBadRequest, "oldPassword is not match current one")
		return
	}

	user.Passwd = utils.HashIt(inputs.NewPassword)
	dt = db.Uic.Save(&user)
	if dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}
	h.JSONR(c, http.StatusOK, "password updated!")
	return
}

func UserInfo(c *gin.Context) {
	user, err := h.GetUser(c)
	if err != nil {
		h.JSONR(c, http.StatusExpectationFailed, err)
		return
	}
	h.JSONR(c, http.StatusOK, user)
	return
}

// anyone should get the user infomation
func GetUser(c *gin.Context) {
	uidtmp := c.Params.ByName("uid")
	if uidtmp == "" {
		h.JSONR(c, badstatus, "user id is missing")
		return
	}
	uid, err := strconv.Atoi(uidtmp)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	fuser := uic.User{ID: int64(uid)}
	if dt := db.Uic.Table("user").Find(&fuser); dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}
	h.JSONR(c, fuser)
	return
}

func GetUserByName(c *gin.Context) {
	name := c.Params.ByName("user_name")
	if name == "" {
		h.JSONR(c, badstatus, "user name is missing")
		return
	}
	fuser := uic.User{}
	if dt := db.Uic.Table("user").Where("name = ?", name).First(&fuser); dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}
	h.JSONR(c, fuser)
	return
}

func IsUserInTeams(c *gin.Context) {
	uidtmp := c.Params.ByName("uid")
	if uidtmp == "" {
		h.JSONR(c, badstatus, "user id is missing")
		return
	}
	uid, err := strconv.Atoi(uidtmp)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	}

	teams_raw := c.DefaultQuery("team_names", "")
	if teams_raw == "" {
		h.JSONR(c, badstatus, err)
		return
	}
	team_names := strings.Split(teams_raw, ",")

	user := uic.User{}
	dt := db.Uic.Table("user").Where("id = ?", uid).First(&user)
	if dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}

	teams := []uic.Team{}
	dt = db.Uic.Table("team").Where("name in (?)", team_names).Find(&teams)

	//add by fandy.fang
	if len(teams) == 0 {
		h.JSONR(c, "false")
		return
	}
	//end by fandy.fang

	if dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}

	tids := []int64{}
	for _, t := range teams {
		tids = append(tids, t.ID)
	}

	tus := []uic.RelTeamUser{}
	dt = db.Uic.Table("rel_team_user").Where("uid = ? and tid in (?)", uid, tids).Find(&tus)

	//add by fandy.fang
	if len(tus) == 0 {
		h.JSONR(c, "false")
		return
	}
	//end by fandy.fang

	if dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}

	h.JSONR(c, "true")
	return
}

//admin usage

type APIAdminChangeUserProfileInput struct {
	UserID int    `json:"user_id" binding:"required"`
	Cnname string `json:"cnname" binding:"required"`
	Email  string `json:"email" binding:"required"`
	Phone  string `json:"phone"`
	IM     string `json:"im"`
	QQ     string `json:"qq"`
}

func AdminChangeUserProfile(c *gin.Context) {
	var inputs APIAdminChangeUserProfileInput
	err := c.BindJSON(&inputs)
	if err != nil {
		h.JSONR(c, http.StatusExpectationFailed, err)
		return
	}

	cuser, err := h.GetUser(c)
	if err != nil {
		h.JSONR(c, http.StatusExpectationFailed, err)
		return
	} else if !cuser.IsAdmin() {
		h.JSONR(c, http.StatusBadRequest, "you don't have permission!")
		return
	}

	user := uic.User{}
	uid := inputs.UserID
	uuser := map[string]interface{}{
		"Cnname": inputs.Cnname,
		"Email":  inputs.Email,
		"Phone":  inputs.Phone,
		"IM":     inputs.IM,
		"QQ":     inputs.QQ,
	}
	dt := db.Uic.Model(&user).Where("id = ?", uid).Update(uuser)
	if dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}
	h.JSONR(c, "user profile updated")
	return
}

type APIAdminUserDeleteInput struct {
	UserID int `json:"user_id" binding:"required"`
}

func AdminUserDelete(c *gin.Context) {
	var inputs APIAdminUserDeleteInput
	err := c.Bind(&inputs)
	if err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	cuser, err := h.GetUser(c)
	if err != nil {
		h.JSONR(c, http.StatusExpectationFailed, err)
		return
	} else if !cuser.IsAdmin() {
		h.JSONR(c, http.StatusBadRequest, "you don't have permission!")
		return
	}
	dt := db.Uic.Delete(&uic.User{}, inputs.UserID)
	if dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}
	h.JSONR(c, fmt.Sprintf("user %v has been delete, affect row: %v", inputs.UserID, dt.RowsAffected))
	return
}

type APIAdminChangePassword struct {
	UserID int    `json:"user_id" binding:"required"`
	Passwd string `json:"password" binding:"required"`
}

func AdminChangePassword(c *gin.Context) {
	var inputs APIAdminChangePassword
	err := c.Bind(&inputs)
	if err != nil {
		h.JSONR(c, http.StatusBadRequest, err)
		return
	}

	cuser, err := h.GetUser(c)
	if err != nil {
		h.JSONR(c, http.StatusExpectationFailed, err)
		return
	} else if !cuser.IsAdmin() {
		h.JSONR(c, http.StatusBadRequest, "you don't have permission!")
		return
	}

	user := uic.User{ID: int64(inputs.UserID)}
	dt := db.Uic.Where(&user).Find(&user)
	if dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}

	user.Passwd = utils.HashIt(inputs.Passwd)
	dt = db.Uic.Save(&user)
	if dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}
	h.JSONR(c, http.StatusOK, "password updated!")
	return
}

func UserList(c *gin.Context) {
	// remove admin checking
	// websession, _ := h.GetSession(c)
	// user := uic.User{Name: websession.Name}
	// dt := db.Uic.Where(&user).Find(&user)
	// switch {
	// case dt.Error != nil:
	// 	h.JSONR(c, http.StatusExpectationFailed, dt.Error)
	// 	return
	// case !user.IsAdmin():
	// 	h.JSONR(c, http.StatusBadRequest, "you don't have permission!")
	// 	return
	// }
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
	q := c.DefaultQuery("q", ".+")
	var user []uic.User
	var dt *gorm.DB
	if limit != -1 && page != -1 {
		dt = db.Uic.Raw(
			fmt.Sprintf("select * from user where name regexp '%s' limit %d,%d", q, page, limit)).Scan(&user)
	} else {
		dt = db.Uic.Table("user").Where("name regexp ?", q).Scan(&user)
	}
	if dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}
	h.JSONR(c, user)
	return
}

type APIRoleUpdate struct {
	UserID int64  `json:"user_id" binding:"required"`
	Admin  string `json:"admin" binding:"required"`
}

func ChangeRoleOfUser(c *gin.Context) {
	var inputs APIRoleUpdate
	err := c.Bind(&inputs)
	if err != nil {
		h.JSONR(c, http.StatusBadRequest, err)
		return
	}
	cuser, err := h.GetUser(c)
	switch {
	case err != nil:
		h.JSONR(c, http.StatusBadRequest, err)
		return
	case !cuser.IsAdmin():
		h.JSONR(c, http.StatusBadRequest, "you don't have permission!")
		return
	}
	var user uic.User
	db.Uic.Find(&user, inputs.UserID)
	switch inputs.Admin {
	case "yes":
		user.Role = 1
	case "no":
		user.Role = 0
	}

	dt := db.Uic.Save(&user)
	if dt.Error != nil {
		h.JSONR(c, http.StatusExpectationFailed, dt.Error)
		return
	}
	h.JSONR(c, fmt.Sprintf("user role update sccuessful, affect row: %v", dt.RowsAffected))
	return
}
