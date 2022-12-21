package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"user-service/utils"

	"user-service/internal/model"
	"user-service/rbac"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGroupRepository_CreateRoleResourceAction(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()

	initializeTest()
	mock := kit.dbmock
	repo := &rbacRepository{
		db:           kit.db,
		cacheManager: kit.cacheManager,
	}

	var (
		role     = rbac.RoleAdmin
		resource = rbac.ResourceUser
		action   = rbac.ActionEditAny
		ctx      = context.TODO()
	)

	rra := &model.RoleResourceAction{
		Role:     role,
		Resource: resource,
		Action:   action,
	}

	t.Run("ok - role resource action already exist", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "resources" WHERE "resources"."id" = $1 ORDER BY "resources"."id" LIMIT 1`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(resource))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "actions" WHERE "actions"."id" = $1 ORDER BY "actions"."id" LIMIT 1`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(action))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "role_resource_actions" WHERE "role" = $1 AND "resource" = $2 AND "action" = $3 LIMIT 1`)).
			WithArgs(role, resource, action).WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow(role))
		err := repo.CreateRoleResourceAction(ctx, rra)
		require.NoError(t, err)
	})
}

func TestGroupRepository_LoadPermission(t *testing.T) {
	kit, closer := initializeRepoTestKit(t)
	defer closer()

	initializeTest()
	mock := kit.dbmock
	repo := &rbacRepository{
		db:           kit.db,
		cacheManager: kit.cacheManager,
	}

	var (
		role     = rbac.RoleAdmin
		resource = rbac.ResourceUser
		action   = rbac.ActionEditAny
		ctx      = context.TODO()
	)

	t.Run("ok", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "role_resource_actions"`)).
			WillReturnRows(sqlmock.NewRows([]string{"role", "resource", "action"}).AddRow(role, resource, action))

		permission, err := repo.LoadPermission(ctx)
		require.NoError(t, err)
		require.NotNil(t, permission)
	})

	t.Run("ok - retrieve data from cache", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		cacheKey := model.RBACPermissionCacheKey
		perm := &rbac.Permission{
			RRA: map[rbac.Role][]rbac.ResourceAction{
				role: {
					rbac.ResourceAction{
						Resource: resource,
						Action:   action,
					},
				},
			},
		}

		err := kit.miniredis.Set(cacheKey, utils.Dump(perm))
		require.NoError(t, err)

		permission, err := repo.LoadPermission(ctx)
		require.NoError(t, err)
		require.NotNil(t, permission)
	})

	t.Run("failed", func(t *testing.T) {
		defer kit.miniredis.FlushDB()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "role_resource_actions"`)).WillReturnError(errors.New("db error"))
		permission, err := repo.LoadPermission(ctx)
		require.Error(t, err)
		require.Nil(t, permission)
	})
}
