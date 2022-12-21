package httpsvc

import (
	"github.com/labstack/echo"
	"user-service/auth"
	"user-service/internal/model"
)

// Service http service
type Service struct {
	echo           *echo.Group
	authUsecase    model.AuthUsecase
	userUsecase    model.UserUsecase
	httpMiddleware *auth.AuthenticationMiddleware
}

// RouteService add dependencies and use group for routing
func RouteService(
	echo *echo.Group,
	authUsecase model.AuthUsecase,
	userUsecase model.UserUsecase,
	authMiddleware *auth.AuthenticationMiddleware,
) {
	srv := &Service{
		echo:           echo,
		authUsecase:    authUsecase,
		userUsecase:    userUsecase,
		httpMiddleware: authMiddleware,
	}
	srv.initRoutes()
}

func (s *Service) initRoutes() {
	// auth
	authRoute := s.echo.Group("/auth")
	authRoute.GET("/me/", s.handleGetCurrentLoginUser(), s.httpMiddleware.MustAuthenticateAccessToken())
	authRoute.POST("/login/", s.handleLoginByEmailPassword())
	authRoute.POST("/refresh/", s.handleRefreshToken())
	authRoute.POST("/logout/", s.handleLogout(), s.httpMiddleware.MustAuthenticateAccessToken())

	userRoute := s.echo.Group("/user")
	userRoute.GET("/", s.handleGetAllUsers(), s.httpMiddleware.MustAuthenticateAccessToken())
	userRoute.POST("/", s.handleCreateUser(), s.httpMiddleware.MustAuthenticateAccessToken())
	userRoute.PUT("/:userID/", s.handleUpdateUser(), s.httpMiddleware.MustAuthenticateAccessToken())
	userRoute.GET("/:userID/", s.handleGetUserByID(), s.httpMiddleware.MustAuthenticateAccessToken())
	userRoute.DELETE("/:userID/", s.handleDeleteUser(), s.httpMiddleware.MustAuthenticateAccessToken())
}
