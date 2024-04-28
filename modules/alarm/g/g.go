package g

import (
	"log"
	"runtime"
)

const (
	VERSION  = "1.0.2"
	CMDB_APP = "VipFalcon"
	CMDB_KEY = "329a55d8ddff6359c63eccb52c611140"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
