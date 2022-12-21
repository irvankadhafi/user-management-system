package repository

import (
	"gorm.io/driver/postgres"
	"os"
	"strconv"
	"testing"
	"time"
	"user-service/cacher"
	"user-service/internal/db"
	"user-service/internal/model/mock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis"
	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/golang/mock/gomock"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"user-service/internal/config"
)

func initializeTest() {
	config.GetConf()
	setupLogger()
}

func setupLogger() {
	formatter := runtime.Formatter{
		ChildFormatter: &log.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		},
		Line: true,
		File: true,
	}

	log.SetFormatter(&formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)

	verbose, _ := strconv.ParseBool(os.Getenv("VERBOSE"))
	if verbose {
		log.SetLevel(log.DebugLevel)
	}
}

func initializeCockroachMockConn() (db *gorm.DB, mock sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	db, err = gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	return
}

type repoTestKit struct {
	miniredis    *miniredis.Miniredis
	dbmock       sqlmock.Sqlmock
	db           *gorm.DB
	cacheManager cacher.CacheManager
	ctrl         *gomock.Controller
	mockUserRepo *mock.MockUserRepository
}

func initializeRepoTestKit(t *testing.T) (kit *repoTestKit, close func()) {
	mr, _ := miniredis.Run()
	r, err := newRedisConnPool("redis://" + mr.Addr())
	require.NoError(t, err)

	k := cacher.NewCacheManager()
	k.SetDisableCaching(false)
	k.SetConnectionPool(r)
	k.SetLockConnectionPool(r)
	k.SetWaitTime(1 * time.Second) // override wait time to 1 second

	dbconn, dbmock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: dbconn}), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	userRepo := mock.NewMockUserRepository(ctrl)
	tk := &repoTestKit{
		cacheManager: k,
		miniredis:    mr,
		ctrl:         ctrl,
		dbmock:       dbmock,
		db:           gormDB,
		mockUserRepo: userRepo,
	}

	close = func() {
		if conn, _ := tk.db.DB(); conn != nil {
			_ = conn.Close()
		}
		tk.miniredis.Close()
	}

	return tk, close
}

func newRedisConnPool(url string) (*redis.Pool, error) {
	redisOpts := &db.RedisConnectionPoolOptions{
		DialTimeout:     config.RedisDialTimeout(),
		ReadTimeout:     config.RedisReadTimeout(),
		WriteTimeout:    config.RedisWriteTimeout(),
		IdleCount:       config.RedisMaxIdleConn(),
		PoolSize:        config.RedisMaxActiveConn(),
		IdleTimeout:     240 * time.Second,
		MaxConnLifetime: 1 * time.Minute,
	}

	c, err := db.NewRedigoRedisConnectionPool(url, redisOpts)
	if err != nil {
		return nil, err
	}

	return c, nil
}
