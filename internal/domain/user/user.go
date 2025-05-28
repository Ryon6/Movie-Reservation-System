package user

import (
	"mrs/internal/domain/role"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex"`
	PasswordHash string
	Email        string `gorm:"uniqueIndex"`
	RoleId       uint
	Role         role.Role
}
