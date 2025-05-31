package mysql

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/user"
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
		logger: logger,
	}
}

// 创建用户
func (r *gormUserRepository) Create(ctx context.Context, usr *user.User) error {
	logger := r.logger.With(applog.String("Method", "gormUserRepository.Create"))
	if err := r.db.WithContext(ctx).Create(usr).Error; err != nil {
		// 封装哨兵错误
		if errors.Is(err, gorm.ErrDuplicatedKey) || errors.Is(err, gorm.ErrRegistered) {
			logger.Warn("user already exists", applog.Error(err))
			return fmt.Errorf("%w: %w", user.ErrUserExists, err)
		}
		logger.Error("failed to create user", applog.Error(err))
		return err
	}
	logger.Info("user created successfully", applog.Uint("user_id", usr.ID))
	return nil
}

// 通过ID获取用户，使用场景需要完整的用户信息，包括关联的角色信息。
func (r *gormUserRepository) FindByID(ctx context.Context, id uint) (*user.User, error) {
	logger := r.logger.With(applog.String("Method", "gormUserRepository.FindByID"), applog.Uint("user_id", id))
	var usr user.User
	// Preload("Role") 会根据 User 结构体中的 Role 字段和外键 RoleID 加载关联的角色信息。
	if err := r.db.WithContext(ctx).Preload("Role").First(&usr, id).Error; err != nil {
		// 封装哨兵错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not found by ID")
			return nil, fmt.Errorf("%w: %w", user.ErrUserNotFound, err)
		}
		logger.Error("failed to find user by ID", applog.Error(err))
		return nil, err
	}
	logger.Info("user found by ID", applog.String("username", usr.Username))
	return &usr, nil
}

// 通过用户名获取用户，使用场景需要完整的用户信息，包括关联的角色信息。
func (r *gormUserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	logger := r.logger.With(applog.String("Method", "gormUserRepository.FindByUsername"), applog.String("username", username))
	var usr user.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("username = ?", username).First(&usr).Error; err != nil {
		// 封装哨兵错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not found by username")
			return nil, fmt.Errorf("%w: %w", user.ErrUserNotFound, err)
		}
		logger.Error("failed to find user by username", applog.Error(err))
		return nil, err
	}
	logger.Info("user found by username", applog.Uint("user_id", usr.ID))
	return &usr, nil
}

// 通过邮箱获取用户，使用场景需要完整的用户信息，包括关联的角色信息。
func (r *gormUserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	logger := r.logger.With(applog.String("Method", "gormUserRepository.FindByEmail"), applog.String("email", email))
	var usr user.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("email = ?", email).First(&usr).Error; err != nil {
		// 封装哨兵错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not found by email")
			return nil, fmt.Errorf("%w: %w", user.ErrUserNotFound, err)
		}
		logger.Error("failed to find user by email", applog.Error(err))
		return nil, err
	}
	logger.Info("user found by email", applog.Uint("user_id", usr.ID))
	return &usr, nil
}

// Update 更新用户信息
func (r *gormUserRepository) Update(ctx context.Context, usr *user.User) error {
	logger := r.logger.With(applog.String("Method", "gormUserRepository.Update"), applog.Uint("user_id", usr.ID))
	if err := r.db.WithContext(ctx).Save(usr).Error; err != nil {
		logger.Error("failed to update user", applog.Error(err))
		return err
	}
	logger.Info("user updated successfully")
	return nil
}

// Delete 删除用户
func (r *gormUserRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "gormUserRepository.Delete"), applog.Uint("user_id", id))
	if err := r.db.WithContext(ctx).Delete(&user.User{}, id).Error; err != nil {
		logger.Error("failed to delete user", applog.Error(err))
		return err
	}
	logger.Info("user delete successfully")
	return nil
}
