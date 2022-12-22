package httpsvc

import (
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"math"
	"net/http"
	"user-service/internal/delivery"
	"user-service/internal/model"
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
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

func (s *Service) handleUpdateUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		requester := delivery.GetAuthUserFromCtx(ctx)
		logger := logrus.WithFields(logrus.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"requesterID": requester.ID,
		})

		req := model.UpdateUserInput{}
		if err := c.Bind(&req); err != nil {
			logrus.Error(err)
			return ErrInvalidArgument
		}

		userID := utils.StringToInt64(c.Param("userID"))
		user, err := s.userUsecase.UpdateByID(ctx, requester, userID, req)
		switch err {
		case nil:
			break
		case usecase.ErrPermissionDenied:
			return ErrPermissionDenied
		case usecase.ErrNotFound:
			return ErrNotFound
		case usecase.ErrRoleNotFound:
			return ErrRoleNotFound
		default:
			logger.Error(err)
			return ErrInternal
		}

		res := userResponse{
			ID:          utils.Int64ToString(user.ID),
			Name:        user.Name,
			Email:       user.Email,
			Role:        user.Role,
			PhoneNumber: user.PhoneNumber,
			CreatedBy:   utils.Int64ToString(user.CreatedBy),
			UpdatedBy:   utils.Int64ToString(user.UpdatedBy),
			CreatedAt:   utils.FormatTimeRFC3339(&user.CreatedAt),
			UpdatedAt:   utils.FormatTimeRFC3339(&user.UpdatedAt),
		}
		return c.JSON(http.StatusOK, res)
	}
}

func (s *Service) handleChangeUserPassword() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		requester := delivery.GetAuthUserFromCtx(ctx)
		logger := logrus.WithFields(logrus.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"requesterID": requester.ID,
		})

		req := model.ChangePasswordInput{}
		if err := c.Bind(&req); err != nil {
			logrus.Error(err)
			return ErrInvalidArgument
		}

		userID := utils.StringToInt64(c.Param("userID"))
		_, err := s.userUsecase.ChangePasswordByID(ctx, requester, userID, req)
		switch err {
		case nil:
			break
		case usecase.ErrPermissionDenied:
			return ErrPermissionDenied
		case usecase.ErrPasswordMismatch:
			return ErrPasswordMismatch
		case usecase.ErrNotFound:
			return ErrNotFound
		default:
			logger.Error(err)
			return ErrInternal
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "success updated password",
		})
	}
}

func (s *Service) handleDeleteUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		requester := delivery.GetAuthUserFromCtx(ctx)
		logger := logrus.WithFields(logrus.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"requesterID": requester.ID,
		})

		userID := utils.StringToInt64(c.Param("userID"))
		user, err := s.userUsecase.DeleteByID(ctx, requester, userID)
		switch err {
		case nil:
			break
		case usecase.ErrPermissionDenied:
			return ErrPermissionDenied
		case usecase.ErrNotFound:
			return ErrNotFound
		case usecase.ErrFailedPrecondition:
			return ErrFailedPrecondition
		default:
			logger.Error(err)
			return ErrInternal
		}

		res := userResponse{
			ID:          utils.Int64ToString(user.ID),
			Name:        user.Name,
			Email:       user.Email,
			Role:        user.Role,
			PhoneNumber: user.PhoneNumber,
			CreatedBy:   utils.Int64ToString(user.CreatedBy),
			UpdatedBy:   utils.Int64ToString(user.UpdatedBy),
			CreatedAt:   utils.FormatTimeRFC3339(&user.CreatedAt),
			UpdatedAt:   utils.FormatTimeRFC3339(&user.UpdatedAt),
		}
		return c.JSON(http.StatusOK, res)
	}
}

func (s *Service) handleCreateUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		requester := delivery.GetAuthUserFromCtx(ctx)
		logger := logrus.WithFields(logrus.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"requesterID": requester.ID,
		})

		req := model.CreateUserInput{}
		if err := c.Bind(&req); err != nil {
			logrus.Error(err)
			return ErrInvalidArgument
		}

		user, err := s.userUsecase.Create(ctx, requester, req)
		switch err {
		case nil:
			break
		case usecase.ErrPermissionDenied:
			return ErrPermissionDenied
		case usecase.ErrRoleNotFound:
			return ErrRoleNotFound
		case usecase.ErrDuplicateEmail:
			return ErrDuplicateEmail
		default:
			logger.Error(err)
			return ErrInternal
		}

		res := userResponse{
			ID:          utils.Int64ToString(user.ID),
			Name:        user.Name,
			Email:       user.Email,
			Role:        user.Role,
			PhoneNumber: user.PhoneNumber,
			CreatedBy:   utils.Int64ToString(user.CreatedBy),
			UpdatedBy:   utils.Int64ToString(user.UpdatedBy),
			CreatedAt:   utils.FormatTimeRFC3339(&user.CreatedAt),
			UpdatedAt:   utils.FormatTimeRFC3339(&user.UpdatedAt),
		}
		return c.JSON(http.StatusCreated, res)
	}
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
			ID:          utils.Int64ToString(user.ID),
			Name:        user.Name,
			Email:       user.Email,
			Role:        user.Role,
			PhoneNumber: user.PhoneNumber,
			CreatedAt:   utils.FormatTimeRFC3339(&user.CreatedAt),
			UpdatedAt:   utils.FormatTimeRFC3339(&user.UpdatedAt),
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

		userID := utils.StringToInt64(c.Param("userID"))
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
		logrus.Warn("WOOYYYY:", utils.Dump(user))
		res := userResponse{
			ID:          utils.Int64ToString(user.ID),
			Name:        user.Name,
			Email:       user.Email,
			Role:        user.Role,
			PhoneNumber: user.PhoneNumber,
			CreatedAt:   utils.FormatTimeRFC3339(&user.CreatedAt),
			UpdatedAt:   utils.FormatTimeRFC3339(&user.UpdatedAt),
		}

		return c.JSON(http.StatusOK, res)
	}
}

func (s *Service) handleGetAllUsers() echo.HandlerFunc {
	type CursorInfo struct {
		Size       int64  `json:"size"`
		Count      int64  `json:"count"`
		CountPage  int64  `json:"countPage"`
		HasMore    bool   `json:"hasMore"`
		Cursor     string `json:"cursor"`
		NextCursor string `json:"nextCursor"`
	}

	type userCursor struct {
		Edges      []userResponse `json:"edges"`
		CursorInfo *CursorInfo    `json:"cursor_info"`
	}
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		requester := delivery.GetAuthUserFromCtx(ctx)

		logger := logrus.WithFields(logrus.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"requesterID": requester.ID,
		})

		page := utils.StringToInt64(c.QueryParam("page"))
		size := utils.StringToInt64(c.QueryParam("size"))
		criterias := model.UserSearchCriteria{
			Query: c.QueryParam("query"),
			Page:  page,
			Size:  size,
		}

		users, count, err := s.userUsecase.FindAll(ctx, requester, criterias)
		switch err {
		case nil:
			break
		case usecase.ErrPermissionDenied:
			return ErrPermissionDenied
		case usecase.ErrNotFound:
			return ErrNotFound
		default:
			logger.Error(err)
			return ErrInternal
		}

		var userResponses []userResponse
		for _, user := range users {
			userResponses = append(userResponses, userResponse{
				ID:          utils.Int64ToString(user.ID),
				Name:        user.Name,
				Email:       user.Email,
				Role:        user.Role,
				PhoneNumber: user.PhoneNumber,
				CreatedAt:   utils.FormatTimeRFC3339(&user.CreatedAt),
				UpdatedAt:   utils.FormatTimeRFC3339(&user.UpdatedAt),
			})
		}
		hasMore := count-(criterias.Page*criterias.Size) > 0
		countPage := int64(math.Ceil(float64(count) / float64(criterias.Size)))
		res := userCursor{
			Edges: userResponses,
			CursorInfo: &CursorInfo{
				Size:      size,
				Count:     count,
				CountPage: countPage,
				HasMore:   hasMore,
				Cursor:    utils.Int64ToString(page),
			},
		}
		if !hasMore {
			res.CursorInfo.NextCursor = ""
		} else {
			res.CursorInfo.NextCursor = utils.Int64ToString(page + 1)
		}

		return c.JSON(http.StatusOK, res)
	}
}
