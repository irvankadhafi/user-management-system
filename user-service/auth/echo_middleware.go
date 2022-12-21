package auth

import (
	"context"
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"
	"user-service/cacher"
	"user-service/internal/model"
	"user-service/rbac"
)

// UserAuthenticator to perform user authentication
type UserAuthenticator interface {
	AuthenticateToken(ctx context.Context, accessToken string) (*User, error)
}

// AuthenticationMiddleware middleware for authentication
type AuthenticationMiddleware struct {
	cacheManager      cacher.CacheManager
	userAuthenticator UserAuthenticator
}

// NewAuthenticationMiddleware AuthMiddleware constructor
func NewAuthenticationMiddleware(
	userAuthenticator UserAuthenticator,
	cacheManager cacher.CacheManager,
) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		userAuthenticator: userAuthenticator,
		cacheManager:      cacheManager,
	}
}

// AuthenticateAccessToken authenticate access token from http `Authorization` header and load a User to context
func (a *AuthenticationMiddleware) AuthenticateAccessToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := getAccessToken(c.Request())
			return a.authenticateAccessToken(c, next, token)
		}
	}
}

// MustAuthenticateAccessToken must authenticate access token from http `Authorization` header and load a User to context
// Differ from AuthenticateAccessToken, if no token provided then return Unauthenticated
func (a *AuthenticationMiddleware) MustAuthenticateAccessToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := getAccessToken(c.Request())
			if token == "" {
				return errorResp(http.StatusUnauthorized, "user is unauthenticated")
			}

			return a.authenticateAccessToken(c, next, token)
		}
	}
}

func (a *AuthenticationMiddleware) authenticateAccessToken(c echo.Context, next echo.HandlerFunc, token string) error {
	// only load user to context when token presented
	if token == "" {
		return next(c)
	}
	ctx := c.Request().Context()

	session, perm, err := a.findSessionAndPermissionFromCache(token)
	switch err {
	default:
		logrus.WithField("sessionCacheError", "find session from cache got error").Error(err)
	case nil:
		if session == nil || perm == nil {
			break // fallback
		}
		if session.IsAccessTokenExpired() {
			return errorResp(http.StatusUnauthorized, "token expired")
		}

		ctx := SetUserToCtx(c.Request().Context(), NewUserFromSession(*session, perm))
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}

	userSession, err := a.userAuthenticator.AuthenticateToken(ctx, token)
	// fallback to rpc
	switch status.Code(err) {
	case codes.OK:
		if userSession == nil { // safety check
			return next(c)
		}

		ctx := SetUserToCtx(c.Request().Context(), *userSession)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	case codes.NotFound:
		return errorResp(http.StatusBadRequest, "token is invalid")
	case codes.Unauthenticated:
		return errorResp(http.StatusUnauthorized, "token is expired")
	default:
		logrus.Error(err)
		return errorResp(http.StatusInternalServerError, "system error")
	}
}

func (a *AuthenticationMiddleware) findSessionAndPermissionFromCache(token string) (*model.Session, *rbac.Permission, error) {
	sess, err := a.findSessionFromCache(token)
	if err != nil {
		return nil, nil, err
	}

	perm, err := findPermissionFromCache(a.cacheManager)
	if err != nil {
		return nil, nil, err
	}

	return sess, perm, nil
}

func (a *AuthenticationMiddleware) findSessionFromCache(token string) (*model.Session, error) {
	reply, err := a.cacheManager.Get(model.NewSessionTokenCacheKey(token))
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	if reply == nil {
		return nil, nil
	}

	bt, _ := reply.([]byte)
	if bt == nil {
		return nil, nil
	}

	sess := &model.Session{}
	err = json.Unmarshal(bt, sess)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	return sess, nil
}

func findPermissionFromCache(cacheManager cacher.CacheManager) (*rbac.Permission, error) {
	cacheKey := model.RBACPermissionCacheKey
	reply, err := cacheManager.Get(cacheKey)
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, nil
	}

	perm := &rbac.Permission{}
	bt, _ := reply.([]byte)
	if err := json.Unmarshal(bt, &perm); err != nil {
		logrus.Error(err)
		return nil, err
	}

	return perm, nil
}

func getAccessToken(req *http.Request) (accessToken string) {
	authHeader := req.Header.Get("Authorization")

	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func errorResp(code int, message string) error {
	return echo.NewHTTPError(code, echo.Map{"message": message})
}
