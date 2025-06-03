package models

import (
	"mrs/internal/domain/shared/vo"
	"mrs/internal/domain/user"

	"gorm.io/gorm"
)

type RoleGorm struct {
	gorm.Model
	Name        string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Description string `gorm:"type:varchar(255)"` // 角色描述（可选）
}

func (r *RoleGorm) ToDomain() *user.Role {
	return &user.Role{
		ID:          vo.RoleID(r.ID),
		Name:        r.Name,
		Description: r.Description,
	}
}

func RoleGormFromDomain(r *user.Role) *RoleGorm {
	return &RoleGorm{
		Model:       gorm.Model{ID: uint(r.ID)},
		Name:        r.Name,
		Description: r.Description,
	}
}
