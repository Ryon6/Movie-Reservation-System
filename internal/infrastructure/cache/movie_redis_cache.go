package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"mrs/internal/domain/movie"
	applog "mrs/pkg/log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisMovieCache struct {
	redisClient       RedisClient
	logger            applog.Logger
	defaultExpiration time.Duration
}

const (
	movieKeyPrefix      = "movie:"
	movieListKeyPrefix  = "movies:list:"
	movieSetKeyPrefix   = "movies:set:"
	movieIndexKeyPrefix = "movies:index:"
)

// 创建一个RedisMovieCache实例
func NewRedisMovieCache(redisClient RedisClient, logger applog.Logger, defaultExpiration time.Duration) *RedisMovieCache {
	return &RedisMovieCache{
		redisClient:       redisClient,
		logger:            logger.With(applog.String("Component", "RedisMovieCache")),
		defaultExpiration: defaultExpiration,
	}
}

// movieKey 生成单个电影的缓存键
func (c *RedisMovieCache) movieKey(movieID uint) string {
	return fmt.Sprintf("%s%d", movieKeyPrefix, movieID)
}

// movieListKey 生成电影列表的缓存键
func (c *RedisMovieCache) movieListKey(params map[string]interface{}) string {
	if len(params) == 0 {
		return movieListKeyPrefix + "all"
	}

	// 规范化参数以确保键的一致性
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(movieListKeyPrefix)
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%v:", k, params[k])) // 构建器追加字符串
	}
	return strings.TrimRight(sb.String(), ":") // 移除字符串右侧的:符号
}

// movieSetKey 生成存储电影ID集合的键
func (c *RedisMovieCache) movieSetKey(params map[string]interface{}) string {
	return movieSetKeyPrefix + strings.TrimPrefix(c.movieListKey(params), movieListKeyPrefix)
}

// movieIndexKey 生成电影的反向索引键
func (c *RedisMovieCache) movieIndexKey(movieID uint) string {
	return fmt.Sprintf("%s%d", movieIndexKeyPrefix, movieID)
}

// invalidateMovieLists 使包含指定电影的所有列表缓存失效
func (c *RedisMovieCache) invalidateMovieLists(ctx context.Context, movieID uint) error {
	logger := c.logger.With(
		applog.String("Method", "invalidateMovieLists"),
		applog.Uint("movie_id", movieID),
	)

	// 获取包含该电影的所有列表集合键
	indexKey := c.movieIndexKey(movieID)
	setKeys, err := c.redisClient.SMembers(ctx, indexKey).Result()
	if err != nil && err != redis.Nil {
		logger.Error("failed to get movie lists", applog.Error(err))
		return fmt.Errorf("failed to get movie lists: %w", err)
	}

	if len(setKeys) == 0 {
		return nil
	}

	// 构造要删除的列表缓存键
	listKeys := make([]string, len(setKeys))
	for i, setKey := range setKeys {
		listKeys[i] = strings.Replace(setKey, movieSetKeyPrefix, movieListKeyPrefix, 1)
	}

	// 使用管道批量删除缓存
	pipe := c.redisClient.Pipeline()

	// 删除列表缓存
	pipe.Del(ctx, listKeys...)
	// 删除集合
	pipe.Del(ctx, setKeys...)
	// 删除反向索引
	pipe.Del(ctx, indexKey)

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Error("failed to invalidate movie lists", applog.Error(err))
		return fmt.Errorf("failed to invalidate movie lists: %w", err)
	}

	logger.Info("invalidated movie lists successfully",
		applog.Int("list_count", len(listKeys)))
	return nil
}

// SetMovie 设置单个电影的缓存
func (c *RedisMovieCache) SetMovie(ctx context.Context, movie *movie.Movie, expiration time.Duration) error {
	logger := c.logger.With(applog.String("Method", "SetMovie"), applog.Uint("movie_id", movie.ID))
	key := c.movieKey(movie.ID)
	data, err := json.Marshal(movie)

	if err != nil {
		logger.Error("failed to marshal movie", applog.Error(err))
		return fmt.Errorf("failed to marshal movie: %w", err)
	}

	if expiration == 0 {
		expiration = c.defaultExpiration
	}

	if err := c.redisClient.Set(ctx, key, data, expiration).Err(); err != nil {
		logger.Error("failed to set movie to redis", applog.Error(err))
		return fmt.Errorf("failed to set movie to redis: %w", err)
	}

	logger.Info("set movie to redis successfully", applog.String("key", key))
	return nil
}

// DeleteMovie 从缓存中删除单个电影及其相关的列表缓存
func (c *RedisMovieCache) DeleteMovie(ctx context.Context, movieID uint) error {
	logger := c.logger.With(
		applog.String("Method", "DeleteMovie"),
		applog.Uint("movie_id", movieID),
	)

	// 首先使相关的列表缓存失效
	if err := c.invalidateMovieLists(ctx, movieID); err != nil {
		logger.Error("failed to invalidate related lists", applog.Error(err))
		// 继续执行，不要因为列表缓存失效失败而中断
	}

	// 删除电影缓存
	key := c.movieKey(movieID)
	if err := c.redisClient.Del(ctx, key).Err(); err != nil {
		if err == redis.Nil {
			logger.Warn("movie not found in redis", applog.String("key", key))
			return fmt.Errorf("%w: %w", ErrKeyNotFound, err)
		}
		logger.Error("failed to delete movie from redis", applog.Error(err))
		return fmt.Errorf("failed to delete movie from redis: %w", err)
	}

	logger.Info("deleted movie and related caches successfully", applog.String("key", key))
	return nil
}

func (c *RedisMovieCache) GetMovie(ctx context.Context, movieID uint) (*movie.Movie, error) {
	logger := c.logger.With(applog.String("Method", "GetMovie"), applog.Uint("movie_id", movieID))
	key := c.movieKey(movieID)

	jsonStr, err := c.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		logger.Info("movie not found in redis", applog.String("key", key))
		return nil, fmt.Errorf("%w: %w", ErrKeyNotFound, err)
	}
	if err != nil {
		logger.Error("failed to get movie from redis", applog.Error(err))
		return nil, fmt.Errorf("failed to get movie from redis: %w", err)
	}

	var movie movie.Movie
	if err := json.Unmarshal([]byte(jsonStr), &movie); err != nil {
		logger.Error("failed to unmarshal movie", applog.Error(err))
		return nil, fmt.Errorf("failed to unmarshal movie: %w", err)
	}
	logger.Info("get movie from redis successfully", applog.String("key", key))
	return &movie, nil
}

// SetMovieList 设置电影列表缓存
func (c *RedisMovieCache) SetMovieList(ctx context.Context, movies []*movie.Movie, params map[string]interface{}, expiration time.Duration) error {
	logger := c.logger.With(
		applog.String("Method", "SetMovieList"),
		applog.Any("params", params),
	)

	if expiration == 0 {
		expiration = c.defaultExpiration
	}

	// 序列化电影列表
	data, err := json.Marshal(movies)
	if err != nil {
		logger.Error("failed to marshal movie list", applog.Error(err))
		return fmt.Errorf("failed to marshal movie list: %w", err)
	}

	// 使用管道批量执行操作
	pipe := c.redisClient.Pipeline()

	// 设置列表缓存
	listKey := c.movieListKey(params)
	setKey := c.movieSetKey(params)

	// 1. 获取旧的电影ID集合
	oldMovieIDs, err := c.redisClient.SMembers(ctx, setKey).Result()
	if err != nil && err != redis.Nil {
		logger.Error("failed to get old movie IDs", applog.Error(err))
		return fmt.Errorf("failed to get old movie IDs: %w", err)
	}

	// 2. 从旧电影的反向索引中移除当前列表
	for _, idStr := range oldMovieIDs {
		movieID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			logger.Error("failed to parse movie ID", applog.Error(err))
			continue
		}
		indexKey := c.movieIndexKey(uint(movieID))
		pipe.SRem(ctx, indexKey, setKey)
	}

	// 3. 设置新的列表数据
	pipe.Set(ctx, listKey, data, expiration)

	// 4. 重置并更新电影ID集合
	pipe.Del(ctx, setKey)
	newMovieIDs := make([]interface{}, len(movies))
	for i, m := range movies {
		newMovieIDs[i] = m.ID
		// 5. 更新每个电影的反向索引
		indexKey := c.movieIndexKey(m.ID)
		pipe.SAdd(ctx, indexKey, setKey)
	}

	// 只有在有电影时才添加到集合
	if len(newMovieIDs) > 0 {
		pipe.SAdd(ctx, setKey, newMovieIDs...)
	}

	// 执行所有操作
	if _, err := pipe.Exec(ctx); err != nil {
		logger.Error("failed to set movie list", applog.Error(err))
		return fmt.Errorf("failed to set movie list: %w", err)
	}

	logger.Info("set movie list successfully",
		applog.String("list_key", listKey),
		applog.String("set_key", setKey),
		applog.Int("old_movies_count", len(oldMovieIDs)),
		applog.Int("new_movies_count", len(newMovieIDs)))
	return nil
}

// GetMovieList 获取电影列表缓存
func (c *RedisMovieCache) GetMovieList(ctx context.Context, params map[string]interface{}) ([]*movie.Movie, error) {
	logger := c.logger.With(
		applog.String("Method", "GetMovieList"),
		applog.Any("params", params),
	)

	listKey := c.movieListKey(params)
	valBytes, err := c.redisClient.Get(ctx, listKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.Info("movie list not found in redis", applog.String("key", listKey))
			return nil, fmt.Errorf("%w: %w", ErrKeyNotFound, err)
		}
		logger.Error("failed to get movie list from redis", applog.Error(err))
		return nil, fmt.Errorf("failed to get movie list from redis: %w", err)
	}

	var movies []*movie.Movie
	if err := json.Unmarshal(valBytes, &movies); err != nil {
		logger.Error("failed to unmarshal movie list", applog.Error(err))
		return nil, fmt.Errorf("failed to unmarshal movie list: %w", err)
	}

	// 更新访问信息
	setKey := c.movieSetKey(params)
	if exists, _ := c.redisClient.Exists(ctx, setKey).Result(); exists == 1 {
		// 如果集合存在，确保反向索引完整
		pipe := c.redisClient.Pipeline()
		for _, m := range movies {
			indexKey := c.movieIndexKey(m.ID)
			pipe.SAdd(ctx, indexKey, setKey)
		}
		if _, err := pipe.Exec(ctx); err != nil {
			logger.Warn("failed to update reverse indices", applog.Error(err))
		}
	}

	logger.Info("get movie list successfully",
		applog.String("list_key", listKey),
		applog.Int("movies_count", len(movies)))
	return movies, nil
}

// // InvalidateMovieList 根据给定的模式删除所有匹配的缓存键。
// func (c *RedisMovieCache) InvalidateMovieList(ctx context.Context, pattern string) error {
// 	logger := c.logger.With(applog.String("Method", "InvalidateMovieList"), applog.String("pattern", pattern))
// 	var cursor uint64
// 	var keysFound int
// 	for {
// 		var keys []string
// 		var err error
// 		keys, cursor, err = c.redisClient.Scan(ctx, cursor, pattern, 50).Result() // 每次扫描50个
// 		if err != nil {
// 			logger.Error("Error scanning movie list keys for invalidation", applog.Error(err))
// 			return fmt.Errorf("error scanning keys: %w", err)
// 		}

// 		if len(keys) > 0 {
// 			if errDel := c.redisClient.Del(ctx, keys...).Err(); errDel != nil {
// 				logger.Error("Error deleting movie list keys during invalidation", applog.Error(errDel))
// 				// 根据策略决定是否继续或返回错误
// 			}
// 			keysFound += len(keys)
// 		}
// 		if cursor == 0 { // 迭代完成
// 			break
// 		}
// 	}
// 	if keysFound > 0 {
// 		logger.Info("Invalidated movie list caches", applog.Int("count", keysFound))
// 	} else {
// 		logger.Debug("No movie list caches to invalidate with pattern")
// 	}
// 	return nil
// }
