package service

import (
	"context"
	"regexp"

	"github.com/remiehneppo/material-management/internal/repository"
	"github.com/remiehneppo/material-management/types"
)

type LoginService interface {
	Login(ctx context.Context, req types.LoginRequest) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context) error
	Refresh(ctx context.Context, oldRefreshToken string) (accessToken, refreshToken string, err error)
}

type loginService struct {
	jwtService JWTService
	userRepo   repository.UserRepository
}

func NewLoginService(jwtService JWTService, userRepo repository.UserRepository) LoginService {
	return &loginService{
		jwtService: jwtService,
		userRepo:   userRepo,
	}
}

func (s *loginService) Login(ctx context.Context, req types.LoginRequest) (accessToken, refreshToken string, err error) {

	if !regexp.MustCompile(types.USERNAME_REGEX).MatchString(req.Username) {
		return "", "", types.ErrUsernameInvalid
	}
	if !regexp.MustCompile(types.PASSWORD_REGEX).MatchString(req.Password) {
		return "", "", types.ErrPasswordInvalid
	}
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return "", "", err
	}
	// Generate tokens
	refreshToken, err = s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	accessToken, err = s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *loginService) Logout(ctx context.Context) error {
	// Invalidate the refresh token in the database
	// This is a placeholder implementation
	return nil
}

func (s *loginService) Refresh(ctx context.Context, oldRefreshToken string) (accessToken, refreshToken string, err error) {
	user, err := s.jwtService.ValidateRefreshToken(oldRefreshToken)
	if err != nil {
		return "", "", err
	}

	// Generate new tokens
	refreshToken, err = s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	accessToken, err = s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
