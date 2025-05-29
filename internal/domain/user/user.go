package user

import (
	"fmt"
	"mrs/internal/domain/role"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string     `gorm:"varchar(100),uniqueIndex,not null"`
	PasswordHash string     `gorm:"varchar(255),not null"`             // 存储密码的哈希值
	Email        string     `gorm:"varchar(255),uniqueIndex,not null"` // 用户邮箱，唯一索引
	FullName     string     `gorm:"varchar(100)"`                      // 用户全名（可选）
	LastLogin    *time.Time // 最后登录时间（可选）,使用指针可为null

	RoleID uint `gorm:"not null"` // 关联的角色ID
	Role   role.Role
}

// 接收明文密码并使用bcrypt哈希化存储
func (u *User) SetPassword(password string) error {
	PasswordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}
	u.PasswordHash = string(PasswordHash)
	return nil
}

// 验证明文密码与存储的哈希值是否匹配
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil // 如果 err 为 nil，表示密码匹配
}
