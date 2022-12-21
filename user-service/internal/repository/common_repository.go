package repository

import (
	"github.com/sirupsen/logrus"
	"user-service/cacher"
)

func storeNil(ck cacher.CacheManager, key string) {
	err := ck.StoreNil(key)
	if err != nil {
		logrus.Error(err)
	}
}
