package usecase

import "errors"

var (
	ErrPermissionDenied           = errors.New("permission denied")
	ErrNotFound                   = errors.New("not found")
	ErrFailedPrecondition         = errors.New("precondition failed")
	ErrRoleNotFound               = errors.New("role not found")
	ErrPasswordMismatch           = errors.New("password mismatch")
	ErrLoginByEmailPasswordLocked = errors.New("user is locked from logging in using email and password")
	ErrUnauthorized               = errors.New("unauthorized")
	ErrAccessTokenExpired         = errors.New("access token expired")
	ErrRefreshTokenExpired        = errors.New("refresh token expired")
	ErrDuplicateEmail             = errors.New("email already exist")
)
