package user

import (
	"fmt"
	"mrs/internal/domain/shared/vo"
	"mrs/internal/utils"
)

// 用户
type User struct {
	ID           vo.UserID
	Username     string
	PasswordHash string
	Email        string
	Role         *Role // 聚合内部可以直接持有同一聚合内其他实体的引用
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

// 验证明文密码与存储的哈希值是否匹配
func (u *User) CheckPassword(password string, hasher utils.PasswordHasher) bool {
	ok, err := hasher.Check(u.PasswordHash, password)
	if err != nil {
		return false
	}
	return ok // 如果 err 为 nil，表示密码匹配
}
