package auth

import (
	"context"
	uuid "github.com/satori/go.uuid"
	"user-service/internal/model"
	"user-service/rbac"
)

type contextKey string

// use module path to make it unique
const userCtxKey contextKey = "user-service/auth.User"

// SetUserToCtx set user to context
func SetUserToCtx(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

// GetUserFromCtx get user from context
func GetUserFromCtx(ctx context.Context) *User {
	user, ok := ctx.Value(userCtxKey).(User)
	if !ok {
		return nil
	}
	return &user
}

// User represent an authenticated user
type User struct {
	ID             uuid.UUID            `json:"id"`
	Role           rbac.Role            `json:"role"`
	SessionID      uuid.UUID            `json:"session_id"`
	RolePermission *rbac.RolePermission `json:"-"`
}

// NewUserFromSession return new user from session
func NewUserFromSession(sess model.Session, perm *rbac.Permission) User {
	rp := rbac.NewRolePermission(sess.Role, perm)
	return User{
		ID:             sess.UserID,
		Role:           sess.Role,
		RolePermission: rp,
		SessionID:      sess.ID,
	}
}

// HasAccess check the user authorization
func (u *User) HasAccess(resource rbac.Resource, action rbac.Action) error {
	if u.RolePermission == nil || !u.RolePermission.HasAccess(resource, action) {
		return ErrAccessDenied
	}

	return nil
}
