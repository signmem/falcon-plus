package falcon_portal

////////////////////////////////////////////////////////////////////////////
// | Field       | Type             | Null | Key | Default | Extra          |
// +-------------+------------------+------+-----+---------+----------------+
// | id          | int(10) unsigned | NO   | PRI | NULL    | auto_increment |
// | metric      | varchar(128)     | NO   |     |         |                |
// | tags        | varchar(256)     | NO   |     |         |                |
// | max_step    | int(11)          | NO   |     | 1       |                |
// | priority    | tinyint(4)       | NO   |     | 0       |                |
// | func        | varchar(16)      | NO   |     | all(#1) |                |
// | op          | varchar(8)       | NO   |     |         |                |
// | right_value | varchar(64)      | NO   |     | NULL    |                |
// | note        | varchar(128)     | NO   |     |         |                |
// | run_begin   | varchar(16)      | NO   |     |         |                |
// | run_end     | varchar(16)      | NO   |     |         |                |
// | tpl_id      | int(10) unsigned | NO   | MUL | 0       |                |
// | fid         | int(10) unsigned | NO   |     | 0       |                |
// +-------------+------------------+------+-----+---------+----------------+
////////////////////////////////////////////////////////////////////////////

type Strategy struct {
	ID         int64  `json:"id" gorm:"column:id"`
	Metric     string `json:"metric" gorm:"column:metric"`
	Tags       string `json:"tags" gorm:"column:tags"`
	MaxStep    int    `json:"max_step" gorm:"column:max_step"`
	Priority   int    `json:"priority" gorm:"column:priority"`
	Func       string `json:"func" gorm:"column:func"`
	Op         string `json:"op" gorm:"column:op"`
	RightValue string `json:"right_value" gorm:"column:right_value"`
	Note       string `json:"note" gorm:"column:note"`
	RunBegin   string `json:"run_begin" gorm:"column:run_begin"`
	RunEnd     string `json:"run_end" gorm:"column:run_end"`
	TplId      int64  `json:"tpl_id" gorm:"column:tpl_id"`
	FId        int64  `json:"fid" gorm:"column:fid"`
}

func (this Strategy) TableName() string {
	return "strategy"
}
