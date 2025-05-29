package mysql

import (
	"context"
	"errors"
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
		logger.Error("failed to create user", applog.Error(err))
		return err
	}
	logger.Info("user created successfully", applog.Uint("user_id", usr.ID))
	return nil
}

// 通过ID获取用户，使用场景需要完整的用户信息，包括关联的角色信息。
func (r *gormUserRepository) FindByID(ctx context.Context, id uint) (*user.User, error) {
	logger := r.logger.With(applog.String("Method", "gormUserRepository.FindByID"))
	var usr user.User
	// Preload("Role") 会根据 User 结构体中的 Role 字段和外键 RoleID 加载关联的角色信息。
	if err := r.db.WithContext(ctx).Preload("Role").First(&usr, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not found by ID", applog.Uint("user_id", id))
			return nil, err
		}
		logger.Error("failed to find user by ID", applog.Uint("user_id", id), applog.Error(err))
		return nil, err
	}
	logger.Info("user found by ID", applog.Uint("ID", usr.ID))
	return &usr, nil
}

// 通过用户名获取用户，使用场景需要完整的用户信息，包括关联的角色信息。
func (r *gormUserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	logger := r.logger.With(applog.String("Method", "gormUserRepository.FindByUsername"))
	var usr user.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("username = ?", username).First(&usr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not found by username", applog.String("username", username))
			return nil, err
		}
		logger.Error("failed to find user by username", applog.String("username", username), applog.Error(err))
		return nil, err
	}
	logger.Info("user found by username", applog.String("username", usr.Username))
	return &usr, nil
}

// 通过邮箱获取用户，使用场景需要完整的用户信息，包括关联的角色信息。
func (r *gormUserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	logger := r.logger.With(applog.String("Method", "gormUserRepository.FindByEmail"))
	var usr user.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("email = ?", email).First(&usr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not found by email", applog.String("email", email))
			return nil, err
		}
		logger.Error("failed to find user by email", applog.String("email", email), applog.Error(err))
		return nil, err
	}
	logger.Info("user found by email", applog.String("email", usr.Email))
	return &usr, nil
}

// Update 更新用户信息
func (r *gormUserRepository) Update(ctx context.Context, usr *user.User) error {
	logger := r.logger.With(applog.String("Method", "gormUserRepository.Update"))
	if err := r.db.WithContext(ctx).Save(usr).Error; err != nil {
		logger.Error("failed to update user", applog.Uint("user_id", usr.ID), applog.Error(err))
		return err
	}
	logger.Info("user updated successfully", applog.Uint("user_id", usr.ID))
	return nil
}

// Delete 删除用户
func (r *gormUserRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "gormUserRepository.Delete"))
	if err := r.db.WithContext(ctx).Delete(&user.User{}, id).Error; err != nil {
		logger.Error("failed to delete user", applog.Uint("user_id", id), applog.Error(err))
	}
	logger.Info("user delete successfully", applog.Uint("user_id", id))
	return nil
}
