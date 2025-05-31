package movie

import "context"

// GenreRepository 定义了类型实体的持久化操作接口。
type GenreRepository interface {
	Create(ctx context.Context, genre *Genre) error
	FindByID(ctx context.Context, id uint) (*Genre, error)
	FindByName(ctx context.Context, name string) (*Genre, error)
	ListAll(ctx context.Context) (*[]Genre, error)
	Update(ctx context.Context, genre *Genre) error
	Delete(ctx context.Context, id uint) (*Genre, error)
	// 常用：查找或创建
	FindOrCreateByName(ctx context.Context, name string) (*Genre, error)
}
