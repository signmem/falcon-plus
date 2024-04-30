package redisdb

import (
	"context"
	"github.com/signmem/redislock"
	"time"
)


func LockRedis(client *redislock.Client, ctx context.Context, lock_key string,
	lock_time int) (lock *redislock.Lock ,err error) {

	lock, err = client.Obtain(ctx, lock_key,
		time.Duration(lock_time) * time.Second, nil)

	if err != nil {
		return lock, err
	}

	return lock, nil
}

func ReleaseLockRedis(lock *redislock.Lock, ctx context.Context) {
	defer lock.Release(ctx)
}