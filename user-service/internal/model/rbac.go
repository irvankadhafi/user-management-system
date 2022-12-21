package model

import (
	"context"
	"user-service/rbac"
)

// RBACRepository repository
type RBACRepository interface {
	CreateRoleResourceAction(ctx context.Context, rra *RoleResourceAction) error
	LoadPermission(ctx context.Context) (*rbac.Permission, error)
}

// RoleResourceAction model
type RoleResourceAction struct {
	Role     rbac.Role
	Resource rbac.Resource
	Action   rbac.Action
}

// RBACPermissionCacheKey cache key
const RBACPermissionCacheKey = "cache:object:rbac:permission"
