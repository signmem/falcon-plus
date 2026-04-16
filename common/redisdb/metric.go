package redisdb

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func RedisServiceWrite(service string, hostname string) ( status bool, err error) {

	key := "/" + service + "/" + hostname
	timeNow := time.Now().Unix()
	timeNowStr := strconv.FormatInt(timeNow, 10)
	err = GSet(Client, key, timeNowStr)
	if err != nil {
		return false, err
	}
	return true, nil
}

func ReidsServerRead(service string, hostname string) (value []byte, err error) {
	key := "/" + service + "/" + hostname
	value, err = GGet(Client, key)
	return
}


func RedisServiceScan(service string) ( keys []string ,  err error) {
	redisPath := "/" + service + "/*"

	// values, err := GetKeys(key)

	allKeys, err := GScans(Client, redisPath)

	if err != nil {
		return
	}

	if len(allKeys) > 0 {
		for _, value := range allKeys {
			repsting := "/" + service + "/"
			lastkey := strings.Replace(value, repsting, "", -1)
			keys = append(keys, lastkey)
		}
	}
	return
}

func RedisServiceExprieScan(service string, expire int64) (normalHost []string,
	expireHost []string, alarm bool, err error) {

	var err_msg string

	redisPath := "/" + service + "/*"
	allKeys, err := GScans(Client, redisPath)

	// alarm redis
	if err != nil {
		err_msg = fmt.Sprintf("RedisServiceExprieScan() RKeys /%s/* error: %s", service, err)
		return normalHost, expireHost, true, errors.New(err_msg)
	}

	if len(allKeys) == 0 {
		return normalHost, expireHost, false, nil
	}

	allValue, err := GMGets(Client, allKeys)

	// alarm redis
	if err != nil {
		err_msg = fmt.Sprintf("RedisServiceExprieScan() MGets /%s/* error: %s", service, err)
		return normalHost, expireHost, true, errors.New(err_msg)
	}

	timeNow := time.Now().Unix()

	if len(allKeys) == 0 || len(allValue) != len(allKeys) {
		err_msg = fmt.Sprintf("RedisServiceExprieScan() Check /%s/* key:%d, value:%d, error: %s",
			len(allKeys), len(allValue), service, err)
		return normalHost, expireHost, false, errors.New(err_msg)
	}

	for num := 0; num < len(allKeys); num ++ {

		var timeStamp string

		timeStamp, ok  := allValue[num].(string)
		if ! ok {
			timeStamp = "0"
		}
		
		timeInt, err := strconv.ParseInt(timeStamp, 10, 64)
		if err != nil {
			continue
		}

		repsting := "/" + service + "/"
		hostname := strings.Replace(allKeys[num], repsting, "", -1)

		if timeNow - timeInt >  expire {

			expireHost = append(expireHost, hostname)
		}  else {
			normalHost = append(normalHost, hostname)
		}
	}


	/*
	//  old logical .....

	if len(allKeys) > 0 {
		for _, fullKey := range allKeys {
			value, err  := Get(fullKey)
			if err != nil {
				err_msg += fmt.Sprintf("RedisServiceExprieScan() get %s error:%s", fullKey, err)
				// alarm redis
				continue
			}
			oldTime, err := strconv.ParseInt(string(value), 10, 64)
			if err != nil {
				err_msg += fmt.Sprintf("RedisServiceExprieScan() change into int64 err %s", err)
				continue
			}
			repsting := "/" + service + "/"
			lastkey := strings.Replace(fullKey, repsting, "", -1)

			if timeNow - oldTime > expire {
				expireHost = append(expireHost, lastkey)
			} else {
				normalHost = append(normalHost, lastkey)
			}
		}
	}
	*/

	return normalHost, expireHost,false, nil

}



func RedisServerDelete(service string, hostname string) error {
	key := "/" + service + "/" + hostname
	return GDelete(Client, key)
}

func RedisServiceExists(service string)  (status bool) {
	key := "/" + service
	status = GExists(Client, key)
	return
}

func RedisServerExists(service string, hostname string )  (status bool) {
	key := "/" + service + "/" + hostname
	status = GExists(Client, key)
	return
}