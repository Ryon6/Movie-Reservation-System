package movie

import "context"

// GenreRepository 定义了类型实体的持久化操作接口。
type GenreRepository interface {
	Create(ctx context.Context, genre *Genre) (*Genre, error)
	FindByID(ctx context.Context, id uint) (*Genre, error)
	FindByName(ctx context.Context, name string) (*Genre, error)
	ListAll(ctx context.Context) ([]*Genre, error)
	Update(ctx context.Context, genre *Genre) error
	// 对于Genre的删除操作，若存在外键约束错误，应该返回错误，由服务层删除Movie中对应记录后再进行删除
	Delete(ctx context.Context, id uint) error
	// 常用：查找或创建
	FindOrCreateByNames(ctx context.Context, names []string) ([]*Genre, error)
}
