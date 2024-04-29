package redisdb

import (
    "fmt"
    "github.com/gomodule/redigo/redis"

)

func Ping() error {

    conn := Pool.Get()
    defer conn.Close()

    _, err := redis.String(conn.Do("PING"))
    if err != nil {
        return fmt.Errorf("cannot 'PING' db: %v", err)
    }
    return nil
}


func Get(key string) ([]byte, error) {
	// 获取某个 key 的值

    conn := Pool.Get()
    defer conn.Close()

    var data []byte
    data, err := redis.Bytes(conn.Do("GET", key))
    if err != nil {
        return data, fmt.Errorf("error getting key %s: %v", key, err)
    }
    return data, err
}

func Set(key string, value string) error {
	// 设定某个 KEY 值

    conn := Pool.Get()
    defer conn.Close()

    _, err := conn.Do("SET", key, value)

    return err
}

func Exists(key string) (bool, error) {
	// 校验某个 key 是否存在

    conn := Pool.Get()
    defer conn.Close()

    ok, err := redis.Bool(conn.Do("EXISTS", key))
    if err != nil {
        return ok, fmt.Errorf("error checking if key %s exists: %v", key, err)
    }
    return ok, err
}

func Delete(key string) error {
	// 删除 key 

    conn := Pool.Get()
    defer conn.Close()

    _, err := conn.Do("DEL", key)
    return err
}

func GetKeys(pattern string) ([]string, error) {
	// pattern = 目录
	// 返回目录下的所有 key name

    conn := Pool.Get()
    defer conn.Close()

    iter := 0
    keys := []string{}
    for {
        arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
        if err != nil {
            return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
        }

        iter, _ = redis.Int(arr[0], nil)
        k, _ := redis.Strings(arr[1], nil)
        keys = append(keys, k...)

        if iter == 0 {
            break
        }
    }

    return keys, nil
}

func Incr(counterKey string) (int, error) {

    conn := Pool.Get()
    defer conn.Close()

    return redis.Int(conn.Do("INCR", counterKey))
}

