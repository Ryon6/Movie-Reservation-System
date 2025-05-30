package response

import "time"

// LoginResponse 定义了成功登录后返回的结构体。
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}
