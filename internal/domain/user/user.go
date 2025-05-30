package user

import (
	"fmt"
	"mrs/internal/domain/role"
	"mrs/internal/utils"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string     `gorm:"varchar(100),uniqueIndex,not null"`
	PasswordHash string     `gorm:"varchar(255),not null"`             // 存储密码的哈希值
	Email        string     `gorm:"varchar(255),uniqueIndex,not null"` // 用户邮箱，唯一索引
	FullName     string     `gorm:"varchar(100)"`                      // 用户全名（可选）
	LastLogin    *time.Time // 最后登录时间（可选）,使用指针可为null

	RoleID uint      `gorm:"not null"`           // 关联的角色ID
	Role   role.Role `gorm:"foreignKey:RoleID "` // 通常会隐式推断，这里显式定义防止出错
}

// 接收明文密码并使用bcrypt哈希化存储
func (u *User) SetPassword(password string, hasher utils.PasswordHasher) error {
	PasswordHash, err := hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}
	u.PasswordHash = string(PasswordHash)
	return nil
}

// // 验证明文密码与存储的哈希值是否匹配
// func (u *User) CheckPassword(password string, hasher utils.PasswordHasher) bool {
// 	// err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
// 	ok, err := hasher.Check(u.PasswordHash, password)
// 	if err != nil {
// 		return false
// 	}
// 	return ok // 如果 err 为 nil，表示密码匹配
// }
