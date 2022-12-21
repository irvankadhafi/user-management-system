package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

// GetConf :nodoc:
func GetConf() {
	viper.AddConfigPath(".")
	viper.AddConfigPath("./..")
	viper.AddConfigPath("./../..")
	viper.SetConfigName("config")

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Warningf("%v", err)
	}
}

// Env :nodoc:
func Env() string {
	return viper.GetString("env")
}

// LogLevel :nodoc:
func LogLevel() string {
	return viper.GetString("log_level")
}

// HTTPPort :nodoc:
func HTTPPort() string {
	return viper.GetString("ports.http")
}

// DisableCaching :nodoc:
func DisableCaching() bool {
	return viper.GetBool("disable_caching")
}

// DatabaseDSN :nodoc:
func DatabaseDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		DatabaseUsername(),
		DatabasePassword(),
		DatabaseHost(),
		DatabaseName(),
		DatabaseSSLMode())
}

// DatabaseHost :nodoc:
func DatabaseHost() string {
	return viper.GetString("postgres.host")
}

// DatabaseName :nodoc:
func DatabaseName() string {
	return viper.GetString("postgres.database")
}

// DatabaseUsername :nodoc:
func DatabaseUsername() string {
	return viper.GetString("postgres.username")
}

// DatabasePassword :nodoc:
func DatabasePassword() string {
	return viper.GetString("postgres.password")
}

// DatabaseSSLMode :nodoc:
func DatabaseSSLMode() string {
	if viper.IsSet("postgres.sslmode") {
		return viper.GetString("postgres.sslmode")
	}
	return "disable"
}

// DatabasePingInterval :nodoc:
func DatabasePingInterval() time.Duration {
	if viper.GetInt("postgres.ping_interval") <= 0 {
		return DefaultDatabasePingInterval
	}
	return time.Duration(viper.GetInt("postgres.ping_interval")) * time.Millisecond
}

// DatabaseRetryAttempts :nodoc:
func DatabaseRetryAttempts() float64 {
	if viper.GetInt("postgres.retry_attempts") > 0 {
		return float64(viper.GetInt("postgres.retry_attempts"))
	}
	return DefaultDatabaseRetryAttempts
}

// DatabaseMaxIdleConns :nodoc:
func DatabaseMaxIdleConns() int {
	if viper.GetInt("postgres.max_idle_conns") <= 0 {
		return DefaultDatabaseMaxIdleConns
	}
	return viper.GetInt("postgres.max_idle_conns")
}

// DatabaseMaxOpenConns :nodoc:
func DatabaseMaxOpenConns() int {
	if viper.GetInt("postgres.max_open_conns") <= 0 {
		return DefaultDatabaseMaxOpenConns
	}
	return viper.GetInt("postgres.max_open_conns")
}

// DatabaseConnMaxLifetime :nodoc:
func DatabaseConnMaxLifetime() time.Duration {
	if !viper.IsSet("postgres.conn_max_lifetime") {
		return DefaultDatabaseConnMaxLifetime
	}
	return time.Duration(viper.GetInt("postgres.conn_max_lifetime")) * time.Millisecond
}

// RedisDialTimeout :nodoc:
func RedisDialTimeout() time.Duration {
	cfg := viper.GetString("redis.dial_timeout")
	return parseDuration(cfg, 5*time.Second)
}

// RedisWriteTimeout :nodoc:
func RedisWriteTimeout() time.Duration {
	cfg := viper.GetString("redis.write_timeout")
	return parseDuration(cfg, 2*time.Second)
}

// RedisReadTimeout :nodoc:
func RedisReadTimeout() time.Duration {
	cfg := viper.GetString("redis.read_timeout")
	return parseDuration(cfg, 2*time.Second)
}

// RedisMaxIdleConn :nodoc:
func RedisMaxIdleConn() int {
	if viper.GetInt("redis.max_idle_conn") > 0 {
		return viper.GetInt("redis.max_idle_conn")
	}
	return 20
}

// RedisMaxActiveConn :nodoc:
func RedisMaxActiveConn() int {
	if viper.GetInt("redis.max_active_conn") > 0 {
		return viper.GetInt("redis.max_active_conn")
	}
	return 50
}

// LoginRetryAttempts :nodoc:
func LoginRetryAttempts() int {
	if viper.IsSet("login.username_password.retry_attempts") {
		return viper.GetInt("login.username_password.retry_attempts")
	}

	return DefaultLoginRetryAttempts
}

// LoginLockTTL :nodoc:
func LoginLockTTL() time.Duration {
	cfg := viper.GetString("login.email_password.lock_ttl")
	return parseDuration(cfg, DefaultLoginLockTTL)
}

// AccessTokenDuration get access token increment duration in hour
func AccessTokenDuration() time.Duration {
	cfg := viper.GetString("session.access_token_duration")
	return parseDuration(cfg, DefaultAccessTokenDuration)
}

// RefreshTokenDuration get refresh token increment duration in hour
func RefreshTokenDuration() time.Duration {
	cfg := viper.GetString("session.refresh_token_duration")
	return parseDuration(cfg, DefaultRefreshTokenDuration)
}

// SecretKey :nodoc:
func SecretKey() string {
	val := viper.GetString("secret_key")
	if val == "" {
		log.Fatal("secret key not provided")
	}

	return val
}

// RedisCacheHost :nodoc:
func RedisCacheHost() string {
	return viper.GetString("redis.cache_host")
}

// RedisLockHost :nodoc:
func RedisLockHost() string {
	return viper.GetString("redis.lock_host")
}

// RedisAuthCacheHost config
func RedisAuthCacheHost() string {
	return viper.GetString("redis.auth_cache_host")
}

// RedisAuthCacheLockHost config
func RedisAuthCacheLockHost() string {
	return viper.GetString("redis.auth_cache_lock_host")
}

// CacheTTL :nodoc:
func CacheTTL() time.Duration {
	cfg := viper.GetString("cache_ttl")
	return parseDuration(cfg, DefaultCacheTTL)
}

func parseDuration(in string, defaultDuration time.Duration) time.Duration {
	dur, err := time.ParseDuration(in)
	if err != nil {
		return defaultDuration
	}
	return dur
}
