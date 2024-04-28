package config

import (
	"runtime"
)

// change log:
//  20220419
//  change api log format  
//  add logrotate function
//

const (
	VERSION = "20220419"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
