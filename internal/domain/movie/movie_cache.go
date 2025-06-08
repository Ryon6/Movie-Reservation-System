package movie

import (
	"context"
	"time"
)

// MovieCache 电影缓存接口
type MovieCache interface {
	GetMovie(ctx context.Context, movieID uint) (*Movie, error)
	SetMovie(ctx context.Context, movie *Movie, expiration time.Duration) error
	DeleteMovie(ctx context.Context, movieID uint) error
	GetMovieList(ctx context.Context, options *MovieQueryOptions) (*MovieCacheListResult, error)
	SetMovieList(ctx context.Context, movies []*Movie, options *MovieQueryOptions, expiration time.Duration) error
}

// MovieCacheListResult 电影列表的查询结果
type MovieCacheListResult struct {
	Movies          []*Movie // 成功获取的电影记录
	AllMovieIDs     []uint   // 列表中所有的电影ID
	MissingMovieIDs []uint   // 缓存中未找到的电影ID
}
