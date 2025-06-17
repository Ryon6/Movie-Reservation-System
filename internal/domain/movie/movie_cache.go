package movie

import (
	"context"
	"fmt"
	"mrs/internal/domain/shared/vo"
	"strings"
	"time"
)

const (
	DefaultExpiration = 10 * time.Minute
)

// MovieCache 电影缓存接口
type MovieCache interface {
	GetMovie(ctx context.Context, movieID vo.MovieID) (*Movie, error)
	SetMovie(ctx context.Context, movie *Movie, expiration time.Duration) error
	DeleteMovie(ctx context.Context, movieID vo.MovieID) error
	GetMovieList(ctx context.Context, options *MovieQueryOptions) (*MovieCacheListResult, error)
	SetMovieList(ctx context.Context, movies []*Movie, options *MovieQueryOptions, expiration time.Duration) error
}

// MovieCacheListResult 电影列表的查询结果
type MovieCacheListResult struct {
	Movies          []*Movie     // 成功获取的电影记录
	AllMovieIDs     []vo.MovieID // 列表中所有的电影ID
	MissingMovieIDs []vo.MovieID // 缓存中未找到的电影ID
}

// 电影缓存键前缀
const (
	movieKeyPrefix     = "movie:"
	movieListKeyPrefix = "movies:list:"
)

// movieKey 生成单个电影的缓存键
func GetMovieKey(movieID vo.MovieID) string {
	return fmt.Sprintf("%s%d", movieKeyPrefix, movieID)
}

// movieListKey 生成电影列表的缓存键
func GetMovieListKey(options *MovieQueryOptions) string {
	if options == nil {
		return movieListKeyPrefix + "all"
	}

	// 规范化参数以确保键的一致性
	var sb strings.Builder
	sb.WriteString(movieListKeyPrefix)
	sb.WriteString(fmt.Sprintf("%s=%v:", "title", options.Title))              // 构建器追加字符串
	sb.WriteString(fmt.Sprintf("%s=%v:", "release_year", options.ReleaseYear)) // 构建器追加字符串
	sb.WriteString(fmt.Sprintf("%s=%v:", "genre_name", options.GenreName))     // 构建器追加字符串
	sb.WriteString(fmt.Sprintf("%s=%v:", "page", options.Page))                // 构建器追加字符串
	sb.WriteString(fmt.Sprintf("%s=%v:", "page_size", options.PageSize))       // 构建器追加字符串

	return strings.TrimRight(sb.String(), ":") // 移除字符串右侧的:符号
}
