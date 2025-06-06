package movie

import "context"

// MovieRepository 定义了电影实体的持久化操作接口。
type MovieRepository interface {
	Create(ctx context.Context, movie *Movie) (*Movie, error)
	// 应支持预加载Genres
	FindByID(ctx context.Context, id uint) (*Movie, error)
	FindByIDs(ctx context.Context, ids []uint) ([]*Movie, error)
	FindByTitle(ctx context.Context, title string) (*Movie, error)
	// 分页和过滤，返回总数
	List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*Movie, int64, error)
	Update(ctx context.Context, movie *Movie) error
	Delete(ctx context.Context, id uint) error
	// 为电影增加、删除和修改类型
	AddGenreToMovie(ctx context.Context, movie *Movie, genre *Genre) error
	RemoveGenreToMovie(ctx context.Context, movie *Movie, genre *Genre) error
	ReplaceGenresForMovie(ctx context.Context, movie *Movie, genres []*Genre) error
}
