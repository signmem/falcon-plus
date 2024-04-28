package falcon_portal

// +-------------+---------------------+------+-----+-------------------+----------------+
// | Field       | Type                | Null | Key | Default           | Extra          |
// +-------------+---------------------+------+-----+-------------------+----------------+
// | id          | bigint(20) unsigned | NO   | PRI | NULL              | auto_increment |
// | cmdb_id     | int(10) unsigned    | NO   |     | 0                 |                |
// | tpl_ids     | varchar(255)        | NO   |     |                   |                |
// | create_time | timestamp           | NO   |     | CURRENT_TIMESTAMP |                |
// +-------------+---------------------+------+-----+-------------------+----------------+

type DeployType struct {
	ID     int64  `json:"id" gorm:"column:id"`
	CmdbID int64  `json:"cmdb_id" gorm:"column:cmdb_id"`
	TplIDs string `json:"tpl_ids" orm:"column:tpl_ids"`
}

type DeployTypeV2 struct {
	ID int64 `json:"id" gorm:"column:id"`
	//CmdbID int64  `json:"cmdb_id" gorm:"column:cmdb_id"`
	CompName string `json:"comp_name" orm:"column:comp_name"`
	TplIDs   string `json:"tpl_ids" orm:"column:tpl_ids"`
}

func (this DeployType) TableName() string {
	return "deploy_type"
}

func (this DeployTypeV2) TableName() string {
	return "deploy_type"
}
