package g

import (
	"runtime"
)

// change log 0.0.4
// fix metric == cpu.idle push to transfer without tags!!

const (
	VERSION = "0.0.5"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
