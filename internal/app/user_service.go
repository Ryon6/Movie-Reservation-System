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
	Register(ctx context.Context, req *request.RegisterUserRequest) (*response.UserProfileResponse, error)
	GetUserProfile(ctx context.Context, req *request.GetUserRequest) (*response.UserProfileResponse, error)
	UpdateUserProfile(ctx context.Context, req *request.UpdateUserRequest) (*response.UserProfileResponse, error)
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

func (s *userService) Register(ctx context.Context, req *request.RegisterUserRequest) (*response.UserProfileResponse, error) {
	logger := s.logger.With(applog.String("Method", "RegisterUser"),
		applog.String("username", req.Username),
		applog.String("email", req.Email))
	// 数据库底层存在用户名和邮箱的唯一性约束，因此不需要再验证

	// 查找默认角色，无需事务，因为角色不会被修改
	defaultRole, err := s.roleRepo.FindByName(ctx, s.defaultRoleName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("Default role not found", applog.String("role_name", s.defaultRoleName))
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
	return response.ToUserProfileResponse(&newUser), nil
}

func (s *userService) GetUserProfile(ctx context.Context, req *request.GetUserRequest) (*response.UserProfileResponse, error) {
	logger := s.logger.With(applog.String("Method", "GetUserProfile"), applog.Uint("user_id", req.ID))
	usr, err := s.userRepo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn("user not found")
			return nil, err
		}
		logger.Error("failed to find user by ID", applog.Error(err))
		return nil, err
	}

	logger.Info("find user successfully")
	return response.ToUserProfileResponse(usr), nil
}

func (s *userService) UpdateUserProfile(ctx context.Context, req *request.UpdateUserRequest) (*response.UserProfileResponse, error) {
	logger := s.logger.With(applog.String("Method", "UpdateUserProfile"), applog.Uint("user_id", req.ID))
	usr := req.ToDomain()
	if req.Password != "" {
		if err := usr.SetPassword(req.Password, s.hasher); err != nil {
			logger.Error("failed to hash password", applog.Error(err))
			return nil, err
		}
	}

	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		// 先检查用户是否存在
		existingUser, err := provider.GetUserRepository().FindByID(ctx, uint(usr.ID))
		if err != nil {
			if errors.Is(err, user.ErrUserNotFound) {
				logger.Warn("user not found")
				return err
			}
			logger.Error("failed to find user by ID", applog.Error(err))
			return err
		}

		// 检查是否需要更新
		if existingUser.Username == usr.Username &&
			existingUser.Email == usr.Email &&
			existingUser.PasswordHash == usr.PasswordHash {
			logger.Info("no need to update user")
			return shared.ErrNoRowsAffected
		}

		// 更新用户
		if err := provider.GetUserRepository().Update(ctx, usr); err != nil {
			logger.Error("failed to update user in repository", applog.Error(err))
			return err
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, shared.ErrNoRowsAffected) {
			logger.Info("no need to update user")
			return nil, err
		}
		logger.Error("failed to execute transaction", applog.Error(err))
		return nil, err
	}

	logger.Info("update user successfully")
	return response.ToUserProfileResponse(usr), nil
}
