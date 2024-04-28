package falcon_portal

// +-------------+---------------------+------+-----+-------------------+----------------+
// | Field       | Type                | Null | Key | Default           | Extra          |
// +-------------+---------------------+------+-----+-------------------+----------------+
// | id          | bigint(20) unsigned | NO   | PRI | NULL              | auto_increment |
// | tpl_id      | int(10) unsigned    | NO   | MUL | 0                 |                |
// | plugin      | varchar(255)        | NO   |     | NULL              |                |
// | create_time | timestamp           | NO   |     | CURRENT_TIMESTAMP |                |
// +-------------+---------------------+------+-----+-------------------+----------------+

type TplPlugin struct {
	ID     int64  `json:"id" gorm:"column:id"`
	TplID  int64  `json:"tpl_id" gorm:"column:tpl_id"`
	Plugin string `json:"plugin" gorm:"column:plugin"`
}

func (this TplPlugin) TableName() string {
	return "tpl_plugin"
}
