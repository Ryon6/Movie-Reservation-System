package app

import (
	"context"
	"errors"
	"fmt"
	"math"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/user"
	"mrs/internal/utils"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type UserService interface {
	Register(ctx context.Context, req *request.RegisterUserRequest) (*response.UserProfileResponse, error) // 注册用户
	GetUser(ctx context.Context, req *request.GetUserRequest) (*response.UserResponse, error)              // 获取用户信息
	UpdateUser(ctx context.Context, req *request.UpdateUserRequest) (*response.UserResponse, error)        // 更新用户信息
	DeleteUser(ctx context.Context, req *request.DeleteUserRequest) error                                  // 删除用户
	ListUsers(ctx context.Context, req *request.ListUserRequest) (*response.ListUserResponse, error)       // 获取用户列表
	CreateRole(ctx context.Context, req *request.CreateRoleRequest) (*response.RoleResponse, error)        // 创建角色
	ListRoles(ctx context.Context) (*response.ListRoleResponse, error)                                     // 获取角色列表
	UpdateRole(ctx context.Context, req *request.UpdateRoleRequest) (*response.RoleResponse, error)        // 更新角色
	DeleteRole(ctx context.Context, req *request.DeleteRoleRequest) error                                  // 删除角色
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

// 注册用户
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

// 获取用户信息
func (s *userService) GetUser(ctx context.Context, req *request.GetUserRequest) (*response.UserResponse, error) {
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
	return response.ToUserResponse(usr), nil
}

// 更新用户信息
func (s *userService) UpdateUser(ctx context.Context, req *request.UpdateUserRequest) (*response.UserResponse, error) {
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
		if (usr.Username == "" || usr.Username == existingUser.Username) &&
			(usr.Email == "" || usr.Email == existingUser.Email) &&
			(usr.PasswordHash == "" || usr.PasswordHash == existingUser.PasswordHash) {
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
	return response.ToUserResponse(usr), nil
}

// 删除用户
func (s *userService) DeleteUser(ctx context.Context, req *request.DeleteUserRequest) error {
	logger := s.logger.With(applog.String("Method", "DeleteUser"), applog.Uint("user_id", req.ID))
	err := s.userRepo.Delete(ctx, req.ID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn("user not found")
			return err
		}
		logger.Error("failed to delete user", applog.Error(err))
		return err
	}

	logger.Info("delete user successfully")
	return nil
}

// 获取用户列表
func (s *userService) ListUsers(ctx context.Context, req *request.ListUserRequest) (*response.ListUserResponse, error) {
	logger := s.logger.With(applog.String("Method", "ListUsers"), applog.Any("request", req))
	options := req.ToDomain()
	users, total, err := s.userRepo.List(ctx, options)
	if err != nil {
		logger.Error("failed to list users", applog.Error(err))
		return nil, err
	}

	pagination := response.PaginationResponse{
		Page:       options.Page,
		PageSize:   options.PageSize,
		TotalCount: int(total),
		TotalPages: int(math.Ceil(float64(total) / float64(options.PageSize))),
	}
	resp := response.ToListUserResponse(users)
	resp.PaginationResponse = pagination
	logger.Info("list users successfully")
	return resp, nil
}

// 创建角色
func (s *userService) CreateRole(ctx context.Context, req *request.CreateRoleRequest) (*response.RoleResponse, error) {
	logger := s.logger.With(applog.String("Method", "CreateRole"), applog.String("role_name", req.Name))
	role := req.ToDomain()

	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		var err error
		role, err = provider.GetRoleRepository().Create(ctx, role)
		if err != nil {
			if errors.Is(err, user.ErrRoleAlreadyExists) {
				logger.Warn("role already exists", applog.String("role_name", req.Name))
				return err
			}
			logger.Error("failed to create role in repository", applog.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to execute transaction", applog.Error(err))
		return nil, err
	}

	logger.Info("create role successfully")
	return response.ToRoleResponse(role), nil
}

// 获取角色列表
func (s *userService) ListRoles(ctx context.Context) (*response.ListRoleResponse, error) {
	logger := s.logger.With(applog.String("Method", "ListRoles"))
	roles, err := s.roleRepo.ListAll(ctx)
	if err != nil {
		logger.Error("failed to list roles", applog.Error(err))
		return nil, err
	}

	logger.Info("list roles successfully")
	return response.ToListRoleResponse(roles), nil
}

// 更新角色
func (s *userService) UpdateRole(ctx context.Context, req *request.UpdateRoleRequest) (*response.RoleResponse, error) {
	logger := s.logger.With(applog.String("Method", "UpdateRole"), applog.Uint("role_id", req.ID))
	role := req.ToDomain()

	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		var err error
		roleRepo := provider.GetRoleRepository()
		// 先检查角色是否存在
		existingRole, err := roleRepo.FindByID(ctx, uint(role.ID))
		if err != nil {
			if errors.Is(err, user.ErrRoleNotFound) {
				logger.Warn("role not found")
				return err
			}
			logger.Error("failed to find role by ID", applog.Error(err))
			return err
		}

		// 检查是否需要更新
		if (role.Name == "" || role.Name == existingRole.Name) &&
			(role.Description == "" || role.Description == existingRole.Description) {
			logger.Info("no need to update role")
			return shared.ErrNoRowsAffected
		}

		// 更新角色
		if err := roleRepo.Update(ctx, role); err != nil {
			logger.Error("failed to update role in repository", applog.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		// 如果角色没有变化，则返回角色信息
		if errors.Is(err, shared.ErrNoRowsAffected) {
			logger.Info("no need to update role")
			return response.ToRoleResponse(role), nil
		}
		logger.Error("failed to execute transaction", applog.Error(err))
		return nil, err
	}

	logger.Info("update role successfully")
	return response.ToRoleResponse(role), nil
}

// 删除角色
func (s *userService) DeleteRole(ctx context.Context, req *request.DeleteRoleRequest) error {
	logger := s.logger.With(applog.String("Method", "DeleteRole"), applog.Uint("role_id", req.ID))

	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		var err error
		roleRepo := provider.GetRoleRepository()
		// 先检查角色是否存在
		existingRole, err := roleRepo.FindByID(ctx, uint(req.ID))
		if err != nil {
			if errors.Is(err, user.ErrRoleNotFound) {
				logger.Warn("role not found")
				return err
			}
			logger.Error("failed to find role by ID", applog.Error(err))
			return err
		}

		// 删除角色
		if err := roleRepo.Delete(ctx, uint(existingRole.ID)); err != nil {
			logger.Error("failed to delete role in repository", applog.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to execute transaction", applog.Error(err))
		return err
	}

	logger.Info("delete role successfully")
	return nil
}
