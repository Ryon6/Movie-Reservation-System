package mysql

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/role"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type gormRoleRepository struct {
	db     *gorm.DB
	logger applog.Logger
}

func NewGormRoleRepository(db *gorm.DB, logger applog.Logger) role.RoleRepository {
	return &gormRoleRepository{
		db:     db,
		logger: logger,
	}
}

func (r *gormRoleRepository) Create(ctx context.Context, rl *role.Role) error {
	logger := r.logger.With(applog.String("Method", "gormRoleRepository.Create"), applog.Uint("role_id", rl.ID))
	if err := r.db.WithContext(ctx).Create(rl).Error; err != nil {
		logger.Error("failed to create role", applog.Error(err))
		return err
	}
	logger.Info("role created successfully")
	return nil
}

func (r *gormRoleRepository) FindByID(ctx context.Context, id uint) (*role.Role, error) {
	logger := r.logger.With(applog.String("Method", "gormRoleRepository.FindByID"), applog.Uint("role_id", id))
	var rl role.Role
	if err := r.db.WithContext(ctx).First(&rl, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("role not found by ID")
			// 封装哨兵错误
			return nil, fmt.Errorf("%w: %w", role.ErrRoleNotFound, err)
		}
		logger.Error("failed to find role by ID", applog.Error(err))
		return nil, err
	}
	logger.Info("role created successfully", applog.String("role_name", rl.Name))
	return &rl, nil
}

func (r *gormRoleRepository) FindByName(ctx context.Context, name string) (*role.Role, error) {
	logger := r.logger.With(applog.String("Method", "gormRoleRepository.FindByName"), applog.String("role_name", name))
	var rl role.Role
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&rl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("role not found by name")
			// 封装哨兵错误
			return nil, fmt.Errorf("%w: %w", role.ErrRoleNotFound, err)
		}
		logger.Error("failed to find role by name", applog.Error(err))
		return nil, err
	}
	logger.Info("role found by name", applog.Uint("role_id", rl.ID))
	return &rl, nil
}

func (r *gormRoleRepository) ListAll(ctx context.Context) ([]*role.Role, error) {
	logger := r.logger.With(applog.String("Method", "gormRoleRepository.ListAll"))
	var rls []*role.Role
	if err := r.db.WithContext(ctx).Find(&rls).Error; err != nil {
		logger.Error("falied to list all roles", applog.Error(err))
		return nil, err
	}
	logger.Info("list all roles successfully")
	return rls, nil
}

func (r *gormRoleRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "gormRoleRepository.ListAll"), applog.Uint("role_id", id))
	if err := r.db.WithContext(ctx).Delete(&role.Role{}, id).Error; err != nil {
		logger.Error("failed to delete role by ID", applog.Error(err))
		return err
	}
	logger.Info("delete role by ID successfully")
	return nil
}
