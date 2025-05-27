package domain

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex"`
	PasswordHash string
	Email        string `gorm:"uniqueIndex"`
	RoleId       uint
	Role         Role
}
