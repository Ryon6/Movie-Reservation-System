package user

import (
	"context"
	"mrs/internal/domain/shared/vo"
)

// UserRepository 定义了用户数据持久化操作的接口。
type UserRepository interface {
	Create(ctx context.Context, user *User) error                                // 创建用户
	FindByID(ctx context.Context, id vo.UserID) (*User, error)                   // 通过ID获取用户
	FindByUsername(ctx context.Context, username string) (*User, error)          // 通过name获取用户
	FindByEmail(ctx context.Context, email string) (*User, error)                // 通过email获取用户
	CheckRoleReferenced(ctx context.Context, roleID vo.RoleID) (bool, error)     // 检查是否存在任何“活跃的”用户关联到这个角色
	Update(ctx context.Context, user *User) error                                // 更新用户
	Delete(ctx context.Context, id vo.UserID) error                              // 删除用户
	List(ctx context.Context, options *UserQueryOptions) ([]*User, int64, error) // 获取所有用户
}

type UserQueryOptions struct {
	Page     int
	PageSize int
	// Username string
	// Email    string
	// RoleName string
	// RoleID   uint
}
