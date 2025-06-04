package app

import (
	"context"
	"errors"
	"mrs/internal/domain/user"
	"mrs/internal/utils"
	applog "mrs/pkg/log"
	"time"

	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, username string, password string) (*AuthResult, error)
}

type authService struct {
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
	userRepo user.UserRepository,
	hasher utils.PasswordHasher,
	jwtManager utils.JWTManager,
	logger applog.Logger,
) AuthService {
	return &authService{
		userRepo:   userRepo,
		hasher:     hasher,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (s *authService) Login(ctx context.Context, username string, password string) (*AuthResult, error) {
	logger := s.logger.With(applog.String("Method", "authService.Login"))

	// 查询用户
	usr, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, user.ErrUserAlreadyExists) {
			logger.Warn("user not Found", applog.String("username", username))
			return nil, user.ErrUserAlreadyExists
		}
		logger.Error("failed to find user by name", applog.String("username", username), applog.Error(err))
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
