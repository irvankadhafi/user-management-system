package usecase

import (
	"context"
	"github.com/sirupsen/logrus"
	"sync"
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

func (u *userUsecase) FindByID(ctx context.Context, requester *model.User, id int64) (*model.User, error) {
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

func (u *userUsecase) FindAll(ctx context.Context, requester *model.User, criteria model.UserSearchCriteria) ([]*model.User, int64, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.Dump(ctx),
		"requester": utils.Dump(requester),
	})

	if !requester.HasAccess(rbac.ResourceUser, rbac.ActionViewAny) {
		return nil, 0, ErrPermissionDenied
	}

	userIDs, count, err := u.userRepo.FindAll(ctx, criteria.Query, criteria.Size, criteria.Page)
	if err != nil {
		logger.Error(err)
		return nil, 0, err
	}
	if len(userIDs) <= 0 || count <= 0 {
		return nil, 0, err
	}

	var wg sync.WaitGroup
	c := make(chan *model.User, len(userIDs))
	for _, id := range userIDs {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()

			user, err := u.FindByID(ctx, requester, id)
			if err != nil { // Ignore error
				return
			}
			c <- user
		}(id)
	}
	wg.Wait()
	close(c)

	// Put 'em in a buffer
	rs := map[int64]*model.User{}
	for s := range c {
		if s != nil {
			rs[s.ID] = s
		}
	}

	var users []*model.User
	// Sort 'em out
	for _, id := range userIDs {
		if user, ok := rs[id]; ok {
			users = append(users, user)
		}
	}

	return users, count, nil
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

	if validRole := rbac.ValidateRole(input.Role); !validRole {
		return nil, ErrRoleNotFound
	}

	if err := input.ValidateAndFormat(); err != nil {
		logger.Error(err)
		return nil, err
	}

	existingUser, err := u.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if existingUser != nil {
		logger.Warn("email already exist")
		return nil, ErrDuplicateEmail
	}

	cipherPwd, err := helper.HashString(input.Password)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	user := &model.User{
		ID:          utils.GenerateID(),
		Name:        input.Name,
		Email:       input.Email,
		Password:    cipherPwd,
		Role:        input.Role,
		PhoneNumber: input.PhoneNumber,
	}

	if err := u.userRepo.Create(ctx, requester.ID, user); err != nil {
		logger.Error(err)
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) ChangePasswordByID(ctx context.Context, requester *model.User, id int64, input model.ChangePasswordInput) (*model.User, error) {
	if !requester.HasAccess(rbac.ResourceUser, rbac.ActionEditAny) {
		return nil, ErrPermissionDenied
	}

	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"requester": utils.Dump(requester),
	})

	if err := input.Validate(); err != nil {
		logger.Error(err)
		return nil, err
	}

	user, err := u.userRepo.FindByID(ctx, id)
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

	err = u.userRepo.UpdatePasswordByID(ctx, id, cipherPassword)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) UpdateByID(ctx context.Context, requester *model.User, id int64, input model.UpdateUserInput) (*model.User, error) {
	if !requester.HasAccess(rbac.ResourceUser, rbac.ActionEditAny) {
		return nil, ErrPermissionDenied
	}

	logger := logrus.WithFields(logrus.Fields{
		"ctx":          utils.DumpIncomingContext(ctx),
		"existingUser": utils.Dump(input),
	})

	if validRole := rbac.ValidateRole(input.Role); !validRole {
		return nil, ErrRoleNotFound
	}

	if err := input.ValidateAndFormat(); err != nil {
		logger.Error(err)
		return nil, err
	}

	existingUser, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// obfuscate the cause of the error
	if existingUser == nil {
		return nil, ErrNotFound
	}

	existingUser.Name = input.Name
	existingUser.PhoneNumber = input.PhoneNumber
	existingUser.Role = input.Role
	user, err := u.userRepo.Update(ctx, requester.ID, existingUser)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) DeleteByID(ctx context.Context, requester *model.User, id int64) (*model.User, error) {
	if !requester.HasAccess(rbac.ResourceUser, rbac.ActionDeleteAny) {
		return nil, ErrPermissionDenied
	}

	if requester.ID == id {
		return nil, ErrFailedPrecondition
	}

	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"requester": utils.Dump(requester),
		"id":        id,
	})

	user, err := u.FindByID(ctx, requester, id)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	deletedUser, err := u.userRepo.DeleteByID(ctx, requester.ID, id)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	user.DeletedAt = deletedUser.DeletedAt

	return user, nil
}

func (u *userUsecase) findByID(ctx context.Context, id int64) (*model.User, error) {
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
