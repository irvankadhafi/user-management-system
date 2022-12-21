package httpsvc

import (
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"net/http"
	"user-service/internal/delivery"
	"user-service/internal/model"
	"user-service/internal/usecase"
	"user-service/utils"
)

type loginResponse struct {
	AccessToken           string `json:"access_token"`
	AccessTokenExpiresAt  string `json:"access_token_expires_at"`
	TokenType             string `json:"token_type"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresAt string `json:"refresh_token_expires_at"`
}

func (s *Service) handleLoginByEmailPassword() echo.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(c echo.Context) error {
		req := request{}
		if err := c.Bind(&req); err != nil {
			logrus.Error(err)
			return ErrInvalidArgument
		}

		session, err := s.authUsecase.LoginByEmailPassword(c.Request().Context(), model.LoginRequest{
			Email:         req.Email,
			PlainPassword: req.Password,
			UserAgent:     c.Request().UserAgent(),
		})
		switch err {
		case nil:
			break
		case usecase.ErrNotFound, usecase.ErrUnauthorized:
			return ErrEmailPasswordNotMatch
		case usecase.ErrLoginByEmailPasswordLocked:
			return ErrLoginByEmailPasswordLocked
		case usecase.ErrPermissionDenied:
			return ErrPermissionDenied
		default:
			logrus.Error(err)
			return ErrInternal
		}

		res := loginResponse{
			AccessToken:           session.AccessToken,
			AccessTokenExpiresAt:  utils.FormatTimeRFC3339(&session.AccessTokenExpiredAt),
			RefreshToken:          session.RefreshToken,
			RefreshTokenExpiresAt: utils.FormatTimeRFC3339(&session.RefreshTokenExpiredAt),
			TokenType:             "Bearer",
		}

		return c.JSON(http.StatusOK, res)
	}
}

func (s *Service) handleRefreshToken() echo.HandlerFunc {
	type request struct {
		RefreshToken string `json:"refresh_token"`
	}

	return func(c echo.Context) error {
		req := request{}
		if err := c.Bind(&req); err != nil {
			logrus.Error(err)
			return ErrInvalidArgument
		}

		session, err := s.authUsecase.RefreshToken(c.Request().Context(), model.RefreshTokenRequest{
			RefreshToken: req.RefreshToken,
			UserAgent:    c.Request().UserAgent(),
		})
		switch err {
		case nil:
		case usecase.ErrRefreshTokenExpired, usecase.ErrNotFound:
			return ErrUnauthenticated
		default:
			logrus.Error(err)
			return ErrInternal
		}

		res := loginResponse{
			AccessToken:           session.AccessToken,
			AccessTokenExpiresAt:  utils.FormatTimeRFC3339(&session.AccessTokenExpiredAt),
			RefreshToken:          session.RefreshToken,
			RefreshTokenExpiresAt: utils.FormatTimeRFC3339(&session.RefreshTokenExpiredAt),
			TokenType:             "Bearer",
		}
		return c.JSON(http.StatusOK, res)
	}
}

func (s *Service) handleLogout() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		requester := delivery.GetAuthUserFromCtx(ctx)

		err := s.authUsecase.DeleteSessionByID(c.Request().Context(), requester.SessionID)
		switch err {
		case nil:
			break
		case usecase.ErrNotFound:
			return ErrNotFound
		default:
			logrus.Error(err)
			return httpValidationOrInternalErr(err)
		}

		return c.NoContent(http.StatusNoContent)
	}
}
