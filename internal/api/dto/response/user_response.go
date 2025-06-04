package response

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleName string `json:"role_name"` // 来自关联的 Role 实体的 Name 字段
	// CreateAt time.Time `json:"create_at"`
	// UpdateAt time.Time `json:"update_at"`
	// IsActive bool      `json:"is_active"`
}
