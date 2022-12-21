package rbac

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRole_HasAccess(t *testing.T) {
	resourceDummy := Resource("dummy")
	actionViewDummy := Action("view_dummy")
	perm := &Permission{}
	perm.Add(RoleAdmin, resourceDummy, actionViewDummy)

	t.Run("the role has access", func(t *testing.T) {
		rolePerm := NewRolePermission(RoleAdmin, perm)
		hasAccess := rolePerm.HasAccess(resourceDummy, actionViewDummy)
		require.True(t, hasAccess)
	})

	t.Run("the role does not have access", func(t *testing.T) {
		rolePerm := NewRolePermission(RoleAdmin, perm)
		hasAccess := rolePerm.HasAccess(resourceDummy, ActionCreateAny)
		require.False(t, hasAccess)
	})

	t.Run("role not exist", func(t *testing.T) {
		rolePerm := NewRolePermission(RoleMember, perm)
		hasAccess := rolePerm.HasAccess(Resource("resource not exists 404"), ActionCreateAny)
		require.False(t, hasAccess)
	})
}
