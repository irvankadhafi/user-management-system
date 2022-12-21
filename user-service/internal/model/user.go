package model

import (
	"context"
	"errors"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"time"
	"user-service/internal/helper"
	"user-service/rbac"
)

// ErrPasswordMismatch error
var ErrPasswordMismatch = errors.New("password mismatch")

// User :nodoc:
type User struct {
	ID          uuid.UUID `json:"id" gorm:"primary_key"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Password    string    `json:"password" gorm:"->:false;<-"` // gorm create & update only (disabled read from db)
	Role        rbac.Role `json:"role"`
	PhoneNumber string    `json:"phone_number"`

	CreatedBy uuid.UUID      `json:"created_by" gorm:"->;<-:create"` // create & read only
	UpdatedBy uuid.UUID      `json:"updated_by"`
	CreatedAt time.Time      `json:"created_at" gorm:"->;<-:create"` // create & read only
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	SessionID uuid.UUID            `json:"session_id" gorm:"-"`
	rolePerm  *rbac.RolePermission `gorm:"-"`
}

// Gender the user's gender
type Gender string

type UserRepository interface {
	Create(ctx context.Context, userID uuid.UUID, user *User) error
	Update(ctx context.Context, userID uuid.UUID, user *User) (*User, error)
	UpdatePasswordByID(ctx context.Context, userID uuid.UUID, password string) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)

	IsLoginByEmailPasswordLocked(ctx context.Context, email string) (bool, error)
	IncrementLoginByEmailPasswordRetryAttempts(ctx context.Context, email string) error
	FindPasswordByID(ctx context.Context, id uuid.UUID) ([]byte, error)
}

type UserUsecase interface {
	FindByID(ctx context.Context, requester *User, id uuid.UUID) (*User, error)
	Create(ctx context.Context, requester *User, input CreateUserInput) (*User, error)
	ChangePassword(ctx context.Context, requester *User, input ChangePasswordInput) (*User, error)
	UpdateProfile(ctx context.Context, requester *User, input UpdateProfileInput) (*User, error)
}

// SetPermission set permission to user
func (u *User) SetPermission(perm *rbac.Permission) {
	if perm == nil {
		return
	}
	u.rolePerm = rbac.NewRolePermission(u.Role, perm)
}

// SetRolePermission set role permission to user
func (u *User) SetRolePermission(rolePerm *rbac.RolePermission) {
	u.rolePerm = rolePerm
}

// GetRolePermission get the role permission
func (u *User) GetRolePermission() *rbac.RolePermission {
	return u.rolePerm
}

type CreateUserInput struct {
	Name                 string `json:"name" validate:"required"`
	Email                string `json:"email" validate:"required,email"`
	Password             string `json:"password" validate:"required,min=6"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,min=6,eqfield=Password"`
	PhoneNumber          string `json:"phone_number" validate:"required,phonenumber"`
}

// ValidateAndFormat validate and format the phone number
func (c *CreateUserInput) ValidateAndFormat() error {
	_ = helper.RemoveLeadingZeroPhoneNumber(&c.PhoneNumber)
	if err := validate.Struct(c); err != nil {
		return err
	}

	if err := helper.FormatPhoneNumberWithCountryCode(&c.PhoneNumber, "ID"); err != nil {
		return err
	}

	return nil
}

type UpdateProfileInput struct {
	Name        string `json:"name" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required,phonenumber"`
}

func (u *UpdateProfileInput) ValidateAndFormat() error {
	_ = helper.RemoveLeadingZeroPhoneNumber(&u.PhoneNumber)
	if err := validate.Struct(u); err != nil {
		return err
	}

	if err := helper.FormatPhoneNumberWithCountryCode(&u.PhoneNumber, "ID"); err != nil {
		return err
	}

	return nil
}

// ChangePasswordInput change user password
type ChangePasswordInput struct {
	OldPassword             string `json:"old_password" validate:"required,min=6"`
	NewPassword             string `json:"new_password" validate:"required,min=6"`
	NewPasswordConfirmation string `json:"new_password_confirmation" validate:"required,min=6,eqfield=NewPassword"`
}

// Validate validate user's password & input body
func (c *ChangePasswordInput) Validate() error {
	if c.NewPassword != c.NewPasswordConfirmation {
		return ErrPasswordMismatch
	}

	return validate.Struct(c)
}

// HasAccess check authorization
func (u *User) HasAccess(resource rbac.Resource, action rbac.Action) bool {
	if u.rolePerm == nil {
		return false
	}

	return u.rolePerm.HasAccess(resource, action)
}
