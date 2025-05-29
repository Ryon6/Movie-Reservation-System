// TODO: 配置依赖注入规范化
package app

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/role"
	"mrs/internal/domain/user"
	"mrs/internal/utils"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type UserService interface {
	RegisterUser(ctx context.Context, username, email, plainPassword, defaultRoleName string) (*user.User, error)
	GetUserByID(ctx context.Context, userID uint) (*user.User, error)
}

type userService struct {
	defaultRoleName string // 注入配置依赖：默认角色
	userRepo        user.UserRepository
	roleRepo        role.RoleRepository
	hasher          utils.PasswordHasher
	logger          applog.Logger
}

func NewUserService(
	defaultRoleName string,
	userRepo user.UserRepository,
	roleRepo role.RoleRepository,
	hasher utils.PasswordHasher,
	logger applog.Logger,
) UserService {
	return &userService{
		defaultRoleName: defaultRoleName,
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		hasher:          hasher,
		logger:          logger,
	}
}

func (s *userService) RegisterUser(ctx context.Context, username, email, plainPassword, defaultRoleName string) (*user.User, error) {
	logger := s.logger.With(applog.String("Method", "userService.RegisterUser"),
		applog.String("username", username),
		applog.String("email", email))
	// 验证用户名是否存在
	_, err := s.userRepo.FindByUsername(ctx, username)
	if err == nil {
		// 如果err为nil说明用户名已存在
		logger.Warn("username already exists")
		return nil, errors.New("username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("failed to check username existence", applog.Error(err))
		return nil, fmt.Errorf("userService.RegisterUser: %w", err)
	}

	// 验证邮箱是否存在
	_, err = s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		logger.Warn("email already exists")
		return nil, errors.New("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("failed to check email existence", applog.Error(err))
		return nil, fmt.Errorf("userService.RegisterUser: %w", err)
	}

	// 查找默认角色
	if defaultRoleName == "" {
		defaultRoleName = s.defaultRoleName
	}
	defaultRole, err := s.roleRepo.FindByName(ctx, defaultRoleName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("Default role not found", applog.String("role_name", defaultRoleName))
			// return nil, domainErrors.NewNotFoundError("default role '%s' not found", defaultRoleName)
			return nil, errors.New("default role not found")
		}
		logger.Error("Failed to find default role", applog.Error(err))
		return nil, err
	}

	// 创建用户实体并生成哈希密码
	newUser := user.User{
		Username: username,
		Email:    email,
		RoleID:   defaultRole.ID,
		Role:     *defaultRole,
	}

	if err := newUser.SetPassword(plainPassword); err != nil {
		logger.Error("failed to hash password", applog.Error(err))
		return nil, errors.New("failed to process password")
	}

	// 保存用户
	if err = s.userRepo.Create(ctx, &newUser); err != nil {
		logger.Error("failed to create user in repository", applog.Error(err))
		// 可能需要处理数据库唯一约束错误，并将其转换为领域错误
		return nil, fmt.Errorf("userService.RegisterUser: %w", err)
	}
	logger.Info("create user successful")
	return &newUser, nil
}

func (s *userService) GetUserByID(ctx context.Context, userID uint) (*user.User, error) {
	logger := s.logger.With(applog.String("Method", "userService.GetUserByID"), applog.Uint("user_id", userID))
	usr, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not found")
			return nil, errors.New("user not found")
		}
		logger.Error("failed to find user by ID", applog.Error(err))
		return nil, fmt.Errorf("userService.GetUserByID: %w", err)
	}

	logger.Info("find user successfully")
	return usr, nil
}
