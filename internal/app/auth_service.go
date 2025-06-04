package app

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/user"
	"mrs/internal/utils"
	applog "mrs/pkg/log"
	"time"
)

type AuthService interface {
	Login(ctx context.Context, username string, password string) (*AuthResult, error)
}

type authService struct {
	uow        shared.UnitOfWork
	userRepo   user.UserRepository
	hasher     utils.PasswordHasher
	jwtManager utils.JWTManager
	logger     applog.Logger
}

// AuthResult 定义认证服务返回的统一数据结构
type AuthResult struct {
	Token     string
	User      *user.User
	ExpiresAt time.Time
}

func NewAuthService(
	uow shared.UnitOfWork,
	userRepo user.UserRepository,
	hasher utils.PasswordHasher,
	jwtManager utils.JWTManager,
	logger applog.Logger,
) AuthService {
	return &authService{
		uow:        uow,
		userRepo:   userRepo,
		hasher:     hasher,
		jwtManager: jwtManager,
		logger:     logger.With(applog.String("Service", "AuthService")),
	}
}

func (s *authService) Login(ctx context.Context, username string, password string) (*AuthResult, error) {
	logger := s.logger.With(applog.String("Method", "Login"), applog.String("username", username))

	// 查询用户
	usr, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn("user not found")
			return nil, user.ErrUserNotFound
		}
		if errors.Is(err, user.ErrUserAlreadyExists) {
			logger.Warn("user already exists")
			return nil, user.ErrUserAlreadyExists
		}
		logger.Error("failed to find user by name", applog.String("username", username), applog.Error(err))
		return nil, fmt.Errorf("authService.Login: %w", err)
	}

	// 密码校验
	match, err := s.hasher.Check(usr.PasswordHash, password)
	if err != nil {
		logger.Error("password check failed with an internal error", applog.String("username", username), applog.Error(err))
		return nil, errors.New("authentication failed due to an internal error")
	}

	if !match {
		logger.Warn("password not match")
		return nil, user.ErrInvalidPassword
	}

	// 生成JWT
	token, err := s.jwtManager.GenerateToken(uint(usr.ID), usr.Username, usr.Role.Name)
	if err != nil {
		logger.Error("failed to generater JWT", applog.String("username", username), applog.Error(err))
		return nil, errors.New("failed to generate authentication token")
	}

	claims, err := s.jwtManager.GetMetadata(token)
	if err != nil {
		logger.Error("failed to get JWT metadata", applog.Error(err))
		return nil, errors.New(("failed to get authentication token metadata"))
	}

	return &AuthResult{
		Token:     token,
		User:      usr,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
