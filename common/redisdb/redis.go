package redisdb

import (
	"context"
	"errors"
	"github.com/gomodule/redigo/redis"
	"os"
	"os/signal"
	"syscall"
	"time"
)


var (
	WRITE = 0
	READ = 0
	ErrNil = errors.New("no matching record found in redis database")
	Ctx    = context.TODO()
	Pool *redis.Pool
	MaxIdle int
	MaxActive int
    IdleTimeOut int
    Server string
)

func NewPool(maxidle int, maxactive int,idletimeout int, server string) *redis.Pool {
    return &redis.Pool{
        MaxIdle:      maxidle,
        MaxActive:    maxactive,
        IdleTimeout:  time.Duration(idletimeout),

        Dial: func() (redis.Conn, error) {
            return redis.Dial("tcp", server)
        },

        TestOnBorrow: func(c redis.Conn, t time.Time) error {
            _, err := c.Do("PING")
            return err
        },
    }
}

func CleanupHook() {

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    signal.Notify(c, syscall.SIGTERM)
    signal.Notify(c, syscall.SIGKILL)
    go func() {
        <-c
        Pool.Close()
        os.Exit(0)
    }()
}

