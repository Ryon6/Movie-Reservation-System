package user

import "mrs/internal/domain/shared/vo"

// 角色
type Role struct {
	Name        string
	Description string
	ID          vo.RoleID
}

// 角色名称
const (
	AdminRoleName = "ADMIN"
	UserRoleName  = "USER"
)
