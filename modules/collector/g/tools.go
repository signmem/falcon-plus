package g

import (
	"github.com/open-falcon/falcon-plus/common/redisdb"
)


func GetRedisServer(serviceType string) (nornalHost []string) {
	Logger.Infof("GetRedisServer() start: %s", serviceType)
	nornalHost, expireHost, _, err := redisdb.RedisServiceExprieScan(serviceType, 60)
	if err != nil {
		Logger.Errorf("GetRedisServer() err:%s", err)
	}
	Logger.Infof("GetRedisServer() info: noralHost %d, expirehost %d", len(nornalHost), len(expireHost))
	return nornalHost
}

func StringSliceMerge(sliceA []string, sliceB []string) ([]string) {
	if len(sliceA) == 0 {
		return sliceB
	}

	if len(sliceB) == 0 {
		return sliceA
	}

	for _, keyA := range sliceA {
		if contains(sliceB, keyA) == false {
			sliceB = append(sliceB, keyA)
		}
	}

	return sliceB
}



func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
