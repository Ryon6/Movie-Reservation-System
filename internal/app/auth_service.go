package app

import (
	"context"
	"errors"
	"mrs/internal/app/dto"
	"mrs/internal/domain/user"
	"mrs/internal/utils"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, username string, password string) (*dto.AuthResult, error)
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

func (s *authService) Login(ctx context.Context, username string, password string) (*dto.AuthResult, error) {
	logger := s.logger.With(applog.String("Method", "authService.Login"))

	// 查询用户
	usr, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not Found", applog.String("username", username))
			return nil, errors.New("invalid username or password")
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
		return nil, errors.New("invalid username or password")
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

	return &dto.AuthResult{
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
