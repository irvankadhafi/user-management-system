package usecase

import (
	"context"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"user-service/internal/helper"
	"user-service/internal/model"
	"user-service/rbac"
	"user-service/utils"
)

type userUsecase struct {
	userRepo model.UserRepository
}

// NewUserUsecase :nodoc:
func NewUserUsecase(userRepo model.UserRepository) model.UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
	}
}

func (u *userUsecase) FindByID(ctx context.Context, requester *model.User, id uuid.UUID) (*model.User, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.Dump(ctx),
		"requester": utils.Dump(requester),
		"id":        id,
	})

	user, err := u.findByID(ctx, id)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if user.ID == requester.ID {
		return user, nil
	}

	if !requester.HasAccess(rbac.ResourceUser, rbac.ActionViewAny) {
		logger.Error(ErrPermissionDenied)
		return nil, ErrPermissionDenied
	}

	return user, nil
}

func (u *userUsecase) Create(ctx context.Context, requester *model.User, input model.CreateUserInput) (*model.User, error) {
	if !requester.HasAccess(rbac.ResourceUser, rbac.ActionCreateAny) {
		return nil, ErrPermissionDenied
	}

	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"requester": utils.Dump(requester),
		"input":     utils.Dump(input),
	})

	if err := input.ValidateAndFormat(); err != nil {
		logger.Error(err)
		return nil, err
	}

	cipherPwd, err := helper.HashString(input.Password)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	user := &model.User{
		ID:          uuid.NewV4(),
		Name:        input.Name,
		Email:       input.Email,
		Password:    cipherPwd,
		Role:        rbac.RoleMember,
		PhoneNumber: input.PhoneNumber,
	}

	if err := u.userRepo.Create(ctx, requester.ID, user); err != nil {
		logger.Error(err)
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) ChangePassword(ctx context.Context, requester *model.User, input model.ChangePasswordInput) (*model.User, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"requester": utils.Dump(requester),
	})

	if err := input.Validate(); err != nil {
		logger.Error(err)
		return nil, err
	}

	user, err := u.userRepo.FindByID(ctx, requester.ID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// obfuscate the cause of the error
	if user == nil {
		return nil, ErrNotFound
	}

	cipherPwd, err := u.userRepo.FindPasswordByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if cipherPwd == nil {
		return nil, ErrNotFound
	}

	// Check user inputed password same with database
	if !helper.IsHashedStringMatch([]byte(input.OldPassword), cipherPwd) {
		return nil, ErrPasswordMismatch
	}

	cipherPassword, err := helper.HashString(input.NewPassword)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = u.userRepo.UpdatePasswordByID(ctx, requester.ID, cipherPassword)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) UpdateProfile(ctx context.Context, requester *model.User, input model.UpdateProfileInput) (*model.User, error) {
	if !requester.HasAccess(rbac.ResourceUser, rbac.ActionEditAny) {
		return nil, ErrPermissionDenied
	}

	logger := logrus.WithFields(logrus.Fields{
		"ctx":             utils.DumpIncomingContext(ctx),
		"existingEpisode": utils.Dump(input),
	})

	if err := input.ValidateAndFormat(); err != nil {
		logger.Error(err)
		return nil, err
	}

	existingEpisode, err := u.userRepo.FindByID(ctx, requester.ID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// obfuscate the cause of the error
	if existingEpisode == nil {
		return nil, ErrNotFound
	}

	existingEpisode.Name = input.Name
	existingEpisode.PhoneNumber = input.Name
	user, err := u.userRepo.Update(ctx, requester.ID, existingEpisode)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) findByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": utils.DumpIncomingContext(ctx),
		"id":  id,
	})

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}

	return user, nil
}
