package config

import "time"

const (
	DefaultDatabaseMaxIdleConns    = 3
	DefaultDatabaseMaxOpenConns    = 5
	DefaultDatabaseConnMaxLifetime = 1 * time.Hour
	DefaultDatabasePingInterval    = 1 * time.Second
	DefaultDatabaseRetryAttempts   = 3

	DefaultLoginRetryAttempts = 3
	DefaultCacheTTL           = 15 * time.Minute
	DefaultLoginLockTTL       = 5 * time.Minute

	DefaultSessionTokenLength   = 50
	DefaultAccessTokenDuration  = 1 * time.Hour
	DefaultRefreshTokenDuration = 24 * time.Hour * 1 // 1 day
)
