package httpsvc

import (
	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
	"user-service/internal/delivery"
	"user-service/internal/usecase"
	"user-service/rbac"
	"user-service/utils"
)

type userResponse struct {
	ID          string    `json:"id" `
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Role        rbac.Role `json:"role"`
	PhoneNumber string    `json:"phone_number"`
	CreatedBy   string    `json:"created_by,omitempty"`
	UpdatedBy   string    `json:"updated_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Service) handleGetCurrentLoginUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		requester := delivery.GetAuthUserFromCtx(ctx)

		logger := logrus.WithFields(logrus.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"requesterID": requester.ID,
		})

		user, err := s.userUsecase.FindByID(ctx, requester, requester.ID)
		switch err {
		case nil:
			break
		case usecase.ErrNotFound:
			return ErrNotFound
		default:
			logger.Error(err)
			return ErrInternal
		}

		res := userResponse{
			ID:          user.ID.String(),
			Name:        user.Name,
			Email:       user.Email,
			Role:        user.Role,
			PhoneNumber: user.PhoneNumber,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}
		return c.JSON(http.StatusOK, res)
	}
}

func (s *Service) handleGetUserByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		requester := delivery.GetAuthUserFromCtx(ctx)

		logger := logrus.WithFields(logrus.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"requesterID": requester.ID,
		})

		userID, err := uuid.FromString(c.Param("userID"))
		if err != nil {
			logger.Error(err)
			return ErrInternal
		}

		user, err := s.userUsecase.FindByID(ctx, requester, userID)
		switch err {
		case nil:
			break
		case usecase.ErrNotFound:
			return ErrNotFound
		case usecase.ErrPermissionDenied:
			return ErrPermissionDenied
		default:
			logger.Error(err)
			return ErrInternal
		}

		res := userResponse{
			ID:          user.ID.String(),
			Name:        user.Name,
			Email:       user.Email,
			Role:        user.Role,
			PhoneNumber: user.PhoneNumber,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}

		return c.JSON(http.StatusOK, res)
	}
}
