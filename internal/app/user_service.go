package app

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/user"
	"mrs/internal/utils"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type UserService interface {
	RegisterUser(ctx context.Context, req *request.RegisterUserRequest) (*response.UserResponse, error)
	GetUserByID(ctx context.Context, req *request.GetUserRequest) (*response.UserResponse, error)
}

type userService struct {
	defaultRoleName string // 注入配置依赖：默认角色
	uow             shared.UnitOfWork
	userRepo        user.UserRepository
	roleRepo        user.RoleRepository
	hasher          utils.PasswordHasher
	logger          applog.Logger
}

func NewUserService(
	defaultRoleName string,
	uow shared.UnitOfWork,
	userRepo user.UserRepository,
	roleRepo user.RoleRepository,
	hasher utils.PasswordHasher,
	logger applog.Logger,
) UserService {
	return &userService{
		defaultRoleName: defaultRoleName,
		uow:             uow,
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		hasher:          hasher,
		logger:          logger.With(applog.String("Service", "UserService")),
	}
}

func (s *userService) RegisterUser(ctx context.Context, req *request.RegisterUserRequest) (*response.UserResponse, error) {
	logger := s.logger.With(applog.String("Method", "RegisterUser"),
		applog.String("username", req.Username),
		applog.String("email", req.Email))
	// 数据库底层存在用户名和邮箱的唯一性约束，因此不需要再验证

	// 查找默认角色，无需事务，因为角色不会被修改
	if req.DefaultRole == "" {
		req.DefaultRole = s.defaultRoleName
	}
	defaultRole, err := s.roleRepo.FindByName(ctx, req.DefaultRole)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("Default role not found", applog.String("role_name", req.DefaultRole))
			return nil, fmt.Errorf("default role not found: %w", user.ErrRoleNotFound)
		}
		logger.Error("Failed to find default role", applog.Error(err))
		return nil, err
	}

	// 创建用户实体并生成哈希密码
	newUser := user.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     defaultRole,
	}

	if err := newUser.SetPassword(req.Password, s.hasher); err != nil {
		logger.Error("failed to hash password", applog.Error(err))
		return nil, fmt.Errorf("%w: %w", user.ErrInvalidPassword, err)
	}

	// 开启事务
	err = s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		if err := provider.GetUserRepository().Create(ctx, &newUser); err != nil {
			logger.Error("failed to create user in repository", applog.Error(err))
			return fmt.Errorf("userService.RegisterUser: %w", err)
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to execute transaction", applog.Error(err))
		return nil, fmt.Errorf("userService.RegisterUser: %w", err)
	}

	logger.Info("create user successful")
	return response.ToUserResponse(&newUser), nil
}

func (s *userService) GetUserByID(ctx context.Context, req *request.GetUserRequest) (*response.UserResponse, error) {
	logger := s.logger.With(applog.String("Method", "GetUserByID"), applog.Uint("user_id", req.ID))
	usr, err := s.userRepo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn("user not found")
			return nil, fmt.Errorf("%w: %w", user.ErrUserNotFound, err)
		}
		logger.Error("failed to find user by ID", applog.Error(err))
		return nil, fmt.Errorf("userService.GetUserByID: %w", err)
	}

	logger.Info("find user successfully")
	return response.ToUserResponse(usr), nil
}
