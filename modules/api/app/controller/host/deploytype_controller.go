package host

import (
	"encoding/json"
	"fmt"
	"github.com/signmem/falcon-plus/modules/api/config"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	h "github.com/signmem/falcon-plus/modules/api/app/helper"
	f "github.com/signmem/falcon-plus/modules/api/app/model/falcon_portal"
)

type APICreateDeployType struct {
	CmdbID int64  `json:"cmdb_id" binding:"required"`
	TplIDs string `json:"tpl_ids" binding:"required"`
}

func CreateDeployType(c *gin.Context) {
	var inputs APICreateDeployType
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	inputs.TplIDs = strings.TrimFunc(inputs.TplIDs, func(r rune) bool {
		return !unicode.IsNumber(r)
	})

	count := 0
	if db.Falcon.Table("deploy_type").Where("cmdb_id = ?", inputs.CmdbID).Count(&count); count != 0 {
		h.JSONR(c, expecstatus, fmt.Sprintf(`The recoard with CmdbID %v exists,
                        please do not create duplicately!`, inputs.CmdbID))
		return
	}

	deploytype := f.DeployType{CmdbID: inputs.CmdbID, TplIDs: inputs.TplIDs}
	if dt := db.Falcon.Create(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf(`Error occurred while insert data to DeployType.
                 Data{CmdbID: %v, TplIDs: %v}`, inputs.CmdbID, inputs.TplIDs))
		return
	}
	h.JSONR(c, deploytype)
	return
}

func UpdateDeployType(c *gin.Context) {
	var inputs APICreateDeployType
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	inputs.TplIDs = strings.TrimFunc(inputs.TplIDs, func(r rune) bool {
		return !unicode.IsNumber(r)
	})

	deploytype := f.DeployType{}
	if dt := db.Falcon.Where("cmdb_id = ?", inputs.CmdbID).Find(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("The recoard with cmdb id %v not found in table DeployType", inputs.CmdbID))
		return
	}

	deploytype = f.DeployType{TplIDs: inputs.TplIDs}
	if dt := db.Falcon.Model(&deploytype).Where("cmdb_id = ?", inputs.CmdbID).Update(deploytype).Find(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf(`Error occurred while update data to DeployType.
                 Data{CmdbID: %v, TplIDs: %v}`, inputs.CmdbID, inputs.TplIDs))
		return
	}
	h.JSONR(c, deploytype)
	return
}

type DeployType struct {
	DictCode  string `json:"dict_code"`
	DictValue string `json:"dict_value"`
}

type CmdbQueryResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	//Object  []map[string]interface{} `json:"object"`
	Object  []DeployType `json:"object"`
	Success bool         `json:"success"`
}

func queryComps() *CmdbQueryResponse {
	client := &http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
	url := `http://cmdb3.api.vip.com/app/deploy_type/query?sys_name=VipFalcon&key=329a55d8ddff6359c63eccb52c611140`
	req, _ := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		config.Logger.Errorf("Errored when sending request to cmdb api %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		config.Logger.Error("Query Data Failed!")
	}

	jsonData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		config.Logger.Errorf("read responses fail %s", err)
		return nil
	}

	var rJson *CmdbQueryResponse
	err = json.Unmarshal(jsonData, &rJson)
	if err != nil {
		config.Logger.Errorf("Unmarshal metric data fail %s", err)
		return nil
	}

	return rJson
}

func UpdateCmdbDeployType(c *gin.Context) {
	type DeployType struct {
		CmdbID   int64  `json:"cmdb_id" gorm:"column:cmdb_id"`
		CompName string `json:"comp_name" orm:"column:comp_name"`
	}

	dtList := queryComps()
	deploytypes := []DeployType{}

	for _, cmdbType := range dtList.Object {
		deploytype := DeployType{}
		if dt := db.Falcon.Where("cmdb_id = ?", cmdbType.DictCode).Find(&deploytype); dt.Error != nil {
			config.Logger.Infof("The recoard with cmdb id %v not found in table DeployType", cmdbType.DictCode)
			continue
		}

		deploytype.CompName = cmdbType.DictValue

		if dt := db.Falcon.Model(&deploytype).Where("cmdb_id = ?", deploytype.CmdbID).Update(deploytype).Find(&deploytype); dt.Error != nil {
			h.JSONR(c, expecstatus, fmt.Sprintf(`Error occurred while update data to DeployType.
		 Data{CmdbID: %v, CmdbType: %v}`, deploytype.CmdbID, cmdbType.DictValue))
			return
		}
		deploytypes = append(deploytypes, deploytype)
	}
	h.JSONR(c, deploytypes)
	return
}

func GetDeployType(c *gin.Context) {
	cmdbIDtmp := c.Params.ByName("cmdb_id")
	if cmdbIDtmp == "" {
		h.JSONR(c, badstatus, "cmdb id is missing")
		return
	}
	cmdbID, err := strconv.Atoi(cmdbIDtmp)
	if err != nil {
		config.Logger.Errorf("cmdb id: %v", cmdbIDtmp)
		h.JSONR(c, badstatus, err)
		return
	}
	deploytype := f.DeployType{}
	if dt := db.Falcon.Where("cmdb_id = ?", cmdbID).Find(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("The recoard with cmdb id %v not found in table DeployType", cmdbID))
		return
	}

	h.JSONR(c, deploytype)
	return
}

func GetAllDeployType(c *gin.Context) {
	deploytypes := []f.DeployType{}
	if dt := db.Falcon.Find(&deploytypes); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}

	h.JSONR(c, deploytypes)
	return
}

func DeleteDeployType(c *gin.Context) {
	cmdbIDtmp := c.Params.ByName("cmdb_id")
	if cmdbIDtmp == "" {
		h.JSONR(c, badstatus, "cmdb id is missing")
		return
	}
	cmdbID, err := strconv.Atoi(cmdbIDtmp)
	if err != nil {
		config.Logger.Debugf("cmdb id: %v", cmdbIDtmp)
		h.JSONR(c, badstatus, err)
		return
	}
	deploytype := f.DeployType{}
	if dt := db.Falcon.Where("cmdb_id = ?", cmdbID).Find(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("The recoard with cmdb id %v not found in table DeployType", cmdbID))
		return
	}

	if dt := db.Falcon.Where("cmdb_id = ?", cmdbID).Delete(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	h.JSONR(c, fmt.Sprintf("DeployType with cmdbID %v has been deleted", cmdbID))
	return
}

type APICreateDeployTypeV2 struct {
	//CmdbID int64  `json:"cmdb_id" binding:"required"`
	CompName string `json:"comp_name" binding:"required"`
	TplIDs   string `json:"tpl_ids" binding:"required"`
}

func CreateDeployTypeV2(c *gin.Context) {
	var inputs APICreateDeployTypeV2
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	inputs.TplIDs = strings.TrimFunc(inputs.TplIDs, func(r rune) bool {
		return !unicode.IsNumber(r)
	})

	count := 0
	if db.Falcon.Table("deploy_type").Where("comp_name = ?", inputs.CompName).Count(&count); count != 0 {
		h.JSONR(c, expecstatus, fmt.Sprintf(`The recoard with CompName %v exists,
			please do not create duplicately!`, inputs.CompName))
		return
	}

	deploytype := f.DeployTypeV2{CompName: inputs.CompName, TplIDs: inputs.TplIDs}
	if dt := db.Falcon.Create(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf(`Error occurred while insert data to DeployType.
		 Data{CompName: %v, TplIDs: %v}`, inputs.CompName, inputs.TplIDs))
		return
	}
	h.JSONR(c, deploytype)
	return
}

func UpdateDeployTypeV2(c *gin.Context) {
	var inputs APICreateDeployTypeV2
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, err)
		return
	}
	inputs.TplIDs = strings.TrimFunc(inputs.TplIDs, func(r rune) bool {
		return !unicode.IsNumber(r)
	})

	deploytype := f.DeployTypeV2{}
	if dt := db.Falcon.Where("comp_name = ?", inputs.CompName).Find(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("The recoard with comp_name %v not found in table DeployType", inputs.CompName))
		return
	}

	deploytype = f.DeployTypeV2{TplIDs: inputs.TplIDs}
	if dt := db.Falcon.Model(&deploytype).Where("comp_name = ?", inputs.CompName).Update(deploytype).Find(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf(`Error occurred while update data to DeployType.
		 Data{CompName: %v, TplIDs: %v}`, inputs.CompName, inputs.TplIDs))
		return
	}
	h.JSONR(c, deploytype)
	return
}

func GetDeployTypeV2(c *gin.Context) {
	comp_name := c.Params.ByName("comp_name")
	if comp_name == "" {
		h.JSONR(c, badstatus, "component name is missing")
		return
	}

	deploytype := f.DeployTypeV2{}
	if dt := db.Falcon.Where("comp_name = ?", comp_name).Find(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("The recoard with component name %v not found in table DeployType", comp_name))
		return
	}

	h.JSONR(c, deploytype)
	return
}

func GetAllDeployTypeV2(c *gin.Context) {
	deploytypes := []f.DeployTypeV2{}
	if dt := db.Falcon.Find(&deploytypes); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}

	h.JSONR(c, deploytypes)
	return
}

type APIDeployTypeOfTplPluginInput struct {
	DeployID int `json:"deploy_id" binding:"required"`
	// SubtypeID int `json:"subtype_id" binding:"required"`
}

func GetDeployTypeOfTplPlugin(c *gin.Context) {
	inputs := APIDeployTypeOfTplPluginInput{}
	if err := c.Bind(&inputs); err != nil {
		h.JSONR(c, badstatus, "deploy_id is blank")
		return
	}

	/*
		var cmdbID int
		if inputs.DeployID == 7 {
			cmdbID = inputs.SubtypeID
		} else {
			cmdbID = inputs.DeployID
		}
	*/

	cmdbID := inputs.DeployID

	deploytype := f.DeployTypeV2{}
	if dt := db.Falcon.Where("cmdb_id = ?", cmdbID).Find(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("The recoard not found with cmdb_id [%v] in the table deploytype!", cmdbID))
		return
	}

	tplIDs := strings.Split(deploytype.TplIDs, ",")

	type template struct {
		Name string `json:"tpl_name" gorm:"column:tpl_name"`
	}

	type plugin struct {
		Name string `json:"plugin" gorm:"column:plugin"`
	}

	ret := struct {
		CompName string
		Template []string
		Plugin   []string
	}{CompName: deploytype.CompName}

	for _, v := range tplIDs {
		q_tpl := template{}
		q_plugin := plugin{}
		dt := db.Falcon.Table("tpl").Select("tpl_name").Where("id = ?", v).Find(&q_tpl)
		if dt.Error != nil {
			config.Logger.Warningf("The recoard not found with tpl_id [%v] in the table tpl_name!", v)
			//h.JSONR(c, expecstatus, fmt.Sprintf("The recoard not found with tpl_id [%v] in the table tpl_name!", v))
			//return
		} else {
			ret.Template = append(ret.Template, q_tpl.Name)
		}

		dt = db.Falcon.Table("tpl_plugin").Where("tpl_id = ?", v).Find(&q_plugin)
		if dt.Error != nil {
			config.Logger.Warningf("The recoard not found with tpl_id [%v] in the table tpl_plugin!", v)
			//h.JSONR(c, expecstatus, fmt.Sprintf("The recoard not found with tpl_id [%v] in the table tpl_plugin!", v))
			//return
		} else {
			ret.Plugin = append(ret.Plugin, q_plugin.Name)
		}
	}

	h.JSONR(c, ret)
	return
}

func DeleteDeployTypeV2(c *gin.Context) {
	comp_name := c.Params.ByName("comp_name")
	if comp_name == "" {
		h.JSONR(c, badstatus, "component name is missing")
		return
	}

	deploytype := f.DeployTypeV2{}
	if dt := db.Falcon.Where("comp_name = ?", comp_name).Find(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("The recoard with cmdb id %v not found in table DeployType", comp_name))
		return
	}

	if dt := db.Falcon.Where("comp_name = ?", comp_name).Delete(&deploytype); dt.Error != nil {
		h.JSONR(c, expecstatus, dt.Error)
		return
	}
	h.JSONR(c, fmt.Sprintf("DeployType with comp_name %v has been deleted", comp_name))
	return
}

type QueryResponse struct {
	Code    int                      `json:"code"`
	Message string                   `json:"message"`
	Object  []map[string]interface{} `json:"object"`
	Success bool                     `json:"success"`
}

func queryComp(domain string) *QueryResponse {
	client := &http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
	url := fmt.Sprintf(`http://cmdb3.api.vip.com/app/query?name=%s&key=329a55d8ddff6359c63eccb52c611140&sys_name=VipFalcon`, domain)
	req, _ := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		config.Logger.Errorf("Errored when sending request to cmdb api %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		config.Logger.Error("Query Data Failed!")
	}

	jsonData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		config.Logger.Errorf("read responses fail %s", err)
		return nil
	}

	var rJson *QueryResponse
	err = json.Unmarshal(jsonData, &rJson)
	if err != nil {
		config.Logger.Errorf("Unmarshal metric data fail %s", err)
		return nil
	}

	return rJson
}

// Get component\template\plugin via domain name, according components configured in
// cmdb (http://cmdb3.sysop.vipshop.com/app_comp_list.jspx) & db table deploy_type
func GetDomainOfTplPlugin(c *gin.Context) {
	domain := c.Params.ByName("domain")
	if domain == "" {
		h.JSONR(c, badstatus, "domain name is missing")
		return
	}

	queryRet := queryComp(domain)
	cmdbCompListTmp := queryRet.Object[0]["comp_list"]

	// string type assertion
	value, ok := cmdbCompListTmp.(string)
	if !ok {
		fmt.Println("Not ok for type string")
		return
	}
	cmdbCompList := strings.Split(value, ",")

	dbCompListTmp := []struct {
		CompName string `json:"comp_name" orm:"column:comp_name"`
	}{}

	if dt := db.Falcon.Table("deploy_type").Select("comp_name").Find(&dbCompListTmp); dt.Error != nil {
		h.JSONR(c, expecstatus, fmt.Sprintf("error occurred while select comp_name from deploy_type"))
		return
	}

	dbCompList := []string{}
	if len(dbCompListTmp) != 0 {
		for _, v := range dbCompListTmp {
			dbCompList = append(dbCompList, v.CompName)
		}
	}
	dbCompListString := strings.Join(dbCompList, " ")

	compList := []string{}
	for _, v := range cmdbCompList {
		if strings.Contains(dbCompListString, v) {
			compList = append(compList, v)
		}
	}

	type innerRet struct {
		CompName string
		Template []string
		Plugin   []string
	}

	finalRet := []innerRet{}

	for _, comp_name := range compList {
		deploytype := f.DeployTypeV2{}
		if dt := db.Falcon.Where("comp_name = ?", comp_name).Find(&deploytype); dt.Error != nil {
			h.JSONR(c, expecstatus, fmt.Sprintf("The recoard not found with comp_name [%v] in the table deploytype!", comp_name))
			return
		}

		tplIDs := strings.Split(deploytype.TplIDs, ",")

		type template struct {
			Name string `json:"tpl_name" gorm:"column:tpl_name"`
		}

		type plugin struct {
			Name string `json:"plugin" gorm:"column:plugin"`
		}

		ret := innerRet{CompName: deploytype.CompName}

		for _, v := range tplIDs {
			q_tpl := template{}
			q_plugin := plugin{}
			dt := db.Falcon.Table("tpl").Select("tpl_name").Where("id = ?", v).Find(&q_tpl)
			if dt.Error != nil {
				config.Logger.Warningf("The recoard not found with tpl_id [%v] in the table tpl_name!", v)
				//h.JSONR(c, expecstatus, fmt.Sprintf("The recoard not found with tpl_id [%v] in the table tpl_name!", v))
				//return
			} else {
				ret.Template = append(ret.Template, q_tpl.Name)
			}

			dt = db.Falcon.Table("tpl_plugin").Where("tpl_id = ?", v).Find(&q_plugin)
			if dt.Error != nil {
				config.Logger.Warningf("The recoard not found with tpl_id [%v] in the table tpl_plugin!", v)
				//h.JSONR(c, expecstatus, fmt.Sprintf("The recoard not found with tpl_id [%v] in the table tpl_plugin!", v))
				//return
			} else {
				ret.Plugin = append(ret.Plugin, q_plugin.Name)
			}
		}

		finalRet = append(finalRet, ret)
	}

	h.JSONR(c, finalRet)
	return
}
