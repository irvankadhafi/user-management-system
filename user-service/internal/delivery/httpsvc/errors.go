package httpsvc

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"net/http"
)

var (
	ErrInvalidArgument            = echo.NewHTTPError(http.StatusBadRequest, "invalid argument")
	ErrEmailPasswordNotMatch      = echo.NewHTTPError(http.StatusUnauthorized, "email or password not match")
	ErrLoginByEmailPasswordLocked = echo.NewHTTPError(http.StatusLocked, "user is locked from logging in using email and password")
	ErrPermissionDenied           = echo.NewHTTPError(http.StatusForbidden, "permission denied")
	ErrInternal                   = echo.NewHTTPError(http.StatusInternalServerError, "internal system error")
	ErrUnauthenticated            = echo.NewHTTPError(http.StatusUnauthorized, "unauthenticated")
	ErrNotFound                   = echo.NewHTTPError(http.StatusNotFound, "record not found")
	ErrDuplicateEmail             = echo.NewHTTPError(http.StatusBadRequest, "email already exist ")
	ErrFailedPrecondition         = echo.NewHTTPError(http.StatusBadRequest, "precondition failed ")
)

// httpValidationOrInternalErr return valdiation or internal error
func httpValidationOrInternalErr(err error) error {
	// Memeriksa apakah err merupakan validator.ValidationErrors
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// Jika tidak ada kesalahan validasi, mengembalikan kesalahan internal
		return ErrInternal
	}

	// Mengubah validator.ValidationErrors menjadi map dengan kunci field dan nilai pesan kesalahan
	fields := make(map[string]string)
	for _, validationError := range validationErrors {
		// Mengambil tag yang digunakan untuk menyebabkan kesalahan validasi
		tag := validationError.Tag()
		// Menambahkan kesalahan validasi ke map
		fields[validationError.Field()] = fmt.Sprintf("Failed on the '%s' tag", tag)
	}

	// Mengembalikan kesalahan HTTP dengan status bad request dan daftar field yang gagal validasi
	return echo.NewHTTPError(http.StatusBadRequest, fields)
}
