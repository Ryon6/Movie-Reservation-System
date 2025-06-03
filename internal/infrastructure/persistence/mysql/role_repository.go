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

type gormRoleRepository struct {
	db     *gorm.DB
	logger applog.Logger
}

func NewGormRoleRepository(db *gorm.DB, logger applog.Logger) user.RoleRepository {
	return &gormRoleRepository{
		db:     db,
		logger: logger.With(applog.String("Repository", "gormRoleRepository")),
	}
}

func (r *gormRoleRepository) Create(ctx context.Context, rl *user.Role) error {
	logger := r.logger.With(applog.String("Method", "Create"), applog.Uint("role_id", uint(rl.ID)), applog.String("role_name", rl.Name))
	roleGorm := models.RoleGormFromDomain(rl)
	if err := r.db.WithContext(ctx).Create(roleGorm).Error; err != nil {
		logger.Error("failed to create role", applog.Error(err))
		return err
	}
	logger.Info("role created successfully")
	return nil
}

func (r *gormRoleRepository) FindByID(ctx context.Context, id uint) (*user.Role, error) {
	logger := r.logger.With(applog.String("Method", "FindByID"), applog.Uint("role_id", id))
	var roleGorm models.RoleGorm
	if err := r.db.WithContext(ctx).First(&roleGorm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("role not found by ID")
			// 封装哨兵错误
			return nil, fmt.Errorf("%w: %w", user.ErrRoleNotFound, err)
		}
		logger.Error("failed to find role by ID", applog.Error(err))
		return nil, err
	}
	logger.Info("role created successfully", applog.String("role_name", roleGorm.Name))
	return roleGorm.ToDomain(), nil
}

func (r *gormRoleRepository) FindByName(ctx context.Context, name string) (*user.Role, error) {
	logger := r.logger.With(applog.String("Method", "FindByName"), applog.String("role_name", name))
	var roleGorm models.RoleGorm
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&roleGorm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("role not found by name")
			// 封装哨兵错误
			return nil, fmt.Errorf("%w: %w", user.ErrRoleNotFound, err)
		}
		logger.Error("failed to find role by name", applog.Error(err))
		return nil, err
	}
	logger.Info("role found by name", applog.Uint("role_id", uint(roleGorm.ID)))
	return roleGorm.ToDomain(), nil
}

func (r *gormRoleRepository) ListAll(ctx context.Context) ([]*user.Role, error) {
	logger := r.logger.With(applog.String("Method", "ListAll"))
	var roleGorms []*models.RoleGorm
	if err := r.db.WithContext(ctx).Find(&roleGorms).Error; err != nil {
		logger.Error("falied to list all roles", applog.Error(err))
		return nil, err
	}
	logger.Info("list all roles successfully", applog.Int("count", len(roleGorms)))
	rls := make([]*user.Role, len(roleGorms))
	for i, roleGorm := range roleGorms {
		rls[i] = roleGorm.ToDomain()
	}
	return rls, nil
}

func (r *gormRoleRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "Delete"), applog.Uint("role_id", id))
	if err := r.db.WithContext(ctx).Delete(&models.RoleGorm{}, id).Error; err != nil {
		logger.Error("failed to delete role by ID", applog.Error(err))
		return err
	}
	logger.Info("delete role by ID successfully")
	return nil
}
