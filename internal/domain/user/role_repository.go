package user

import "context"

type RoleRepository interface {
	Create(ctx context.Context, role *Role) (*Role, error)      // 创建角色
	FindByID(ctx context.Context, id uint) (*Role, error)       // 根据ID查找角色
	FindByName(ctx context.Context, name string) (*Role, error) // 根据名称查找角色
	ListAll(ctx context.Context) ([]*Role, error)               // 列出所有角色
	Delete(ctx context.Context, id uint) error                  // 通常角色是预定义且不轻易删除
}
