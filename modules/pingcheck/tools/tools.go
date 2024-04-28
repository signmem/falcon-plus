package tools

import (
	"strconv"
	"strings"
	"time"
)

func DeleteStrSliceEle(items []string, item string) []string {
	newitems := []string{}

	for _, i := range items {
		if i != item {
			newitems = append(newitems, i)
		}
	}

	return newitems
}

func GenHostNameSlice(sliceKey []string, path string ) (hosts []string) {
	if len(sliceKey) == 0 {
		return hosts
	}
	removePath := path + "/"
	for _, fullPath := range sliceKey {
		host := strings.Replace(fullPath, removePath, "",  -1 )
		hosts = append( hosts, host )
	}
	return hosts
}

func SliceContains(str string, s []string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}


func GetDstSliceNotInSrcSlice(src []string, dst []string) (diffSlice []string) {

	// get dstString  from not in srcString

	if len(src) == 0 {
		return dst
	}

	if len(dst) == 0 {
		return diffSlice
	}

	for  _, dstValue := range dst {
		match := false
		for  _, srcValue := range src {
			if dstValue == srcValue  {
				match = true
			}
		}

		if match == false {
			diffSlice = append(diffSlice, dstValue)
		}
	}
	return diffSlice

}

func IntInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}


func GetNow() (nowTime TimeStruct) {
	// use to get hour and min
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


type TimeStruct struct {
	Hour   string
	Minite string
}
