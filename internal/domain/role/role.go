package role

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	Name        string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Description string `gorm:"type:varvahr(255)"` // 角色描述（可选）
}

const (
	AdminRoleName = "ADMIN"
	UserRoleName  = "USER"
)
