package g

import (
	"math/rand"
	"strconv"
	"time"
)

type TimeStruct struct {
	Hour   string
	Minite string
}

type RsyncResponse struct {
	Server       []string   `json:"server"`
	Source       string     `json:"source"`
	Sync         bool       `json:"sync"`
	SyncSrc      string     `json:"syncsrc"`
	SyncPort     string     `json:"port"`
	SyncDest     string     `json:"syncdest"`
	SyncDelete   bool       `json:"syncdelete"`
	Version      string     `json:"version"`
	RpmUpdate    bool       `json:"rpmupdate"`
	RpmVersion   struct {
		El5      string     `json:"el5"`
		El6      string     `json:"el6"`
		El7      string     `json:"el7"`
		El8      string     `json:"el8"`
	}                       `json:"rpmversion"`
}

var JobTime TimeStruct

func GetNow() TimeStruct {
	var nowTime TimeStruct
	hour, _   := strconv.Atoi(time.Now().Format("15"))  // meaning get hour
	minite, _ := strconv.Atoi(time.Now().Format("04"))  // meaning get minite
	if hour < 10 {
		nowTime.Hour = "0" + strconv.Itoa(hour)
	} else {
		nowTime.Hour = strconv.Itoa(hour)
	}
	if minite < 10 {
		nowTime.Minite = "0" + strconv.Itoa(minite)
	} else {
		nowTime.Minite = strconv.Itoa(minite)
	}

	return nowTime
}

func GenCronTime() {
	rand.Seed(int64(time.Now().UnixNano()))
	hour := rand.Intn(18 - 11) + 11  // meaning 11 ~ 17
	rand.Seed(int64(time.Now().UnixNano()))
	minite := rand.Intn(59 - 0)      // meaning 0 ~ 59

	if hour < 10 {
		JobTime.Hour = "0" + strconv.Itoa(hour)
	}else {
		JobTime.Hour = strconv.Itoa(hour)
	}
	if minite < 10 {
		JobTime.Minite = "0" + strconv.Itoa(minite)
	} else {
		JobTime.Minite = strconv.Itoa(minite)
	}
	// JobTime.Hour = "16"
	// JobTime.Minite = "03"   // use for test
}