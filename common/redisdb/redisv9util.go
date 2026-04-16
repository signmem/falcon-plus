package redisdb

import (
	"context"
	redis9 "github.com/redis/go-redis/v9"
)

func GKeys(client *redis9.Client, path string) (keys []string, err error) {

	ctx := context.Background()

	rkeys := client.Keys(ctx, path)

	return rkeys.Val(), rkeys.Err()
}

func GMGets(client *redis9.Client, keys []string) (values []interface{}, err error) {

	ctx := context.Background()
	rvalue := client.MGet(ctx, keys...)
	return rvalue.Val(), rvalue.Err()
}

func GScans(client *redis9.Client, path string) (keys []string, err error) {
	ctx := context.Background()
	iter := client.Scan(ctx,0, path, 0).Iterator()

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	err = iter.Err()
	return keys, err
}

func GGet(client *redis9.Client, key string) (value []byte, err error) {
	ctx := context.Background()
	value, err = client.Get(ctx, key).Bytes()
	return
}

func GGetByte(client *redis9.Client, key string) (value []byte, err error) {
	ctx := context.Background()
	value, err = client.Get(ctx, key).Bytes()
	return
}

func GGetString(client *redis9.Client, key string) (value string, err error) {
	ctx := context.Background()
	value, err = client.Get(ctx, key).Result()
	return
}

func GExists(client *redis9.Client, key string) (bool) {
	ctx := context.Background()
	status, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false
	}
	if status > 0 {
		return true
	}
	return false
}


func GDelete(client *redis9.Client, key string) (err error) {
	ctx := context.Background()
	err = client.Del(ctx, key).Err()
	return
}


func GSet(client *redis9.Client, key string, value string) (err error) {
	ctx := context.Background()
	err = client.Set(ctx, key, value, 0).Err()
	return
}