// TODO: 修改user的验证部分，采用该接口验证

package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher 定义了密码哈希和验证的接口。
type PasswordHasher interface {
	// Hash 对给定的明文密码进行哈希处理。
	Hash(password string) (string, error)
	// Check 将明文密码与已哈希的密码进行比较。
	// 如果匹配则返回 true，否则返回 false。
	// 如果在比较过程中发生除不匹配之外的错误（例如哈希格式无效），则返回该错误。
	Check(hashPassword, password string) (bool, error)
}

// bcryptHasher 是 PasswordHasher 接口的使用 bcrypt 算法的实现。
type bcryptHasher struct {
	cost int // bcrypt 的计算成本
}

func NewBcryptHasher(cost int) *bcryptHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &bcryptHasher{
		cost: cost,
	}
}

func (b *bcryptHasher) Hash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", fmt.Errorf("bcryptHasher.Hash: %w", err)
	}
	return string(hashedBytes), nil
}

func (b *bcryptHasher) Check(hashPassword, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	if err == nil {
		return true, nil
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	return false, fmt.Errorf("bcryptHasher.Check: %w", err)
}
