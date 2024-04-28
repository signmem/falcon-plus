package cmdb

type HostInfo struct {
	DomainName      string          `json:"domain"`
	HostName        string          `json:"hostname"`
	IPAddr          string          `json:"ipaddr"`
	RoomName		string			`json:"roomname"`
}


type CmdbTotalObject struct {
	Op     int    `json:"op"`        // 1 add, 2, update, 3 delete
	APIKey string `json:"api_key"`   //  /server/query    /app/query
	Object []struct {
		Name              string		`json:"name"`        // app name
		DeployType        int			`json:"deploy_type"` // deploy type, tomcat, web, web
		AppName           string		`json:"app_name"`
		ServerName        string		`json:"server_name"`
		Status            int			`json:"status"`  // 0,1,2,3,5,99,100  maintain   4  online
		Type              int			`json:"type"`    // 0 hypervisor , 1 vmware
		Ip 				  string		`json:"ip"`
		UseType           int			`json:"use_type"`  // 0 product,
		PoolID            int			`json:"pool_id"`   // deploy pool id
		BusGroupName 	  string 		`json:"bus_group_name"` // yewuzu
		OSType 			  string 		`json:"os_type"`
		RoomName		  string 		`json:"room_name"` // room
		Tag 			  string 		`json:"tag"`
	} `json:"object"`
	PreObject []struct {
		Name              string		`json:"name"`        // app name
		DeployType        int			`json:"deploy_type"` // deploy type, tomcat, web, web
		AppName           string		`json:"app_name"`
		ServerName        string		`json:"server_name"`
		Status            int			`json:"status"`  // 0,1,2,3,5,99,100  maintain   4  online
		Type              int			`json:"type"`    // 0 hypervisor , 1 vmware
		Ip 				  string		`json:"ip"`
		UseType           int			`json:"use_type"`  // 0 product,
		PoolID            int			`json:"pool_id"`   // deploy pool id
		BusGroupName 	  string 		`json:"bus_group_name"` // yewuzu
		OSType 			  string 		`json:"os_type"`
		RoomName		  string 		`json:"room_name"` // room
		Tag 			  string 		`json:"tag"`
	} `json:"pre_object"`
}