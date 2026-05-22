package selector

import (
	"context"
	"github.com/signmem/falcon-plus/common/redisdb"
	"github.com/signmem/falcon-plus/modules/status-collectd/g"
	"github.com/signmem/redislock"
	"sync"
	"time"
)

var (
	Role string
	selectorLock sync.RWMutex
	lock *redislock.Lock
)


func Start() {

	config := g.Config().Redis
	lockKey := config.LockKey
	lockTime := config.LockTime
	server := config.Server + ":" + config.Port
	maxActive := config.MaxActive
	maxIdle := config.MaxIdle
	askLockTime := config.AskLockTime

	// 安全设置初始角色
	selectorLock.Lock()
	Role = "slave"
	selectorLock.Unlock()

	redisClient, err := redisdb.RedisClient(maxIdle, maxActive, server)

	if err != nil {
		g.Logger.Errorf("[SELECTOR] connect redis error:%s", err)
		return
	}

	defer func() {
		redisClient.Close()
	}()

	ctx := context.Background()
	locker := redislock.New(redisClient)

	for {
		time.Sleep(500 * time.Millisecond)

		selectorLock.RLock()
		currentRole := Role
		selectorLock.RUnlock()

		if currentRole == "slave" {

			newLock, err := redisdb.LockRedis(locker, ctx, lockKey, lockTime)

			if err != nil {
				selectorLock.Lock()
				Role = "slave"
				selectorLock.Unlock()
				time.Sleep(time.Duration(askLockTime) * time.Second)
				continue
			}

			g.Logger.Debugf("[SELECTOR CHECK] role change, now is: %s", Role)

			lock = newLock
			// 切换为 master
			selectorLock.Lock()
			Role = "master"
			selectorLock.Unlock()
			g.Logger.Debugf("[SELECTOR] become master successfully")
			time.Sleep( 1 * time.Second)

		} else {

			// ========== master 角色：续约锁 ==========
			if lock == nil {
				// 空锁保护，防止 panic
				g.Logger.Error("[SELECTOR] lock is nil, switch to slave")
				selectorLock.Lock()
				Role = "slave"
				selectorLock.Unlock()
				continue
			}

			// 续约锁
			err = lock.Refresh(ctx, time.Duration(lockTime) * time.Second, nil)
			if err != nil {
				g.Logger.Errorf("[SELECTOR] refresh lock failed: %s", err)

				// 续约失败，主动降级为 slave
				selectorLock.Lock()
				Role = "slave"
				selectorLock.Unlock()

				g.Logger.Debugf("[SELECTOR CHECK] role change, now is: %s", Role)
				time.Sleep( 1 * time.Second)
				continue
			}
		}
	}
}
