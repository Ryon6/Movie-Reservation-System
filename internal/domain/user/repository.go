package user

import (
	"context"
)

// UserRepository 定义了用户数据持久化操作的接口。
type UserRepository interface {
	Create(ctx context.Context, user *User) error                       // 创建用户
	FindByID(ctx context.Context, id uint) (*User, error)               // 通过ID获取用户
	FindByUsername(ctx context.Context, username string) (*User, error) // 通过name获取用户
	FindByEmail(ctx context.Context, email string) (*User, error)       // 通过email获取用户
	Update(ctx context.Context, user *User) error                       // 更新用户
	Delete(ctx context.Context, id uint) error                          // 删除用户
}
