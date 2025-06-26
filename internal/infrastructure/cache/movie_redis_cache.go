package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"mrs/internal/domain/movie"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/shared/vo"
	applog "mrs/pkg/log"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisMovieCache 电影缓存实现
type RedisMovieCache struct {
	redisClient *redis.Client
	logger      applog.Logger
}

// 创建一个RedisMovieCache实例
func NewRedisMovieCache(redisClient *redis.Client, logger applog.Logger) movie.MovieCache {
	return &RedisMovieCache{
		redisClient: redisClient,
		logger:      logger.With(applog.String("Component", "RedisMovieCache")),
	}
}

// SetMovie 设置单个电影的缓存
func (c *RedisMovieCache) SetMovie(ctx context.Context, mv *movie.Movie, expiration time.Duration) error {
	logger := c.logger.With(applog.String("Method", "SetMovie"), applog.Uint("movie_id", uint(mv.ID)))
	key := movie.GetMovieKey(mv.ID)
	data, err := json.Marshal(mv)

	if err != nil {
		logger.Error("failed to marshal movie", applog.Error(err))
		return fmt.Errorf("failed to marshal movie: %w", err)
	}

	if expiration == 0 {
		expiration = movie.DefaultMovieExpiration
	}

	if err := c.redisClient.Set(ctx, key, data, expiration).Err(); err != nil {
		logger.Error("failed to set movie to redis", applog.Error(err))
		return fmt.Errorf("failed to set movie to redis: %w", err)
	}

	logger.Info("set movie to redis successfully", applog.String("key", key))
	return nil
}

// DeleteMovie 从缓存中删除单个电影及其相关的列表缓存
func (c *RedisMovieCache) DeleteMovie(ctx context.Context, movieID vo.MovieID) error {
	logger := c.logger.With(
		applog.String("Method", "DeleteMovie"),
		applog.Uint("movie_id", uint(movieID)),
	)

	// 删除电影缓存
	key := movie.GetMovieKey(movieID)
	if err := c.redisClient.Del(ctx, key).Err(); err != nil {
		if err == redis.Nil {
			logger.Warn("movie not found in redis", applog.String("key", key))
			return fmt.Errorf("%w: %w", shared.ErrCacheMissing, err)
		}
		logger.Error("failed to delete movie from redis", applog.Error(err))
		return fmt.Errorf("failed to delete movie from redis: %w", err)
	}

	logger.Info("deleted movie and related caches successfully", applog.String("key", key))
	return nil
}

// GetMovie 获取单个电影缓存
func (c *RedisMovieCache) GetMovie(ctx context.Context, movieID vo.MovieID) (*movie.Movie, error) {
	logger := c.logger.With(applog.String("Method", "GetMovie"), applog.Uint("movie_id", uint(movieID)))
	key := movie.GetMovieKey(movieID)

	valBytes, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.Info("movie not found in redis", applog.String("key", key))
			return nil, fmt.Errorf("%w: %w", shared.ErrCacheMissing, err)
		}
		logger.Error("failed to get movie from redis", applog.Error(err))
		return nil, fmt.Errorf("failed to get movie from redis: %w", err)
	}

	var movie movie.Movie
	if err := json.Unmarshal(valBytes, &movie); err != nil {
		logger.Error("failed to unmarshal movie", applog.Error(err))
		return nil, fmt.Errorf("failed to unmarshal movie: %w", err)
	}
	logger.Info("get movie from redis successfully", applog.String("key", key))
	return &movie, nil
}

// SetMovies 批量设置多个电影缓存
func (c *RedisMovieCache) SetMovies(ctx context.Context, movies []*movie.Movie, expiration time.Duration) error {
	logger := c.logger.With(applog.String("Method", "SetMovies"), applog.Int("movies_count", len(movies)))

	pipe := c.redisClient.Pipeline()

	if expiration == 0 {
		expiration = movie.DefaultMovieExpiration
	}

	for _, mv := range movies {
		data, err := json.Marshal(mv)
		if err != nil {
			logger.Error("failed to marshal movie", applog.Error(err))
			continue
		}
		pipe.Set(ctx, movie.GetMovieKey(mv.ID), data, expiration)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Error("failed to set movies", applog.Error(err))
		return fmt.Errorf("failed to set movies: %w", err)
	}

	logger.Info("set movies successfully", applog.Int("movies_count", len(movies)))
	return nil
}

// SetMovieList 设置电影列表缓存（只存储ID）
func (c *RedisMovieCache) SetMovieList(ctx context.Context, movies []*movie.Movie, options *movie.MovieQueryOptions, expiration time.Duration) error {
	logger := c.logger.With(
		applog.String("Method", "SetMovieList"),
		applog.Any("options", options),
	)

	if expiration == 0 {
		expiration = movie.DefaultListExpiration
	}

	// 提取电影ID列表
	movieIDs := make([]uint, len(movies))
	for i, m := range movies {
		movieIDs[i] = uint(m.ID)
	}

	// 序列化ID列表
	data, err := json.Marshal(movieIDs)
	if err != nil {
		logger.Error("failed to marshal movie ID list", applog.Error(err))
		return fmt.Errorf("failed to marshal movie ID list: %w", err)
	}

	// 设置列表缓存（仅包含ID）
	listKey := movie.GetMovieListKey(options)
	if err := c.redisClient.Set(ctx, listKey, data, expiration).Err(); err != nil {
		logger.Error("failed to set movie ID list", applog.Error(err))
		return fmt.Errorf("failed to set movie ID list: %w", err)
	}

	// 同时缓存单个电影记录
	if err := c.SetMovies(ctx, movies, expiration); err != nil {
		logger.Error("failed to set individual movies", applog.Error(err))
		return fmt.Errorf("failed to set individual movies: %w", err)
	}

	logger.Info("set movie_id list and individual movies successfully",
		applog.String("list_key", listKey),
		applog.Int("movies_count", len(movies)))
	return nil
}

// GetMovieList 获取电影列表缓存，同时返回缓存状态信息
func (c *RedisMovieCache) GetMovieList(ctx context.Context, options *movie.MovieQueryOptions) (*movie.MovieCacheListResult, error) {
	logger := c.logger.With(
		applog.String("Method", "GetMovieList"),
		applog.Any("options", options),
	)

	result := &movie.MovieCacheListResult{
		Movies: make([]*movie.Movie, 0),
	}

	// 获取ID列表
	listKey := movie.GetMovieListKey(options)
	valBytes, err := c.redisClient.Get(ctx, listKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.Info("movie_id list not found in redis", applog.String("key", listKey))
			return nil, fmt.Errorf("%w: %w", shared.ErrCacheMissing, err)
		}
		logger.Error("failed to get movie_id list from redis", applog.Error(err))
		return nil, fmt.Errorf("failed to get movie_id list from redis: %w", err)
	}

	if err := json.Unmarshal(valBytes, &result.AllMovieIDs); err != nil {
		logger.Error("failed to unmarshal movie_id list", applog.Error(err))
		return nil, fmt.Errorf("failed to unmarshal movie_id list: %w", err)
	}

	if len(result.AllMovieIDs) == 0 {
		logger.Info("empty movie_id list in redis", applog.String("key", listKey))
		return result, nil
	}

	// 批量获取电影详情
	pipe := c.redisClient.Pipeline()
	movieCmds := make([]*redis.StringCmd, len(result.AllMovieIDs))
	for i, id := range result.AllMovieIDs {
		movieCmds[i] = pipe.Get(ctx, movie.GetMovieKey(id))
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		logger.Error("failed to get movies from redis", applog.Error(err))
		return nil, fmt.Errorf("failed to get movies from redis: %w", err)
	}

	// 处理结果并记录缺失的电影ID
	missingIDs := make(map[uint]struct{})
	for i, cmd := range movieCmds {
		movieID := result.AllMovieIDs[i]
		movieBytes, err := cmd.Bytes()
		if err != nil {
			logger.Warn("failed to get movie details",
				applog.Error(err),
				applog.Uint("movie_id", uint(movieID)))
			missingIDs[uint(movieID)] = struct{}{}
			continue
		}

		var movie movie.Movie
		if err := json.Unmarshal(movieBytes, &movie); err != nil {
			logger.Warn("failed to unmarshal movie",
				applog.Error(err),
				applog.Uint("movie_id", uint(movieID)))
			missingIDs[uint(movieID)] = struct{}{}
			continue
		}
		result.Movies = append(result.Movies, &movie)
	}

	// 将缺失的ID转换为切片并排序
	if len(missingIDs) > 0 {
		result.MissingMovieIDs = make([]vo.MovieID, 0, len(missingIDs))
		for id := range missingIDs {
			result.MissingMovieIDs = append(result.MissingMovieIDs, vo.MovieID(id))
		}
		sort.Slice(result.MissingMovieIDs, func(i, j int) bool {
			return result.MissingMovieIDs[i] < result.MissingMovieIDs[j]
		})
	}

	logger.Info("get movie list successfully",
		applog.String("list_key", listKey),
		applog.Int("total_ids", len(result.AllMovieIDs)),
		applog.Int("found_movies", len(result.Movies)),
		applog.Int("missing_movies", len(result.MissingMovieIDs)))

	return result, nil
}
