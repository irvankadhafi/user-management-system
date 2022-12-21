package cacher

import (
	"github.com/go-redsync/redsync/v4"
	redigosync "github.com/go-redsync/redsync/v4/redis/redigo"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/jpillora/backoff"
	"time"
)

const (
	// Override these when constructing the cache keeper
	defaultTTL          = 10 * time.Second
	defaultNilTTL       = 5 * time.Minute
	defaultLockDuration = 1 * time.Minute
	defaultLockTries    = 1
	defaultWaitTime     = 15 * time.Second
)

var nilValue = []byte("null")

type (
	CacheManager interface {
		Get(key string) (any, error)
		GetOrLock(key string) (any, *redsync.Mutex, error)
		StoreWithoutBlocking(Item) error
		StoreMultiWithoutBlocking([]Item) error
		DeleteByKeys([]string) error
		StoreNil(cacheKey string) error
		IncreaseCachedValueByOne(key string) error
		Expire(string, time.Duration) error

		GetTTL(string) (int64, error)

		AcquireLock(string) (*redsync.Mutex, error)
		SetDefaultTTL(time.Duration)
		SetNilTTL(time.Duration)
		SetConnectionPool(*redigo.Pool)
		SetLockConnectionPool(*redigo.Pool)
		SetLockDuration(time.Duration)
		SetLockTries(int)
		SetWaitTime(time.Duration)
		SetDisableCaching(bool)
	}
)

type cacheManager struct {
	connPool       *redigo.Pool
	nilTTL         time.Duration
	defaultTTL     time.Duration
	waitTime       time.Duration
	disableCaching bool

	lockConnPool *redigo.Pool
	lockDuration time.Duration
	lockTries    int
}

// NewCacheManager creates and returns a new cacheManager struct with default values for its fields.
func NewCacheManager() CacheManager {
	return &cacheManager{
		defaultTTL:     defaultTTL,
		nilTTL:         defaultNilTTL,
		lockDuration:   defaultLockDuration,
		lockTries:      defaultLockTries,
		waitTime:       defaultWaitTime,
		disableCaching: false,
	}
}

// Get retrieves an item with the given key from the cache.
// If the item is not present in the cache, or if there is an error retrieving it, the method returns nil and an error.
func (cm *cacheManager) Get(key string) (cachedItem any, err error) {
	if cm.disableCaching {
		return
	}

	cachedItem, err = get(cm.connPool.Get(), key)
	if err != nil && err != ErrKeyNotExist && err != redigo.ErrNil || cachedItem != nil {
		return
	}

	return nil, nil
}

// GetOrLock retrieves an item with the given key from the cache.
// If the item is not present in the cache, the method tries to acquire a lock on the key.
// If the lock is acquired, the method returns the item and the lock.
// If the lock is not acquired, the method waits for a short period of time before trying to acquire the lock again.
// This process is repeated until the lock is acquired or the elapsed time exceeds the waitTime field of the cacheManager struct.
// If the lock is not acquired within the specified time, the method returns nil and an error indicating that the wait time has been exceeded.
// The mutex returned by the method is a pointer to a redsync.Mutex struct, which represents a distributed mutex.
func (cm *cacheManager) GetOrLock(key string) (cachedItem any, mutex *redsync.Mutex, err error) {
	if cm.disableCaching {
		return
	}

	cachedItem, err = get(cm.connPool.Get(), key)
	if err != nil && err != ErrKeyNotExist && err != redigo.ErrNil || cachedItem != nil {
		return
	}

	mutex, err = cm.AcquireLock(key)
	if err == nil {
		return
	}

	start := time.Now()
	for {
		b := &backoff.Backoff{
			Min:    20 * time.Millisecond,
			Max:    200 * time.Millisecond,
			Jitter: true,
		}

		if !cm.isLocked(key) {
			cachedItem, err = get(cm.connPool.Get(), key)
			if err != nil {
				if err == ErrKeyNotExist {
					mutex, err = cm.AcquireLock(key)
					if err == nil {
						return nil, mutex, nil
					}

					goto Wait
				}
				return nil, nil, err
			}
			return cachedItem, nil, nil
		}

	Wait:
		elapsed := time.Since(start)
		if elapsed >= cm.waitTime {
			break
		}

		time.Sleep(b.Duration())
	}

	return nil, nil, ErrWaitTooLong
}

// IncreaseCachedValueByOne will increments the number stored at key by one.
// If the key does not exist, it is set to 0 before performing the operation
func (cm *cacheManager) IncreaseCachedValueByOne(key string) error {
	if cm.disableCaching {
		return nil
	}

	client := cm.connPool.Get()
	defer func() {
		_ = client.Close()
	}()

	_, err := client.Do("INCR", key)
	return err
}

// Expire Set expire a key
func (cm *cacheManager) Expire(key string, duration time.Duration) (err error) {
	if cm.disableCaching {
		return nil
	}

	client := cm.connPool.Get()
	defer func() {
		_ = client.Close()
	}()

	_, err = client.Do("EXPIRE", key, int64(duration.Seconds()))
	return
}

func (cm *cacheManager) GetTTL(name string) (value int64, err error) {
	client := cm.connPool.Get()
	defer func() {
		_ = client.Close()
	}()

	val, err := client.Do("TTL", name)
	if err != nil {
		return
	}

	value = val.(int64)
	return
}

// StoreWithoutBlocking stores an item in the cache without blocking.
// the method stores the item in the cache using the SETEX Redis command.
// The cacheTTL for the item is determined by the decideCacheTTL method.
func (cm *cacheManager) StoreWithoutBlocking(c Item) error {
	if cm.disableCaching {
		return nil
	}

	client := cm.connPool.Get()
	defer func() {
		_ = client.Close()
	}()

	_, err := client.Do("SETEX", c.GetKey(), cm.decideCacheTTL(c), c.GetValue())
	return err
}

// StoreMultiWithoutBlocking stores multiple items in the cache without blocking.
// the method uses the MULTI and EXEC Redis commands to store the items in the cache using the SETEX command.
// The cacheTTL for each item is determined by the decideCacheTTL method.
func (cm *cacheManager) StoreMultiWithoutBlocking(items []Item) error {
	if cm.disableCaching {
		return nil
	}

	client := cm.connPool.Get()
	defer func() {
		_ = client.Close()
	}()

	err := client.Send("MULTI")
	if err != nil {
		return err
	}
	for _, item := range items {
		err = client.Send("SETEX", item.GetKey(), cm.decideCacheTTL(item), item.GetValue())
		if err != nil {
			return err
		}
	}

	_, err = client.Do("EXEC")
	return err
}

// DeleteByKeys deletes items from the cache with the given keys.
// the method uses the DEL Redis command to delete the items from the cache.
func (cm *cacheManager) DeleteByKeys(keys []string) error {
	if cm.disableCaching {
		return nil
	}

	if len(keys) <= 0 {
		return nil
	}

	client := cm.connPool.Get()
	defer func() {
		_ = client.Close()
	}()

	var redisKeys []any
	for _, key := range keys {
		redisKeys = append(redisKeys, key)
	}

	_, err := client.Do("DEL", redisKeys...)
	return err
}

// AcquireLock tries to acquire a lock on the given key.
// The lock is implemented using a distributed mutex.
// The method returns a pointer to the mutex and an error indicating whether the lock was successfully acquired.
func (cm *cacheManager) AcquireLock(key string) (*redsync.Mutex, error) {
	p := redigosync.NewPool(cm.lockConnPool)
	r := redsync.New(p)
	m := r.NewMutex("lock:"+key,
		redsync.WithExpiry(cm.lockDuration),
		redsync.WithTries(cm.lockTries))

	return m, m.Lock()
}

// SetDefaultTTL sets the defaultTTL field of the cacheManager struct to the given duration.
// The defaultTTL field specifies the default time-to-live (TTL) for items in the cache.
func (cm *cacheManager) SetDefaultTTL(d time.Duration) {
	cm.defaultTTL = d
}

// SetNilTTL sets the nilTTL field of the cacheManager struct to the given duration.
// The nilTTL field specifies the TTL for items that have a value of nil.
func (cm *cacheManager) SetNilTTL(d time.Duration) {
	cm.nilTTL = d
}

// SetLockConnectionPool sets the lockConnPool field of the cacheManager struct to the given connection pool.
// The lockConnPool field is a connection pool for a Redis cache used for implementing locks.
func (cm *cacheManager) SetConnectionPool(c *redigo.Pool) {
	cm.connPool = c
}

// SetLockConnectionPool sets the lockConnPool field of the cacheManager struct to the given connection pool.
// The lockConnPool field is a connection pool for a Redis cache used for implementing locks.
func (cm *cacheManager) SetLockConnectionPool(c *redigo.Pool) {
	cm.lockConnPool = c
}

// SetLockDuration sets the lockDuration field of the cacheManager struct to the given duration.
// The lockDuration field specifies the duration for which a lock should be held when acquiring a lock.
func (cm *cacheManager) SetLockDuration(d time.Duration) {
	cm.lockDuration = d
}

// SetLockTries sets the lockTries field of the cacheManager struct to the given number of tries.
// The lockTries field specifies the number of times the lock should be tried when acquiring a lock.
func (cm *cacheManager) SetLockTries(t int) {
	cm.lockTries = t
}

// SetWaitTime sets the waitTime field of the cacheManager struct to the given duration.
// The waitTime field specifies the maximum amount of time to wait when trying to acquire a lock.
func (cm *cacheManager) SetWaitTime(d time.Duration) {
	cm.waitTime = d
}

// SetDisableCaching sets the disableCaching field of the cacheManager struct to the given boolean value.
// If disableCaching is true, caching is disabled and the cacheManager methods do nothing and return nil.
// If disableCaching is false, caching is enabled and the cacheManager methods perform their normal operations.
func (cm *cacheManager) SetDisableCaching(b bool) {
	cm.disableCaching = b
}

// StoreNil stores a nil value in the cache with the given key and the nilTTL specified in the cacheManager struct.
// If the disableCaching field of the cacheManager struct is true, the method does nothing and returns nil.
// Otherwise, the method stores the nil value in the cache using the StoreWithoutBlocking method.
func (cm *cacheManager) StoreNil(cacheKey string) error {
	item := NewItemWithCustomTTL(cacheKey, nilValue, cm.nilTTL)
	err := cm.StoreWithoutBlocking(item)
	return err
}

func (cm *cacheManager) decideCacheTTL(c Item) (ttl int64) {
	if ttl = c.GetTTLInt64(); ttl > 0 {
		return
	}

	return int64(cm.defaultTTL.Seconds())
}

func (cm *cacheManager) isLocked(key string) bool {
	client := cm.lockConnPool.Get()
	defer func() {
		_ = client.Close()
	}()

	reply, err := client.Do("GET", "lock:"+key)
	if err != nil || reply == nil {
		return false
	}

	return true
}
