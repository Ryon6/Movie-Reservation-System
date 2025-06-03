package models

import "gorm.io/gorm"

type RoleGorm struct {
	gorm.Model
	Name        string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Description string `gorm:"type:varchar(255)"` // 角色描述（可选）
}
