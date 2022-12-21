package db

import (
	"errors"
	goredis "github.com/go-redis/redis/v8"
	redigo "github.com/gomodule/redigo/redis"
	"time"
)

// RedisConnectionPoolOptions options for the redis connection
type RedisConnectionPoolOptions struct {
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleCount       int
	PoolSize        int
	IdleTimeout     time.Duration
	MaxConnLifetime time.Duration
}

// NewRedigoRedisConnectionPool uses redigo library to establish the redis connection pool
func NewRedigoRedisConnectionPool(url string, opt *RedisConnectionPoolOptions) (*redigo.Pool, error) {
	if !isValidRedisStandaloneURL(url) {
		return nil, errors.New("invalid redis URL: " + url)
	}

	return &redigo.Pool{
		MaxIdle:     opt.IdleCount,
		MaxActive:   opt.PoolSize,
		IdleTimeout: opt.IdleTimeout,
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.DialURL(url)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		MaxConnLifetime: opt.MaxConnLifetime,
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Wait: true, // wait for connection available when maxActive is reached
	}, nil
}

func isValidRedisStandaloneURL(url string) bool {
	_, err := goredis.ParseURL(url)
	return err == nil
}
