package auth

import (
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"user-service/rbac"
)

func TestUser_HasAccess(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		perm := rbac.NewPermission()
		perm.Add(rbac.RoleAdmin, rbac.ResourceMember, rbac.ActionCreateAny)
		user := User{
			ID:             uuid.NewV4(),
			Role:           rbac.RoleAdmin,
			RolePermission: rbac.NewRolePermission(rbac.RoleAdmin, perm),
		}

		err := user.HasAccess(rbac.ResourceMember, rbac.ActionCreateAny)
		require.NoError(t, err)
	})

	t.Run("error access denied", func(t *testing.T) {
		user := User{
			ID:   uuid.NewV4(),
			Role: rbac.RoleMember,
		}

		err := user.HasAccess(rbac.ResourceMember, rbac.ActionCreateAny)
		require.Error(t, err)
		require.Equal(t, ErrAccessDenied, err)
	})
}
