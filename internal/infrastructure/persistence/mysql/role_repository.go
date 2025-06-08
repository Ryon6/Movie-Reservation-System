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

func (r *gormRoleRepository) Create(ctx context.Context, rl *user.Role) (*user.Role, error) {
	logger := r.logger.With(applog.String("Method", "Create"), applog.Uint("role_id", uint(rl.ID)), applog.String("role_name", rl.Name))
	roleGorm := models.RoleGormFromDomain(rl)
	if err := r.db.WithContext(ctx).Create(roleGorm).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			logger.Warn("role already exists", applog.String("role_name", rl.Name))
			return nil, user.ErrRoleAlreadyExists
		}
		logger.Error("database create role error", applog.Error(err))
		return nil, fmt.Errorf("database create role error: %w", err)
	}
	logger.Info("create role successfully")
	return roleGorm.ToDomain(), nil
}

func (r *gormRoleRepository) FindByID(ctx context.Context, id uint) (*user.Role, error) {
	logger := r.logger.With(applog.String("Method", "FindByID"), applog.Uint("role_id", id))
	var roleGorm models.RoleGorm
	if err := r.db.WithContext(ctx).First(&roleGorm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("role not found by ID")
			// 封装哨兵错误
			return nil, fmt.Errorf("%w(id): %w", user.ErrRoleNotFound, err)
		}
		logger.Error("database find role by ID error", applog.Error(err))
		return nil, fmt.Errorf("database find role by ID error: %w", err)
	}
	logger.Info("find role by ID successfully")
	return roleGorm.ToDomain(), nil
}

func (r *gormRoleRepository) FindByName(ctx context.Context, name string) (*user.Role, error) {
	logger := r.logger.With(applog.String("Method", "FindByName"), applog.String("role_name", name))
	var roleGorm models.RoleGorm
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&roleGorm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("role not found by name")
			// 封装哨兵错误
			return nil, fmt.Errorf("%w(name): %w", user.ErrRoleNotFound, err)
		}
		logger.Error("database find role by name error", applog.Error(err))
		return nil, fmt.Errorf("database find role by name error: %w", err)
	}
	logger.Info("find role by name successfully")
	return roleGorm.ToDomain(), nil
}

func (r *gormRoleRepository) ListAll(ctx context.Context) ([]*user.Role, error) {
	logger := r.logger.With(applog.String("Method", "ListAll"))
	var roleGorms []*models.RoleGorm
	if err := r.db.WithContext(ctx).Find(&roleGorms).Error; err != nil {
		logger.Error("database list all roles error", applog.Error(err))
		return nil, fmt.Errorf("database list all roles error: %w", err)
	}
	logger.Info("list all roles successfully", applog.Int("count", len(roleGorms)))
	rls := make([]*user.Role, len(roleGorms))
	for i, roleGorm := range roleGorms {
		rls[i] = roleGorm.ToDomain()
	}
	return rls, nil
}

// 更新角色
func (r *gormRoleRepository) Update(ctx context.Context, role *user.Role) error {
	logger := r.logger.With(applog.String("Method", "Update"), applog.Uint("role_id", uint(role.ID)))
	roleGorm := models.RoleGormFromDomain(role)
	if err := r.db.WithContext(ctx).Model(&models.RoleGorm{}).Where("id = ?", roleGorm.ID).Updates(roleGorm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("role not found")
			return fmt.Errorf("%w(id): %w", user.ErrRoleNotFound, err)
		}
		logger.Error("database update role error", applog.Error(err))
		return fmt.Errorf("database update role error: %w", err)
	}
	logger.Info("update role successfully")
	return nil
}

// 删除角色
func (r *gormRoleRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "Delete"), applog.Uint("role_id", id))
	if err := r.db.WithContext(ctx).Delete(&models.RoleGorm{}, id).Error; err != nil {
		logger.Error("database delete role by ID error", applog.Error(err))
		return fmt.Errorf("database delete role by ID error: %w", err)
	}
	logger.Info("delete role successfully")
	return nil
}
