package response

import (
	"mrs/internal/app"
	"time"
)

type UserReponse struct {
	ID       uint      `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	RoleName string    `json:"role_name"` // 来自关联的 Role 实体的 Name 字段
	CreateAt time.Time `json:"create_at"`
	UpdateAt time.Time `json:"update_at"`
	// IsActive bool      `json:"is_active"`
}

// 从领域用户实体转换为 UserResponse DTO
func FromDomainUser(rst *app.LoginResult) UserReponse {
	return UserReponse{
		ID:       rst.UserID,
		Username: rst.Username,
		Email:    rst.Email,
		RoleName: rst.RoleName,
		CreateAt: rst.CreateAt,
		UpdateAt: rst.UpdateAt,
	}
}
