package cacher

import (
	"github.com/go-redsync/redsync/v4"
	redigo "github.com/gomodule/redigo/redis"
)

func get(client redigo.Conn, key string) (value any, err error) {
	defer func() {
		_ = client.Close()
	}()

	err = client.Send("MULTI")
	if err != nil {
		return nil, err
	}
	err = client.Send("EXISTS", key)
	if err != nil {
		return nil, err
	}
	err = client.Send("GET", key)
	if err != nil {
		return nil, err
	}
	res, err := redigo.Values(client.Do("EXEC"))
	if err != nil {
		return nil, err
	}

	val, ok := res[0].(int64)
	if !ok || val <= 0 {
		return nil, ErrKeyNotExist
	}

	return res[1], nil
}

// SafeUnlock safely unlock mutex
func SafeUnlock(mutex *redsync.Mutex) {
	if mutex != nil {
		_, _ = mutex.Unlock()
	}
}
