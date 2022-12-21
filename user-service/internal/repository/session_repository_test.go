package repository

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
	"time"
	"user-service/internal/model"
	"user-service/utils"
)

func TestSessionRepository_Create(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock

	initializeTest()
	repo := sessionRepo{
		db:           kit.db,
		cacheManager: kit.cacheManager,
	}
	ctx := context.TODO()
	sessionID := utils.GenerateID()
	userID := utils.GenerateID()

	t.Run("ok", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("^INSERT INTO \"sessions\"").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(sessionID))
		mock.ExpectCommit()

		err := repo.Create(ctx, &model.Session{
			ID:           sessionID,
			UserID:       userID,
			AccessToken:  "at",
			RefreshToken: "rt",
		})
		require.NoError(t, err)
	})

	t.Run("handle error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("^INSERT INTO \"sessions\"").
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Create(ctx, &model.Session{})
		require.Error(t, err)
	})
}

func TestSessionRepository_FindByToken(t *testing.T) {
	initializeTest()
	kit, closer := initializeRepoTestKit(t)
	defer closer()

	mock := kit.dbmock
	repo := sessionRepo{
		db:           kit.db,
		cacheManager: kit.cacheManager,
		userRepo:     kit.mockUserRepo,
	}

	var (
		sessionID = utils.GenerateID()
		userID    = utils.GenerateID()
		token     = "token"
		ctx       = context.TODO()
	)

	t.Run("ok - access token", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"sessions\"").
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "access_token", "access_token_expired_at"}).
				AddRow(sessionID, userID, "access-token", time.Now().Add(time.Hour)))

		kit.mockUserRepo.EXPECT().FindByID(ctx, userID).Times(1).Return(&model.User{ID: userID}, nil)
		res, err := repo.FindByToken(ctx, model.AccessToken, token)
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("ok - refresh token", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"sessions\"").
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "access_token", "access_token_expired_at"}).
				AddRow(sessionID, userID, "access-token", time.Now().Add(time.Hour)))

		kit.mockUserRepo.EXPECT().FindByID(ctx, userID).Times(1).Return(&model.User{ID: userID}, nil)
		res, err := repo.FindByToken(ctx, model.RefreshToken, token)
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("ok - retrieve from cache", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		cacheKey := model.NewSessionTokenCacheKey(token)
		err := kit.miniredis.Set(cacheKey, utils.Dump(model.Session{ID: sessionID}))
		require.NoError(t, err)

		res, err := repo.FindByToken(ctx, model.RefreshToken, token)
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("failed - find session by id return not found", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"sessions\"").
			WillReturnError(gorm.ErrRecordNotFound)

		res, err := repo.FindByToken(ctx, model.AccessToken, token)
		require.NoError(t, err)
		require.Nil(t, res)

		cacheVal, err := kit.miniredis.Get(model.NewSessionTokenCacheKey(token))
		require.NoError(t, err)
		require.Equal(t, `null`, cacheVal)
	})

	t.Run("failed - find session by id return error", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"sessions\"").
			WillReturnError(errors.New("db error"))
		res, err := repo.FindByToken(ctx, model.AccessToken, token)
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("failed - find user by id return nil", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"sessions\"").
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "access_token", "access_token_expired_at"}).
				AddRow(sessionID, userID, "access-token", time.Now().Add(time.Hour)))

		kit.mockUserRepo.EXPECT().FindByID(ctx, userID).Times(1).Return(nil, nil)
		res, err := repo.FindByToken(ctx, model.AccessToken, token)
		require.NoError(t, err)
		require.Nil(t, res)
	})

	t.Run("failed - find user by id return error", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"sessions\"").
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "access_token", "access_token_expired_at"}).
				AddRow(sessionID, userID, "access-token", time.Now().Add(time.Hour)))

		kit.mockUserRepo.EXPECT().FindByID(ctx, userID).Times(1).Return(nil, errors.New("db error"))
		res, err := repo.FindByToken(ctx, model.AccessToken, token)
		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestSessionRepository_FindByID(t *testing.T) {
	initializeTest()
	kit, closer := initializeRepoTestKit(t)
	defer closer()

	mock := kit.dbmock
	repo := sessionRepo{
		db:           kit.db,
		cacheManager: kit.cacheManager,
		userRepo:     kit.mockUserRepo,
	}

	var (
		sessionID = utils.GenerateID()
		userID    = utils.GenerateID()
	)

	ctx := context.TODO()
	session := model.Session{ID: sessionID}

	t.Run("success - retrieve from db", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"sessions\"").
			WithArgs(session.ID).
			WillReturnRows(sqlmock.
				NewRows([]string{"id", "user_id", "access_token", "access_token_expired_at"}).
				AddRow(sessionID, userID, "access-token", time.Now().Add(time.Hour)))

		kit.mockUserRepo.EXPECT().FindByID(ctx, userID).Times(1).Return(&model.User{ID: userID}, nil)
		res, err := repo.FindByID(ctx, session.ID)
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("success - retrieve from token", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		cacheKey := repo.newCacheKeyByID(sessionID)
		err := kit.miniredis.Set(cacheKey, utils.Dump(model.Session{ID: sessionID}))
		require.NoError(t, err)

		res, err := repo.FindByID(ctx, session.ID)
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("failed - find session by id return nil", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"sessions\"").
			WithArgs(session.ID).
			WillReturnError(gorm.ErrRecordNotFound)

		res, err := repo.FindByID(ctx, session.ID)
		require.NoError(t, err)
		require.Nil(t, res)

		cacheVal, err := kit.miniredis.Get(repo.newCacheKeyByID(session.ID))
		require.NoError(t, err)
		require.Equal(t, `null`, cacheVal)
	})

	t.Run("failed - find session by id return err", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"sessions\"").
			WithArgs(session.ID).
			WillReturnError(errors.New("db error"))

		res, err := repo.FindByID(ctx, session.ID)
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("failed - find user by id return nil", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"sessions\"").
			WithArgs(session.ID).
			WillReturnRows(sqlmock.
				NewRows([]string{"id", "user_id", "access_token", "access_token_expired_at"}).
				AddRow(sessionID, userID, "access-token", time.Now().Add(time.Hour)))

		kit.mockUserRepo.EXPECT().FindByID(ctx, userID).Times(1).Return(nil, nil)
		res, err := repo.FindByID(ctx, session.ID)
		require.NoError(t, err)
		require.Nil(t, res)
	})

	t.Run("failed - find user by id return error", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"sessions\"").
			WithArgs(session.ID).
			WillReturnRows(sqlmock.
				NewRows([]string{"id", "user_id", "access_token", "access_token_expired_at"}).
				AddRow(sessionID, userID, "access-token", time.Now().Add(time.Hour)))

		kit.mockUserRepo.EXPECT().FindByID(ctx, userID).Times(1).Return(nil, errors.New("db error"))
		res, err := repo.FindByID(ctx, session.ID)
		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestSessionRepository_CacheToken(t *testing.T) {
	initializeTest()
	kit, closer := initializeRepoTestKit(t)
	defer closer()

	repo := sessionRepo{cacheManager: kit.cacheManager}

	var (
		sessionID = utils.GenerateID()
		token     = "token"
		ctx       = context.TODO()
	)

	t.Run("success - exist", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		cacheKey := model.NewSessionTokenCacheKey(token)
		err := kit.miniredis.Set(cacheKey, utils.Dump(model.Session{ID: sessionID}))
		require.NoError(t, err)

		isTokenExist, err := repo.CheckToken(ctx, token)
		require.NoError(t, err)
		require.True(t, isTokenExist)
	})

	t.Run("success - not exist", func(t *testing.T) {
		isTokenExist, err := repo.CheckToken(ctx, token)
		require.NoError(t, err)
		require.False(t, isTokenExist)
	})
}

func TestSessionRepository_RefreshToken(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock

	initializeTest()
	ctx := context.TODO()
	repo := sessionRepo{
		db:           kit.db,
		cacheManager: kit.cacheManager,
		userRepo:     kit.mockUserRepo,
	}

	now := time.Now()
	userID := utils.GenerateID()
	sess := &model.Session{
		ID:                    utils.GenerateID(),
		UserID:                userID,
		AccessToken:           "at",
		RefreshToken:          "rt",
		AccessTokenExpiredAt:  now,
		RefreshTokenExpiredAt: now,
		UserAgent:             "user-agent",
	}
	oldSess := *sess

	t.Run("ok", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("^UPDATE \"sessions\" SET ").WithArgs(
			sess.AccessToken,
			sess.RefreshToken,
			sess.AccessTokenExpiredAt,
			sess.RefreshTokenExpiredAt,
			sess.UserAgent,
			sqlmock.AnyArg(),
			sess.ID,
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		mock.ExpectQuery("^SELECT .+ FROM \"sessions\"").
			WithArgs(sess.ID).
			WillReturnRows(sqlmock.
				NewRows([]string{"id", "user_id", "access_token", "access_token_expired_at"}).
				AddRow(sess.ID, userID, sess.AccessToken, sess.AccessTokenExpiredAt))

		kit.mockUserRepo.EXPECT().FindByID(ctx, userID).Times(1).Return(&model.User{ID: userID}, nil)

		sess, err := repo.RefreshToken(ctx, &oldSess, sess)
		require.NoError(t, err)
		require.NotNil(t, sess)
	})

	t.Run("handle error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("^UPDATE \"sessions\" SET ")
		mock.ExpectCommit().WillReturnError(errors.New("db error"))

		sess, err := repo.RefreshToken(ctx, &oldSess, sess)
		require.Error(t, err)
		require.Nil(t, sess)
	})
}

func TestSessionRepo_Delete(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock

	initializeTest()
	repo := sessionRepo{
		db:           kit.db,
		cacheManager: kit.cacheManager,
		userRepo:     kit.mockUserRepo,
	}

	ctx := context.TODO()

	t.Run("ok", func(t *testing.T) {
		defer kit.miniredis.FlushDB()

		sess := &model.Session{
			ID:                    utils.GenerateID(),
			UserID:                utils.GenerateID(),
			AccessToken:           "at",
			RefreshTokenExpiredAt: time.Now(),
		}
		cacheKey := repo.newCacheKeyByID(sess.ID)
		err := kit.miniredis.Set(cacheKey, utils.Dump(model.Session{ID: sess.ID}))
		require.NoError(t, err)

		mock.ExpectBegin()
		mock.ExpectExec("^DELETE FROM \"sessions\"").
			WithArgs(sess.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		isExists := kit.miniredis.Del(cacheKey)
		require.Equal(t, true, isExists)

		err = repo.Delete(ctx, sess)
		require.NoError(t, err)
	})

	t.Run("error when delete session by id", func(t *testing.T) {
		defer kit.miniredis.FlushDB()

		sess := &model.Session{
			ID:                    utils.GenerateID(),
			UserID:                utils.GenerateID(),
			AccessToken:           "at",
			RefreshTokenExpiredAt: time.Now(),
		}
		cacheKey := repo.newCacheKeyByID(sess.ID)
		err := kit.miniredis.Set(cacheKey, utils.Dump(model.Session{ID: sess.ID}))
		require.NoError(t, err)

		mock.ExpectBegin()
		mock.ExpectExec("^DELETE FROM \"sessions\"").
			WithArgs(sess.ID).
			WillReturnError(errors.New("error"))
		mock.ExpectCommit()

		isExists := kit.miniredis.Del(cacheKey)
		require.Equal(t, true, isExists)

		err = repo.Delete(ctx, sess)
		require.Error(t, err)
	})
}
