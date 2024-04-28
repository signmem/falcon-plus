package g

import (
	"fmt"
	"log"
	"os"
)


func InitLog()  {
	logfile := Config().LogFile
	loggerFile, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println()
		panic(err)
	}
	log.SetOutput(loggerFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Ltime)
}
