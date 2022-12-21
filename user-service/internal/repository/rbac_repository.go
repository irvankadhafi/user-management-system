package repository

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
	"user-service/cacher"
	"user-service/internal/config"
	"user-service/internal/model"
	"user-service/rbac"
	"user-service/utils"
)

// rbacRepository repository
type rbacRepository struct {
	db           *gorm.DB
	cacheManager cacher.CacheManager
}

// NewRBACRepository constructor
func NewRBACRepository(db *gorm.DB, cacheManager cacher.CacheManager) model.RBACRepository {
	return &rbacRepository{
		db:           db,
		cacheManager: cacheManager,
	}
}

// CreateRoleResourceAction create new record ignore if exists
func (r *rbacRepository) CreateRoleResourceAction(ctx context.Context, rra *model.RoleResourceAction) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": utils.DumpIncomingContext(ctx),
		"rra": utils.Dump(rra),
	})

	if err := r.createResource(ctx, rra.Resource); err != nil {
		logger.Error(err)
		return err
	}

	if err := r.createAction(ctx, rra.Action); err != nil {
		logger.Error(err)
		return err
	}

	err := r.db.WithContext(ctx).Take(&rra, `"role" = ? AND "resource" = ? AND "action" = ?`, rra.Role, rra.Resource, rra.Action).Error
	switch err {
	case nil:
		return nil
	case gorm.ErrRecordNotFound:
		err := r.db.WithContext(ctx).Create(&rra).Error
		if err != nil {
			logger.Error(err)
		}
		return err
	default:
		logger.Error(err)
		return err
	}
}

// LoadPermission find the permissions
func (r *rbacRepository) LoadPermission(ctx context.Context) (*rbac.Permission, error) {
	cacheKey := model.RBACPermissionCacheKey
	reply, mu, err := r.cacheManager.GetOrLock(cacheKey)
	if err != nil {
		return nil, err
	}
	defer cacher.SafeUnlock(mu)

	perm := &rbac.Permission{}
	if reply != nil {
		bt, _ := reply.([]byte)
		if err := json.Unmarshal(bt, &perm); err != nil {
			logrus.Error(err)
			return nil, err
		}
		return perm, nil
	}

	if !config.DisableCaching() && mu == nil {
		return nil, nil
	}

	// fallback to db
	rras, err := r.findAllRoleResourceAction(ctx)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	for _, rra := range rras {
		perm.Add(rra.Role, rra.Resource, rra.Action)
	}

	cacheItem := cacher.NewItemWithCustomTTL(cacheKey, utils.Dump(perm), time.Hour*24)
	err = r.cacheManager.StoreWithoutBlocking(cacheItem)
	if err != nil {
		logrus.Error(err)
	}

	return perm, nil
}

func (r *rbacRepository) createResource(ctx context.Context, rsc rbac.Resource) error {
	type Resource struct {
		ID rbac.Resource
	}
	resource := Resource{ID: rsc}
	err := r.db.WithContext(ctx).FirstOrCreate(&resource).Error
	if err != nil {
		logrus.Error(err)
	}
	return err
}

func (r *rbacRepository) createAction(ctx context.Context, act rbac.Action) error {
	type Action struct {
		ID rbac.Action
	}
	resource := Action{ID: act}
	err := r.db.WithContext(ctx).FirstOrCreate(&resource).Error
	if err != nil {
		logrus.Error(err)
	}
	return err
}

func (r *rbacRepository) findAllRoleResourceAction(ctx context.Context) (rra []model.RoleResourceAction, err error) {
	err = r.db.WithContext(ctx).Debug().Find(&rra).Error
	return rra, err
}
