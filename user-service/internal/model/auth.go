package model

import (
	"context"
	"user-service/rbac"
)

// LoginRequest request
type LoginRequest struct {
	Email, PlainPassword, UserAgent string
}

// RefreshTokenRequest request
type RefreshTokenRequest struct {
	RefreshToken, UserAgent string
}

// AuthUsecase usecases about IAM
type AuthUsecase interface {
	LoginByEmailPassword(ctx context.Context, req LoginRequest) (*Session, error)
	// AuthenticateToken authenticate the given token
	AuthenticateToken(ctx context.Context, accessToken string) (*User, error)
	FindRolePermission(ctx context.Context, role rbac.Role) (*rbac.RolePermission, error)
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*Session, error)
	DeleteSessionByID(ctx context.Context, sessionID int64) error
}
