package service

import (
	"context"
	"regexp"

	"github.com/remiehneppo/material-management/internal/repository"
	"github.com/remiehneppo/material-management/types"
	"github.com/remiehneppo/material-management/utils"
)

type UserService interface {
	GetProfile(ctx context.Context) (*types.User, error)
	Update(ctx context.Context, req *types.UpdateUserInfoRequest) error
	UpdatePassword(ctx context.Context, req *types.UpdateUserPasswordRequest) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetProfile(ctx context.Context) (*types.User, error) {
	userCache, ok := ctx.Value("user").(*types.User)
	if !ok {
		return nil, types.ErrUserNotFound
	}
	user, err := s.userRepo.FindByID(ctx, userCache.ID)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (s *userService) Update(ctx context.Context, req *types.UpdateUserInfoRequest) error {
	userCache, ok := ctx.Value("user").(*types.User)
	if !ok {
		return types.ErrUserNotFound
	}
	user, err := s.userRepo.FindByID(ctx, userCache.ID)
	if err != nil {
		return err
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Workspace != "" {
		user.Workspace = req.Workspace
	}
	if req.WorkspaceRole != "" {
		user.WorkspaceRole = req.WorkspaceRole
	}
	user.ID = ""
	user.UpdateAt = utils.GetCurrentTimestampSecond()

	return s.userRepo.Update(ctx, userCache.ID, user)
}

func (s *userService) UpdatePassword(ctx context.Context, req *types.UpdateUserPasswordRequest) error {
	userCache, ok := ctx.Value("user").(*types.User)
	if !ok {
		return types.ErrUserNotFound
	}
	user, err := s.userRepo.FindByID(ctx, userCache.ID)
	if err != nil {
		return err
	}
	if req.OldPassword != user.Password {
		return types.ErrPasswordIncorrect
	}
	if regexp.MustCompile(types.PASSWORD_REGEX).MatchString(req.NewPassword) == false {
		return types.ErrPasswordInvalid
	}
	user.Password = req.NewPassword
	user.UpdateAt = utils.GetCurrentTimestampSecond()
	user.ID = ""

	return s.userRepo.Update(ctx, userCache.ID, user)
}
