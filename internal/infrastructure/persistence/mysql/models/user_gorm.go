package models

import (
	"mrs/internal/domain/shared/vo"
	"mrs/internal/domain/user"
	"time"

	"gorm.io/gorm"
)

// 用户表
type UserGorm struct {
	gorm.Model
	Username     string     `gorm:"varchar(100),uniqueIndex,not null"`
	PasswordHash string     `gorm:"varchar(255),not null"`             // 存储密码的哈希值
	Email        string     `gorm:"varchar(255),uniqueIndex,not null"` // 用户邮箱，唯一索引
	FullName     string     `gorm:"varchar(100)"`                      // 用户全名（可选）
	LastLogin    *time.Time // 最后登录时间（可选）,使用指针可为null

	RoleID uint     `gorm:"not null"`           // 关联的角色ID
	Role   RoleGorm `gorm:"foreignKey:RoleID "` // 通常会隐式推断，这里显式定义防止出错
}

// TableName 指定表名
func (UserGorm) TableName() string {
	return "users"
}

func (u *UserGorm) ToDomain() *user.User {
	return &user.User{
		ID:           vo.UserID(u.ID),
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role.ToDomain(),
	}
}

func UserGormFromDomain(u *user.User) *UserGorm {
	return &UserGorm{
		Model:        gorm.Model{ID: uint(u.ID)},
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         *RoleGormFromDomain(u.Role),
	}
}
