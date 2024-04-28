package redisdb

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"errors"
)

func RedisServiceWrite(service string, hostname string) ( status bool, err error) {

	key := "/" + service + "/" + hostname
	timeNow := time.Now().Unix()
	timeNowStr := strconv.FormatInt(timeNow, 10)
	err = Set(key, timeNowStr)
	if err != nil {
		return false, err
	}
	return true, nil
}

func ReidsServerRead(service string, hostname string) (value []byte, err error) {
	key := "/" + service + "/" + hostname
	value, err = Get(key)
	return
}


func RedisServiceScan(service string) ( keys []string ,  err error) {
	key := "/" + service + "/*"
	values, err := GetKeys(key)

	if err != nil {
		return
	}

	if len(values) > 0 {
		for _, value := range values {
			repsting := "/" + service + "/"
			lastkey := strings.Replace(value, repsting, "", -1)
			keys = append(keys, lastkey)
		}
	}
	return
}

func RedisServiceExprieScan(service string, expire int64) ( normalHost []string, expireHost []string,
	alarm bool, err error) {

	key := "/" + service + "/*"
	allKeys, err := GetKeys(key)

	// alarm redis
	if err != nil {
		err_msg := fmt.Sprintf("RedisServiceExprieScan() get /agent.alive/* error: %s", err)
		log.Printf(err_msg)
		return normalHost, expireHost, true, errors.New(err_msg)
	}

	timeNow := time.Now().Unix()
	if len(allKeys) > 0 {
		for _, fullKey := range allKeys {
			value, err  := Get(fullKey)
			if err != nil {
				err_msg := fmt.Sprintf("RedisServiceExprieScan() get /agent.alive/%s error:%s", fullKey, err)
				log.Printf(err_msg)
				// alarm redis
				continue
			}
			oldTime, err := strconv.ParseInt(string(value), 10, 64)
			if err != nil {
				log.Printf("RedisServiceExprieScan() change into int64 err %s", err)
				continue
			}
			repsting := "/" + service + "/"
			lastkey := strings.Replace(fullKey, repsting, "", -1)

			if timeNow - oldTime > expire {
				expireHost = append(expireHost, lastkey)
				err  := Delete(lastkey)
				if err != nil {
					log.Printf("delete key %s error, %s", lastkey, err)
					continue
				}
			} else {
				normalHost = append(normalHost, lastkey)
			}
		}
	}

	return normalHost, expireHost,false, nil

}



func RedisServerDelete(service string, hostname string) error {
	key := "/" + service + "/" + hostname
	return Delete(key)
}

func RedisServiceExists(service string)  (status bool, err error) {
	key := "/" + service
	status, err = Exists(key)
	return
}

func RedisServerExists(service string, hostname string )  (status bool, err error) {
	key := "/" + service + "/" + hostname
	status, err = Exists(key)
	return
}