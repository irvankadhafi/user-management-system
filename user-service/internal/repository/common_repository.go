package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
	"user-service/cacher"
)

func storeNil(ck cacher.CacheManager, key string) {
	err := ck.StoreNil(key)
	if err != nil {
		logrus.Error(err)
	}
}

func withSize(size int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(int(size))
	}
}

func withOffset(offset int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(int(offset))
	}
}

func withNameLike(query string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(query)+"%")
	}
}

func orderByCreatedAtDesc(db *gorm.DB) *gorm.DB {
	return db.Order("created_at DESC")
}

func orderByCreatedAtAsc(db *gorm.DB) *gorm.DB {
	return db.Order("created_at ASC")
}

func orderByUpdatedAtDesc(db *gorm.DB) *gorm.DB {
	return db.Order("updated_at DESC")
}

func orderByUpdatedAtAsc(db *gorm.DB) *gorm.DB {
	return db.Order("updated_at ASC")
}
