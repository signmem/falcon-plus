package redisdb

import (
	"github.com/gsm/redislock"
	redis9 "github.com/redis/go-redis/v9"
	"context"
	"time"
	"errors"
)



func LockRedis(client *redis9.Client, lock_key string, lock_time int) (lock *redislock.Lock ,err error) {
	defer client.Close()
	ctx := context.Background()
	locker := redislock.New(client)
	lock, err = locker.Obtain(ctx, lock_key, time.Duration(lock_time) * time.Second, nil)

	if err != nil {
		return lock, err
	}

	return lock, nil
}

func ReleaseLockRedis(lock *redislock.Lock) (err error) {
	ctx := context.Background()
	err = lock.Release(ctx)
	return err
}

func FlushLockRedis(lock *redislock.Lock, lock_time int) (status bool, err error) {
	ctx := context.Background()
	ttl, err := lock.TTL(ctx)

	if err != nil {
		return false, err
	}

	if ttl > 0 {
		err = lock.Refresh(ctx, time.Duration(lock_time) * time.Second, nil )

		if err != nil {
			return false, err
		}
	} else {
		err = errors.New("")
		return false, err
	}
	return true, nil

}