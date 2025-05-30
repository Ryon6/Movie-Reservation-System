package dto

import (
	"mrs/internal/domain/user"
	"time"
)

// AuthResult 定义认证服务返回的统一数据结构
type AuthResult struct {
	Token     string    `json:"token"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	RoleName  string    `json:"role_name"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	CreateAt  time.Time `json:"create_at"`
	UpdateAt  time.Time `json:"update_at"`
}

type UserProfile struct {
	ID       uint      `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	RoleName string    `json:"role_name"`
	CreateAt time.Time `json:"create_at"`
	UpdateAt time.Time `json:"update_at"`
}

func ToUserProfile(usr *user.User) *UserProfile {
	return &UserProfile{
		ID:       usr.ID,
		Username: usr.Username,
		Email:    usr.Email,
		RoleName: usr.Role.Name,
		CreateAt: usr.CreatedAt,
		UpdateAt: usr.UpdatedAt,
	}
}
