package repository

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"reflect"
	"strconv"
	"testing"
	"user-service/internal/config"
	"user-service/internal/model"
	"user-service/rbac"
	"user-service/utils"
)

func TestUserRepository_Create(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock

	initializeTest()

	ctx := context.TODO()
	repo := &userRepository{
		db:           kit.db,
		cacheManager: kit.cacheManager,
	}

	user := &model.User{
		ID:          utils.GenerateID(),
		Email:       "test@irvan.com",
		Role:        rbac.RoleMember,
		PhoneNumber: "0819992299",
	}
	userRequesterID := utils.GenerateID()

	t.Run("ok", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id",
			"email",
			"role",
			"phone_number",
		})
		rows.AddRow(user.ID, user.Email, user.Role, user.PhoneNumber)

		mock.ExpectBegin()
		mock.ExpectQuery(`^INSERT INTO "users"`).
			WillReturnRows(rows)
		mock.ExpectCommit()

		err := repo.Create(ctx, userRequesterID, user)
		require.NoError(t, err)
	})

	t.Run("handle error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`^INSERT INTO "users"`).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Create(ctx, userRequesterID, user)
		require.Error(t, err)
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock

	initializeTest()
	repo := userRepository{
		db:           kit.db,
		cacheManager: kit.cacheManager,
	}
	ctx := context.TODO()

	userID := utils.GenerateID()

	t.Run("ok - retrieve from db", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"users\"").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

		res, err := repo.FindByID(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.True(t, kit.miniredis.Exists(repo.newCacheKeyByID(userID)))
	})

	t.Run("failed - find user by id return nil", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"users\"").WillReturnError(gorm.ErrRecordNotFound)
		res, err := repo.FindByID(ctx, userID)
		require.NoError(t, err)
		require.Nil(t, res)

		cacheVal, err := kit.miniredis.Get(repo.newCacheKeyByID(userID))
		require.NoError(t, err)
		require.Equal(t, `null`, cacheVal)
	})

	t.Run("failed - find user by id return err", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"users\"").WillReturnError(errors.New("db error"))
		res, err := repo.FindByID(ctx, userID)
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("not found cache", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		viper.Set("disable_caching", false)

		cacheKey := repo.newCacheKeyByID(userID)
		err := kit.cacheManager.StoreNil(cacheKey)
		require.NoError(t, err)

		res, err := repo.FindByID(ctx, userID)
		require.NoError(t, err)
		require.Nil(t, res)
	})
}

func TestUserRepository_FindByEEmail(t *testing.T) {
	initializeTest()
	kit, closer := initializeRepoTestKit(t)
	defer closer()

	mock := kit.dbmock
	repo := &userRepository{
		db:           kit.db,
		cacheManager: kit.cacheManager,
	}

	var (
		userID    = utils.GenerateID()
		userEmail = "test@irvan.com"
		ctx       = context.TODO()
	)

	user := model.User{
		ID:    userID,
		Email: userEmail,
	}

	t.Run("ok - retrieve from db", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"users\"").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(user.ID))

		findByIDPatch := gomonkey.ApplyMethod(reflect.TypeOf(repo), "FindByID", func(repo *userRepository, ctx context.Context, id int64) (*model.User, error) {
			return &user, nil
		})
		defer findByIDPatch.Reset()

		res, err := repo.FindByEmail(ctx, user.Email)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.True(t, kit.miniredis.Exists(repo.newCacheKeyByEmail(user.Email)))
	})

	t.Run("ok - retrieve from cache", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		cacheKey := repo.newCacheKeyByEmail(userEmail)
		err := kit.miniredis.Set(cacheKey, utils.Int64ToString(userID))
		require.NoError(t, err)

		findByIDPatch := gomonkey.ApplyMethod(reflect.TypeOf(repo), "FindByID", func(repo *userRepository, ctx context.Context, id int64) (*model.User, error) {
			return &user, nil
		})
		defer findByIDPatch.Reset()

		res, err := repo.FindByEmail(ctx, user.Email)
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("failed - find by email return nil", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"users\"").WillReturnError(gorm.ErrRecordNotFound)

		res, err := repo.FindByEmail(ctx, userEmail)
		require.NoError(t, err)
		require.Nil(t, res)

		cacheVal, err := kit.miniredis.Get(repo.newCacheKeyByEmail(userEmail))
		require.NoError(t, err)
		require.Equal(t, `null`, cacheVal)
	})

	t.Run("failed - find by email return err", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery("^SELECT .+ FROM \"users\"").WillReturnError(errors.New("db error"))

		res, err := repo.FindByEmail(ctx, userEmail)
		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestUserRepository_Update(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock

	initializeTest()

	ctx := context.TODO()
	repo := &userRepository{
		db:           kit.db,
		cacheManager: kit.cacheManager,
	}

	userRequesterID := utils.GenerateID()

	user := &model.User{
		ID:       utils.GenerateID(),
		Name:     "John Doe",
		Email:    "john@doe.com",
		Role:     rbac.RoleMember,
		Password: "hahahihi",
	}

	patch := gomonkey.ApplyMethod(reflect.TypeOf(repo), "FindByID", func(_ *userRepository, _ context.Context, _ int64) (*model.User, error) {
		return user, nil
	})
	defer patch.Reset()

	t.Run("ok", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^UPDATE "users" SET (.+)`).
			WithArgs(
				user.Name,
				user.Email,
				user.Password,
				user.Role,
				userRequesterID,  // updated_by
				sqlmock.AnyArg(), // updated_at
				user.ID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		res, err := repo.Update(ctx, userRequesterID, user)
		require.NoError(t, err)
		require.NotEmpty(t, res)
	})

	t.Run("handle error", func(t *testing.T) {
		mock.ExpectQuery(`^UPDATE "users" SET (.+)`).
			WillReturnError(errors.New("db error"))

		res, err := repo.Update(ctx, userRequesterID, user)
		require.Error(t, err)
		require.Empty(t, res)
	})
}

func TestUserRepository_IsLoginByEmailPasswordLocked(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()

	initializeTest()
	ctx := context.TODO()
	userEmail := "john.doe@mail.com"

	repo := &userRepository{cacheManager: kit.cacheManager}

	t.Run("success, login is locked", func(t *testing.T) {
		viper.Set("disable_caching", false)
		attempts := config.DefaultLoginRetryAttempts
		lockKey := repo.newLoginByEmailPasswordAttemptsCacheKeyByEmail(userEmail)
		err := kit.miniredis.Set(lockKey, strconv.Itoa(attempts))
		require.NoError(t, err)

		kit.miniredis.SetTTL(lockKey, config.LoginLockTTL())

		isLocked, err := repo.IsLoginByEmailPasswordLocked(ctx, userEmail)
		require.NoError(t, err)
		require.True(t, isLocked)
	})

	t.Run("success, login is unlocked", func(t *testing.T) {
		viper.Set("disable_caching", false)
		attempts := int64(0)
		lockKey := repo.newLoginByEmailPasswordAttemptsCacheKeyByEmail(userEmail)
		err := kit.miniredis.Set(lockKey, utils.Int64ToString(attempts))
		require.NoError(t, err)

		kit.miniredis.SetTTL(lockKey, config.LoginLockTTL())

		isLocked, err := repo.IsLoginByEmailPasswordLocked(ctx, userEmail)
		require.NoError(t, err)
		require.False(t, isLocked)
	})
}

func TestUserRepository_IncrementLoginByEmailPasswordRetryAttempts(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()

	initializeTest()
	ctx := context.TODO()
	userEmail := "john.doe@mail.com"
	repo := &userRepository{cacheManager: kit.cacheManager}
	key := repo.newLoginByEmailPasswordAttemptsCacheKeyByEmail(userEmail)

	t.Run("success", func(t *testing.T) {
		viper.Set("disable_caching", false)
		err := repo.IncrementLoginByEmailPasswordRetryAttempts(ctx, userEmail)
		require.NoError(t, err)

		attempts, err := kit.miniredis.Get(key)
		require.NoError(t, err)
		require.Equal(t, "1", attempts)
	})
}

func TestUserRepository_FindPasswordByID(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()
	mock := kit.dbmock

	initializeTest()
	repo := &userRepository{
		db:           kit.db,
		cacheManager: kit.cacheManager,
	}

	ctx := context.TODO()
	password := "password"

	user := model.User{
		ID:       utils.GenerateID(),
		Email:    "test@email.com",
		Password: password,
	}

	t.Run("ok - retrieve from db", func(t *testing.T) {
		mock.ExpectQuery("^SELECT .+ FROM \"users\"").WillReturnRows(sqlmock.NewRows([]string{"password"}).AddRow("wowPassword"))

		findByIDPatch := gomonkey.ApplyMethod(reflect.TypeOf(repo), "FindByID", func(repo *userRepository, ctx context.Context, id int64) (*model.User, error) {
			return &user, nil
		})
		defer findByIDPatch.Reset()

		res, err := repo.FindPasswordByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("ok - retrieve from cache", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		viper.Set("disable_caching", false)
		err := kit.miniredis.Set(
			repo.newPasswordCacheKeyByID(user.ID),
			password,
		)
		require.NoError(t, err)

		pwd, err := repo.FindPasswordByID(ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, password, string(pwd))
	})

	t.Run("failed - find password by id return nil", func(t *testing.T) {
		mock.ExpectQuery("^SELECT .+ FROM \"users\"").WillReturnError(gorm.ErrRecordNotFound)
		res, err := repo.FindPasswordByID(ctx, user.ID)
		require.NoError(t, err)
		require.Nil(t, res)
	})

	t.Run("failed - find password by id return err", func(t *testing.T) {
		mock.ExpectQuery("^SELECT .+ FROM \"users\"").WillReturnError(errors.New("db err"))
		res, err := repo.FindPasswordByID(ctx, user.ID)
		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestUserRepository_UpdatePasswordByID(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()

	initializeTest()
	mock := kit.dbmock
	ctx := context.TODO()

	var (
		userID   = utils.GenerateID()
		password = "iniPaswordH3h3h"
	)

	repo := &userRepository{
		db:           kit.db,
		cacheManager: kit.cacheManager,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^UPDATE "users" SET `).WithArgs(password, sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdatePasswordByID(ctx, userID, password)
		require.NoError(t, err)
	})

	t.Run("failed update password", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^UPDATE "users" SET `).WithArgs(password, sqlmock.AnyArg(), userID).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.UpdatePasswordByID(ctx, userID, password)
		require.Error(t, err)
	})
}
