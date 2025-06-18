package request

import (
	"mrs/internal/domain/shared/vo"
	"mrs/internal/domain/user"
)

// RegisterUserRequest 定义了用户注册请求的结构体。
type RegisterUserRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=100"`
	Email    string `json:"email" binding:"required,email"`
}

type GetUserRequest struct {
	ID uint
}

type UpdateUserRequest struct {
	ID       uint
	Username string `json:"username" binding:"omitempty,alphanum,min=3,max=50"`
	Password string `json:"password" binding:"omitempty,min=8,max=100"`
	Email    string `json:"email" binding:"omitempty,email"`
}

func (r *UpdateUserRequest) ToDomain() *user.User {
	return &user.User{
		ID:       vo.UserID(r.ID),
		Username: r.Username,
		Email:    r.Email,
	}
}

type DeleteUserRequest struct {
	ID uint
}

// ListUserRequest 定义了用户列表请求的结构体。
type ListUserRequest struct {
	PaginationRequest
}

func (r *ListUserRequest) ToDomain() *user.UserQueryOptions {
	return &user.UserQueryOptions{
		Page:     r.Page,
		PageSize: r.PageSize,
	}
}

// CreateRoleRequest 定义了创建角色请求的结构体。
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

func (r *CreateRoleRequest) ToDomain() *user.Role {
	return &user.Role{
		Name:        r.Name,
		Description: r.Description,
	}
}

// UpdateRoleRequest 定义了更新角色请求的结构体。
type UpdateRoleRequest struct {
	ID          uint   `json:"id" binding:"required,min=1"`
	Name        string `json:"name" binding:"omitempty,min=3,max=50"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

func (r *UpdateRoleRequest) ToDomain() *user.Role {
	return &user.Role{
		ID:          vo.RoleID(r.ID),
		Name:        r.Name,
		Description: r.Description,
	}
}

// DeleteRoleRequest 定义了删除角色请求的结构体。
type DeleteRoleRequest struct {
	ID uint
}

// 为用户分配角色
type AssignRoleToUserRequest struct {
	UserID uint `json:"user_id" binding:"required"`
	RoleID uint `json:"role_id" binding:"required"`
}
