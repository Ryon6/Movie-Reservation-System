package models

import (
	"mrs/internal/domain/shared/vo"
	"mrs/internal/domain/user"

	"gorm.io/gorm"
)

// 角色表
type RoleGorm struct {
	gorm.Model
	Name        string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Description string `gorm:"type:varchar(255)"` // 角色描述（可选）
}

// TableName 指定表名
func (RoleGorm) TableName() string {
	return "roles"
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
