package response

import "mrs/internal/domain/user"

type UserProfileResponse struct {
	// ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleName string `json:"role_name"` // 来自关联的 Role 实体的 Name 字段
	// CreateAt time.Time `json:"create_at"`
	// UpdateAt time.Time `json:"update_at"`
	// IsActive bool      `json:"is_active"`
}

func ToUserProfileResponse(user *user.User) *UserProfileResponse {
	return &UserProfileResponse{
		// ID:       uint(user.ID),
		Username: user.Username,
		Email:    user.Email,
		RoleName: user.Role.Name,
	}
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleName string `json:"role_name"`
	RoleID   uint   `json:"role_id"`
}

func ToUserResponse(user *user.User) *UserResponse {
	return &UserResponse{
		ID:       uint(user.ID),
		Username: user.Username,
		Email:    user.Email,
		RoleName: user.Role.Name,
		RoleID:   uint(user.Role.ID),
	}
}

type ListUserResponse struct {
	PaginationResponse
	Users []*UserResponse `json:"users"`
}

func ToListUserResponse(users []*user.User) *ListUserResponse {
	userResponses := make([]*UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, ToUserResponse(user))
	}
	return &ListUserResponse{
		Users: userResponses,
	}
}

// RoleResponse 定义了角色响应的结构体。
type RoleResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ToRoleResponse(role *user.Role) *RoleResponse {
	return &RoleResponse{
		ID:          uint(role.ID),
		Name:        role.Name,
		Description: role.Description,
	}
}

// ListRoleResponse 定义了角色列表响应的结构体。
type ListRoleResponse struct {
	Roles []*RoleResponse `json:"roles"`
}

func ToListRoleResponse(roles []*user.Role) *ListRoleResponse {
	roleResponses := make([]*RoleResponse, 0, len(roles))
	for _, role := range roles {
		roleResponses = append(roleResponses, ToRoleResponse(role))
	}
	return &ListRoleResponse{Roles: roleResponses}
}
