package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
	"user-service/cacher"
	"user-service/internal/config"
	"user-service/internal/model"
	"user-service/utils"
)

type sessionRepo struct {
	db           *gorm.DB
	cacheManager cacher.CacheManager
	userRepo     model.UserRepository
}

// NewSessionRepository sessionRepo constructor
func NewSessionRepository(
	db *gorm.DB,
	cacheKeeper cacher.CacheManager,
	userRepo model.UserRepository,
) model.SessionRepository {
	return &sessionRepo{
		db:           db,
		cacheManager: cacheKeeper,
		userRepo:     userRepo,
	}
}

func (s *sessionRepo) Create(ctx context.Context, sess *model.Session) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":    utils.DumpIncomingContext(ctx),
		"userID": sess.UserID,
	})

	if err := s.db.WithContext(ctx).Debug().Create(sess).Error; err != nil {
		logger.Error(err)
		return err
	}

	if err := s.cacheToken(sess); err != nil {
		logger.Error(err)
	}
	return nil
}

func (s *sessionRepo) FindByToken(ctx context.Context, tokenType model.TokenType, token string) (*model.Session, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"tokenType": tokenType,
	})

	cacheKey := model.NewSessionTokenCacheKey(token)
	if !config.DisableCaching() {
		reply, mu, err := s.findFromCacheByKey(cacheKey)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		defer cacher.SafeUnlock(mu)

		if mu == nil {
			return reply, nil
		}
	}

	sess := &model.Session{}
	var err error
	switch tokenType {
	case model.AccessToken:
		err = s.db.Take(sess, "access_token = ?", token).Error
	case model.RefreshToken:
		err = s.db.Take(sess, "refresh_token = ?", token).Error
	}
	switch err {
	case nil:
	case gorm.ErrRecordNotFound:
		storeNil(s.cacheManager, cacheKey)
		return nil, nil
	default:
		logger.Error(err)
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, sess.UserID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	sess.Role = user.Role
	if err = s.cacheToken(sess); err != nil {
		logger.Error(err)
	}

	return sess, nil
}

func (s *sessionRepo) FindByID(ctx context.Context, id int64) (*model.Session, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": utils.DumpIncomingContext(ctx),
		"id":  id,
	})

	cacheKey := s.newCacheKeyByID(id)
	reply, mu, err := s.findFromCacheByKey(cacheKey)
	if err != nil {
		return nil, err
	}
	defer cacher.SafeUnlock(mu)

	if mu == nil {
		return reply, nil
	}

	sess := model.Session{}
	err = s.db.WithContext(ctx).Take(&sess, "id = ?", id).Error
	switch err {
	case nil:
	case gorm.ErrRecordNotFound:
		storeNil(s.cacheManager, cacheKey)
		return nil, nil
	default:
		logger.Error(err)
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, sess.UserID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if user == nil {
		return nil, nil
	}
	sess.Role = user.Role
	if err = s.cacheToken(&sess); err != nil {
		logger.Error(err)
	}

	return &sess, nil
}

func (s *sessionRepo) CheckToken(ctx context.Context, token string) (exist bool, err error) {
	reply, err := s.cacheManager.Get(model.NewSessionTokenCacheKey(token))
	if err != nil {
		return false, err
	}

	bt, _ := reply.([]byte)
	return string(bt) != "", nil
}

func (s *sessionRepo) RefreshToken(ctx context.Context, oldSess, sess *model.Session) (*model.Session, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     utils.DumpIncomingContext(ctx),
		"session": utils.Dump(sess),
	})

	sess.UpdatedAt = time.Now()
	err := s.db.WithContext(ctx).Model(model.Session{}).Select(
		"access_token",
		"refresh_token",
		"access_token_expired_at",
		"refresh_token_expired_at",
		"user_agent",
		"ip_address",
		"updated_at",
	).Where("id = ?", sess.ID).Updates(sess).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if err = s.deleteCaches(oldSess); err != nil {
		logger.Error(err)
	}
	if err = s.deleteCaches(sess); err != nil {
		logger.Error(err)
	}

	return s.FindByID(ctx, sess.ID)
}

func (s *sessionRepo) Delete(ctx context.Context, session *model.Session) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":     utils.DumpIncomingContext(ctx),
		"session": utils.Dump(session),
	})

	err := s.db.WithContext(ctx).Delete(session).Error
	if err != nil {
		logger.Error(err)
		return err
	}

	if err := s.deleteCaches(session); err != nil {
		logger.Error(err)
	}

	return nil
}

func (s *sessionRepo) cacheToken(session *model.Session) error {
	sess, err := json.Marshal(session)
	if err != nil {
		return err
	}

	now := time.Now()
	return s.cacheManager.StoreMultiWithoutBlocking([]cacher.Item{
		cacher.NewItemWithCustomTTL(model.NewSessionTokenCacheKey(session.AccessToken), sess, session.AccessTokenExpiredAt.Sub(now)),
		cacher.NewItemWithCustomTTL(s.newCacheKeyByID(session.ID), sess, session.AccessTokenExpiredAt.Sub(now)),
		cacher.NewItemWithCustomTTL(model.NewSessionTokenCacheKey(session.RefreshToken), sess, session.RefreshTokenExpiredAt.Sub(now)),
	})
}

func (s *sessionRepo) deleteCaches(session *model.Session) error {
	return s.cacheManager.DeleteByKeys([]string{
		model.NewSessionTokenCacheKey(session.AccessToken),
		model.NewSessionTokenCacheKey(session.RefreshToken),
		s.newCacheKeyByID(session.ID),
	})
}

func (s *sessionRepo) newCacheKeyByID(id int64) string {
	return fmt.Sprintf("cache:object:session:id:%d", id)
}

func (s *sessionRepo) findFromCacheByKey(key string) (reply *model.Session, mu *redsync.Mutex, err error) {
	var rep interface{}
	rep, mu, err = s.cacheManager.GetOrLock(key)
	if err != nil || rep == nil {
		return
	}

	reply = utils.InterfaceBytesToType[*model.Session](rep)
	return
}
