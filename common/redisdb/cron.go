package redisdb

import (
	"time"
)

func RedisMetric() {
	for {
		READ = WRITE
		WRITE = 0
		time.Sleep( 60 * time.Second )
	}
}