package g

import (
	"runtime"
)

const (
	VERSION         = "20220419"
	TREND_INTERVALS = 3600 //聚合区间为1小时
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
