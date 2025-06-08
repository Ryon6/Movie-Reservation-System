package mysql

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/user"
	"mrs/internal/infrastructure/persistence/mysql/models"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type gormUserRepository struct {
	db     *gorm.DB
	logger applog.Logger
}

func NewGormUserRepository(db *gorm.DB, logger applog.Logger) user.UserRepository {
	return &gormUserRepository{
		db:     db,
		logger: logger.With(applog.String("Repository", "gormUserRepository")),
	}
}

// 创建用户
func (r *gormUserRepository) Create(ctx context.Context, usr *user.User) error {
	logger := r.logger.With(applog.String("Method", "Create"), applog.Uint("user_id", uint(usr.ID)), applog.String("username", usr.Username))
	userGorm := models.UserGormFromDomain(usr)
	if err := r.db.WithContext(ctx).Create(userGorm).Error; err != nil {
		// 封装哨兵错误
		if errors.Is(err, gorm.ErrDuplicatedKey) || errors.Is(err, gorm.ErrRegistered) {
			logger.Warn("user already exists", applog.Error(err))
			return fmt.Errorf("%w(id): %w", user.ErrUserAlreadyExists, err)
		}
		logger.Error("database create user error", applog.Error(err))
		return fmt.Errorf("database create user error: %w", err)
	}
	logger.Info("create user successfully")
	return nil
}

// 通过ID获取用户，使用场景需要完整的用户信息，包括关联的角色信息。
func (r *gormUserRepository) FindByID(ctx context.Context, id uint) (*user.User, error) {
	logger := r.logger.With(applog.String("Method", "FindByID"), applog.Uint("user_id", id))
	var userGorm models.UserGorm
	// Preload("Role") 会根据 User 结构体中的 Role 字段和外键 RoleID 加载关联的角色信息。
	if err := r.db.WithContext(ctx).Preload("Role").First(&userGorm, id).Error; err != nil {
		// 封装哨兵错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not found by ID")
			return nil, fmt.Errorf("%w(id): %w", user.ErrUserNotFound, err)
		}
		logger.Error("database find user by ID error", applog.Error(err))
		return nil, fmt.Errorf("database find user by ID error: %w", err)
	}
	logger.Info("find user by ID successfully")
	return userGorm.ToDomain(), nil
}

// 通过用户名获取用户，使用场景需要完整的用户信息，包括关联的角色信息。
func (r *gormUserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	logger := r.logger.With(applog.String("Method", "FindByUsername"), applog.String("username", username))
	var userGorm models.UserGorm
	if err := r.db.WithContext(ctx).Preload("Role").Where("username = ?", username).First(&userGorm).Error; err != nil {
		// 封装哨兵错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not found by username")
			return nil, fmt.Errorf("%w(username): %w", user.ErrUserNotFound, err)
		}
		logger.Error("database find user by username error", applog.Error(err))
		return nil, fmt.Errorf("database find user by username error: %w", err)
	}
	logger.Info("find user by username successfully")
	return userGorm.ToDomain(), nil
}

// 通过邮箱获取用户，使用场景需要完整的用户信息，包括关联的角色信息。
func (r *gormUserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	logger := r.logger.With(applog.String("Method", "FindByEmail"), applog.String("email", email))
	var userGorm models.UserGorm
	if err := r.db.WithContext(ctx).Preload("Role").Where("email = ?", email).First(&userGorm).Error; err != nil {
		// 封装哨兵错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not found by email")
			return nil, fmt.Errorf("%w(email): %w", user.ErrUserNotFound, err)
		}
		logger.Error("database find user by email error", applog.Error(err))
		return nil, fmt.Errorf("database find user by email error: %w", err)
	}
	logger.Info("find user by email successfully")
	return userGorm.ToDomain(), nil
}

// Update 更新用户信息
func (r *gormUserRepository) Update(ctx context.Context, usr *user.User) error {
	logger := r.logger.With(applog.String("Method", "Update"), applog.Uint("user_id", uint(usr.ID)))
	userGorm := models.UserGormFromDomain(usr)
	// 使用Updates方法更新用户信息，避免使用Save方法，因为Save方法会保存所有字段，包括零值
	if err := r.db.WithContext(ctx).Model(&models.UserGorm{}).Where("id = ?", userGorm.ID).Updates(userGorm).Error; err != nil {
		logger.Error("database update user error", applog.Error(err))
		return fmt.Errorf("database update user error: %w", err)
	}
	logger.Info("update user successfully")
	return nil
}

// Delete 删除用户
func (r *gormUserRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "Delete"), applog.Uint("user_id", id))
	if err := r.db.WithContext(ctx).Delete(&models.UserGorm{}, id).Error; err != nil {
		logger.Error("database delete user error", applog.Error(err))
		return fmt.Errorf("database delete user error: %w", err)
	}
	logger.Info("delete user successfully")
	return nil
}
