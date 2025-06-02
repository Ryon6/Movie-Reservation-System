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
	Token     string    `json:"token"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	RoleName  string    `json:"role_name"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	CreateAt  time.Time `json:"create_at"`
	UpdateAt  time.Time `json:"update_at"`
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
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, user.ErrUserExists) {
			logger.Warn("user not Found", applog.String("username", username))
			return nil, user.ErrUserExists
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
	token, err := s.jwtManager.GenerateToken(usr.ID, usr.Username, usr.Role.Name)
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
		UserID:    usr.ID,
		Username:  usr.Username,
		Email:     usr.Email,
		RoleName:  usr.Role.Name,
		Token:     token,
		ExpiresAt: claims.ExpiresAt.Time,
		CreateAt:  usr.CreatedAt,
		UpdateAt:  usr.UpdatedAt,
	}, nil
}
