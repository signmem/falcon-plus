package g

import (
	"bufio"
	"io/ioutil"
	"log"
	"io"
	"os"
	"regexp"
	"strings"
	"github.com/ShellCode33/VM-Detection/vmdetect"
	"encoding/json"
)

var (
	POOLNAME = "None"
	VirtualMachine = false
	SkipSwapMonitor = false    // default monitor swap usage -> skip = false
)

func InitPoolAndVMTags() {

	cmdbFile := Config().CmdbFile

	_, err := os.Stat(cmdbFile)
	if err != nil {    // go 1.6 not supper errors.Is function
		log.Printf("[ERROR] InitPoolTags() stat file error", err)
		return
	}

	f, err := os.Open(cmdbFile)
	if err != nil {
		log.Printf("[ERROR] InitPoolTags() open file error", err)
		return
	}
	defer f.Close()

	br  := bufio.NewReader(f)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		match , err := regexp.MatchString("pool_name", string(a))
		if err != nil {
			log.Printf("[ERROR] InitPoolTags() match string error", err)
			return
		}

		if match {
			pnames := strings.Split(string(a), "=")
			pname := strings.Split(pnames[1], ",")   // multi deploy pool
			POOLNAME = strings.Trim(pname[0], " ")
			return
		}

	}
	return
}

func CheckVirtualMachine() {
	isInsideVM, _ := vmdetect.IsRunningInVirtualMachine()

	if isInsideVM {
		VirtualMachine = true
	}
}


func CheckSwapMonitor()  {
	// notice : SkipSwapMonitor -> false ( do monitor swap )
	// notice : SkipSwapMonitor -> true  ( do not monitor swap )

	CheckVirtualMachine()
	virtualFile := Config().VMConfig

	_, err := os.Stat(virtualFile)
	if err != nil {
		log.Println("[info] CheckSwapMonitor() vmconfig file not exists")
		SkipSwapMonitor = false
		return
	}

	f, err := os.Open(virtualFile)
	if err != nil {
		log.Printf("[ERROR] CheckSwapMonitor() open file error: %s", err)
		SkipSwapMonitor = false
		return
	}
	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)
	var vmswapmonitor VMconfig
	err = json.Unmarshal([]byte(byteValue), &vmswapmonitor)

	if err != nil {
		log.Printf("[ERROR] CheckSwapMonitor() json decode error:%s ", err )
		SkipSwapMonitor = false
		return
	}

	SkipSwapMonitor = vmswapmonitor.SkipSwapMonitor
	VirtualMachine = vmswapmonitor.Virtual

	return
}