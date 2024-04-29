package host

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/signmem/falcon-plus/modules/api/app/utils"
	"github.com/signmem/falcon-plus/modules/api/config"
)

var db config.DBPool

const badstatus = http.StatusBadRequest
const expecstatus = http.StatusExpectationFailed

func Routes(r *gin.Engine) {
	db = config.Con()
	hostr := r.Group("/api/v1")
	hostr.Use(utils.AuthSessionMidd)

	hostr_vip := r.Group("/api/vip")
	hostr_vip.Use(utils.AuthSessionMidd)

	//hostgroup
	hostr.GET("/hostgroup", GetHostGroups)
	hostr.POST("/hostgroup", CrateHostGroup)
	hostr.POST("/hostgroup/host", BindHostToHostGroup)
	hostr.PUT("/hostgroup/host", UnBindAHostToHostGroup)
	hostr.GET("/hostgroup/:host_group", GetHostGroup)
	hostr.DELETE("/hostgroup/:host_group", DeleteHostGroup)
	hostr_vip.POST("/hostgroup/host", BindHostToHostGroupV2)
	hostr_vip.PUT("/hostgroup/host", UnBindAHostToHostGroupV2)
	hostr_vip.GET("/hostgroup/:host_group", GetHostGroupInfo)
	hostr_vip.DELETE("/hostgroup/:host_group", DeleteHostGroupV2)
	hostr_vip.GET("/hostgroups", GetHostGroupsWithTplPlugin)

	//plugins
	hostr.GET("/hostgroup/:host_group/plugins", GetPluginOfGrp)
	hostr.GET("/plugins", GetPlugins)
	hostr.POST("/plugin", CreatePlugin)
	hostr.DELETE("/plugin/:id", DeletePlugin)
	hostr_vip.POST("/plugin", CreatePluginV2) // bind plugin to hostgroup
	hostr_vip.PUT("/plugin", UnBindPluginToGrp)
	hostr_vip.GET("/tplplugin/:tpl_id", GetTplRelatedPlugin)
	hostr_vip.GET("/plugin/:tpl_name", GetTplRelatedPluginViaName)
	hostr_vip.GET("/tplplugin", GetAllTplRelatedPlugin)
	hostr_vip.POST("/tplplugin/insert", CreateTplPlugin)
	hostr_vip.PUT("/tplplugin/update", UpdateTplPlugin)
	hostr_vip.DELETE("/tplplugin/:tpl_id", DeleteTplPlugin)

	//aggreator
	hostr.GET("/hostgroup/:host_group/aggregators", GetAggregatorListOfGrp)
	hostr.GET("/aggregator/:id", GetAggregator)
	hostr.POST("/aggregator", CreateAggregator)
	hostr.PUT("/aggregator", UpdateAggregator)
	hostr.DELETE("/aggregator/:id", DeleteAggregator)

	//template
	hostr.POST("/hostgroup/template", BindTemplateToGroup)
	hostr.PUT("/hostgroup/template", UnBindTemplateToGroup)
	hostr.GET("/hostgroup/:host_group/template", GetTemplateOfHostGroup)
	hostr_vip.POST("/hostgroup/template", BindTemplateToGroupV2)
	hostr_vip.PUT("/hostgroup/template", UnBindTemplateToGroupV2)
	hostr_vip.GET("/hostgroup/:host_group/template", GetTemplateOfHostGroupV2)

	//host
	hostr.GET("/host/:host_id/template", GetTplsRelatedHost)
	hostr.GET("/host/:host_id/hostgroup", GetGrpsRelatedHost)
	hostr.POST("/host/maintain/set", SetMaintain)
	hostr.POST("/host/maintain/unset", UnsetMaintain)
	hostr_vip.GET("/hosts", GetHosts)
	hostr_vip.GET("/host/:hostname", GetHostIP)
	hostr_vip.GET("/host_list", GetHostList)
	hostr_vip.GET("/host_iplist", GetIPList)
	hostr_vip.DELETE("/host/:hostname", DeleteHost)
	hostr_vip.DELETE("/hosts", DeleteHosts)
	hostr_vip.POST("/hosts", AddHosts)
	hostr_vip.POST("/host", CheckHostBindHostGroup)

	//deploy type
	hostr_vip.GET("/deploytype/:cmdb_id", GetDeployType)
	hostr_vip.GET("/deploytype", GetAllDeployType)
	hostr_vip.POST("/deploytype/insert", CreateDeployType)
	hostr_vip.PUT("/deploytype/update", UpdateDeployType)
	hostr_vip.PUT("/deploytype", UpdateCmdbDeployType)
	hostr_vip.DELETE("/deploytype/:cmdb_id", DeleteDeployType)

	hostr_vip.GET("/deploytype2/:comp_name", GetDeployTypeV2)
	hostr_vip.GET("/deploytype2", GetAllDeployTypeV2)
	hostr_vip.POST("/cmdb_tplplugin", GetDeployTypeOfTplPlugin)
	hostr_vip.GET("/comp_tplplugin/:domain", GetDomainOfTplPlugin)
	hostr_vip.POST("/deploytype2/insert", CreateDeployTypeV2)
	hostr_vip.PUT("/deploytype2/update", UpdateDeployTypeV2)
	hostr_vip.DELETE("/deploytype2/:comp_name", DeleteDeployTypeV2)
}
