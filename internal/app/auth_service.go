package app

import (
	"context"
	"errors"
	"mrs/internal/domain/user"
	"mrs/internal/utils"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, username string, password string) (string, error)
}

type authService struct {
	userRepo   user.UserRepository
	hasher     utils.PasswordHasher
	jwtManager utils.JWTManager
	logger     applog.Logger
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

func (s *authService) Login(ctx context.Context, username string, password string) (string, error) {
	logger := s.logger.With(applog.String("Method", "authService.Login"))

	// 查询用户
	usr, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not Found", applog.String("username", username))
			return "", errors.New("invalid username or password")
		}
		logger.Error("failed to find user by name", applog.String("username", username), applog.Error(err))
	}

	// 密码校验
	match, err := s.hasher.Check(usr.PasswordHash, password)
	if err != nil {
		logger.Error("password check failed with an internal error", applog.String("username", username), applog.Error(err))
		return "", errors.New("authentication failed due to an internal error")
	}

	if !match {
		logger.Warn("password not match")
		return "", errors.New("invalid username or password")
	}

	// 生成JWT
	token, err := s.jwtManager.GenerateToken(usr.ID, usr.Username, usr.Role.Name)
	if err != nil {
		logger.Error("failed to generater JWT", applog.String("username", username), applog.Error(err))
		return "", errors.New("failed to generate authentication token")
	}

	return token, nil
}
