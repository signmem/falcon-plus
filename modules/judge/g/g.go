package g

import (
	"log"
	"runtime"
)

// change log
// 2.0.1: bugfix HistoryData limit
// 2.0.2: clean stale data
// 2.0.3: add timeout to sync strategies and expressions and filter datapoint without strategies
// 2.0.4: 内存优化
const (
	VERSION = "2.0.4"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
