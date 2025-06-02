package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"mrs/internal/domain/movie"
	applog "mrs/pkg/log"
	"sort"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	showtimeKeyPrefix     = "showtime:"
	showtimeListKeyPrefix = "showtimes:list:"
)

type ShowtimeListResult struct {
	Showtimes          []*movie.Showtime // 成功获取的放映记录
	AllShowtimeIDs     []uint            // 列表中所有的放映ID
	MissingShowtimeIDs []uint            // 缓存中未找到的放映ID
}

type RedisShowtimeCache struct {
	redisClient       RedisClient
	logger            applog.Logger
	defaultExpiration time.Duration
}

// 创建一个RedisShowtimeCache实例
func NewRedisShowtimeCache(redisClient RedisClient, logger applog.Logger, defaultExpiration time.Duration) *RedisShowtimeCache {
	return &RedisShowtimeCache{
		redisClient:       redisClient,
		logger:            logger.With(applog.String("Component", "RedisShowtimeCache")),
		defaultExpiration: defaultExpiration,
	}
}

// showtimeKey 生成单个放映的缓存键
func (c *RedisShowtimeCache) showtimeKey(showtimeID uint) string {
	return fmt.Sprintf("%s%d", showtimeKeyPrefix, showtimeID)
}

// showtimeListKey 生成放映列表的缓存键
func (c *RedisShowtimeCache) showtimeListKey(params map[string]interface{}) string {
	if len(params) == 0 {
		return showtimeListKeyPrefix + "all"
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(showtimeListKeyPrefix)
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%v:", k, params[k])) // 构建器追加字符串
	}
	return strings.TrimRight(sb.String(), ":") // 移除字符串右侧的:符号
}

// SetShowtime 设置单个放映的缓存
func (c *RedisShowtimeCache) SetShowtime(ctx context.Context, showtime *movie.Showtime, expiration time.Duration) error {
	logger := c.logger.With(applog.String("Method", "SetShowtime"), applog.Uint("showtime_id", showtime.ID))

	if expiration == 0 {
		expiration = c.defaultExpiration
	}

	data, err := json.Marshal(showtime)
	if err != nil {
		logger.Error("failed to marshal showtime", applog.Error(err))
		return fmt.Errorf("failed to marshal showtime: %w", err)
	}

	key := c.showtimeKey(showtime.ID)
	if err := c.redisClient.Set(ctx, key, data, expiration).Err(); err != nil {
		logger.Error("failed to set showtime", applog.Error(err))
		return fmt.Errorf("failed to set showtime: %w", err)
	}

	logger.Info("set showtime successfully", applog.String("key", key))
	return nil
}

// GetShowtime 获取单个放映的缓存
func (c *RedisShowtimeCache) GetShowtime(ctx context.Context, showtimeID uint) (*movie.Showtime, error) {
	logger := c.logger.With(applog.String("Method", "GetShowtime"), applog.Uint("showtime_id", showtimeID))
	key := c.showtimeKey(showtimeID)

	valBytes, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.Info("showtime not found in redis", applog.String("key", key))
			return nil, fmt.Errorf("%w: %w", ErrKeyNotFound, err)
		}
		logger.Error("failed to get showtime from redis", applog.Error(err))
		return nil, fmt.Errorf("failed to get showtime from redis: %w", err)
	}

	var showtime movie.Showtime
	if err := json.Unmarshal(valBytes, &showtime); err != nil {
		logger.Error("failed to unmarshal showtime", applog.Error(err))
		return nil, fmt.Errorf("failed to unmarshal showtime: %w", err)
	}
	logger.Info("get showtime from redis successfully", applog.String("key", key))
	return &showtime, nil
}

func (c *RedisShowtimeCache) DeleteShowtime(ctx context.Context, showtimeID uint) error {
	logger := c.logger.With(applog.String("Method", "DeleteShowtime"), applog.Uint("showtime_id", showtimeID))
	key := c.showtimeKey(showtimeID)
	if err := c.redisClient.Del(ctx, key).Err(); err != nil {
		// 如果缓存不存在，则返回错误
		if err == redis.Nil {
			logger.Info("showtime not found in redis", applog.String("key", key))
			return fmt.Errorf("%w: %w", ErrKeyNotFound, err)
		}
		logger.Error("failed to delete showtime", applog.Error(err))
		return fmt.Errorf("failed to delete showtime: %w", err)
	}

	logger.Info("delete showtime successfully", applog.String("key", key))
	return nil
}

// SetShowtimes 批量设置多个放映的缓存
func (c *RedisShowtimeCache) SetShowtimes(ctx context.Context, showtimes []*movie.Showtime, expiration time.Duration) error {
	logger := c.logger.With(applog.String("Method", "SetShowtimes"), applog.Int("showtimes_count", len(showtimes)))

	if expiration == 0 {
		expiration = c.defaultExpiration
	}

	pipe := c.redisClient.Pipeline()
	for _, showtime := range showtimes {
		data, err := json.Marshal(showtime)
		if err != nil {
			logger.Error("failed to marshal showtime", applog.Error(err))
			continue
		}
		pipe.Set(ctx, c.showtimeKey(showtime.ID), data, expiration)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Error("failed to set showtimes", applog.Error(err))
		return fmt.Errorf("failed to set showtimes: %w", err)
	}

	logger.Info("set showtimes successfully", applog.Int("showtimes_count", len(showtimes)))
	return nil
}

// SetShowtimeList 设置放映列表的缓存
func (c *RedisShowtimeCache) SetShowtimeList(ctx context.Context, showtimes []*movie.Showtime,
	params map[string]interface{}, expiration time.Duration) error {
	logger := c.logger.With(applog.String("Method", "SetShowtimeList"), applog.Any("params", params))

	if expiration == 0 {
		expiration = c.defaultExpiration
	}

	// 提取电影ID列表
	showtimeIDs := make([]uint, len(showtimes))
	for i, s := range showtimes {
		showtimeIDs[i] = s.ID
	}

	// 序列化ID列表
	data, err := json.Marshal(showtimeIDs)
	if err != nil {
		logger.Error("failed to marshal showtime ID list", applog.Error(err))
		return fmt.Errorf("failed to marshal showtime ID list: %w", err)
	}

	// 设置列表缓存（仅包含ID）
	listKey := c.showtimeListKey(params)
	if err := c.redisClient.Set(ctx, listKey, data, expiration).Err(); err != nil {
		logger.Error("failed to set showtime ID list", applog.Error(err))
		return fmt.Errorf("failed to set showtime ID list: %w", err)
	}

	// 同时缓存单个电影记录
	if err := c.SetShowtimes(ctx, showtimes, expiration); err != nil {
		logger.Error("failed to set individual showtimes", applog.Error(err))
		return fmt.Errorf("failed to set individual showtimes: %w", err)
	}

	logger.Info("set showtime_id list and individual showtimes successfully",
		applog.String("list_key", listKey),
		applog.Int("showtimes_count", len(showtimes)))

	return nil
}

func (c *RedisShowtimeCache) GetShowtimeList(ctx context.Context, params map[string]interface{}) (*ShowtimeListResult, error) {
	logger := c.logger.With(
		applog.String("Method", "GetShowtimeList"),
		applog.Any("params", params),
	)

	result := &ShowtimeListResult{
		Showtimes: make([]*movie.Showtime, 0),
	}

	// 获取ID列表
	listKey := c.showtimeListKey(params)
	valBytes, err := c.redisClient.Get(ctx, listKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.Info("showtime_id list not found in redis", applog.String("key", listKey))
			return nil, fmt.Errorf("%w: %w", ErrKeyNotFound, err)
		}
		logger.Error("failed to get showtime_id list from redis", applog.Error(err))
		return nil, fmt.Errorf("failed to get showtime_id list from redis: %w", err)
	}

	if err := json.Unmarshal(valBytes, &result.AllShowtimeIDs); err != nil {
		logger.Error("failed to unmarshal showtime_id list", applog.Error(err))
		return nil, fmt.Errorf("failed to unmarshal showtime_id list: %w", err)
	}

	if len(result.AllShowtimeIDs) == 0 {
		logger.Info("empty showtime_id list in redis", applog.String("key", listKey))
		return result, nil
	}

	// 批量获取电影详情
	pipe := c.redisClient.Pipeline()
	showtimeCmds := make([]*redis.StringCmd, len(result.AllShowtimeIDs))
	for i, id := range result.AllShowtimeIDs {
		showtimeCmds[i] = pipe.Get(ctx, c.showtimeKey(id))
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		logger.Error("failed to get showtimes from redis", applog.Error(err))
		return nil, fmt.Errorf("failed to get showtimes from redis: %w", err)
	}

	// 处理结果并记录缺失的电影ID
	missingIDs := make(map[uint]struct{})
	for i, cmd := range showtimeCmds {
		showtimeID := result.AllShowtimeIDs[i]
		showtimeBytes, err := cmd.Bytes()
		if err != nil {
			logger.Warn("failed to get showtime details",
				applog.Error(err),
				applog.Uint("showtime_id", showtimeID))
			missingIDs[showtimeID] = struct{}{}
			continue
		}

		var showtime movie.Showtime
		if err := json.Unmarshal(showtimeBytes, &showtime); err != nil {
			logger.Warn("failed to unmarshal showtime",
				applog.Error(err),
				applog.Uint("showtime_id", showtimeID))
			missingIDs[showtimeID] = struct{}{}
			continue
		}
		result.Showtimes = append(result.Showtimes, &showtime)
	}

	// 将缺失的ID转换为切片并排序
	if len(missingIDs) > 0 {
		result.MissingShowtimeIDs = make([]uint, 0, len(missingIDs))
		for id := range missingIDs {
			result.MissingShowtimeIDs = append(result.MissingShowtimeIDs, id)
		}
		sort.Slice(result.MissingShowtimeIDs, func(i, j int) bool {
			return result.MissingShowtimeIDs[i] < result.MissingShowtimeIDs[j]
		})
	}

	logger.Info("get showtime list successfully",
		applog.String("list_key", listKey),
		applog.Int("total_ids", len(result.AllShowtimeIDs)),
		applog.Int("found_showtimes", len(result.Showtimes)),
		applog.Int("missing_showtimes", len(result.MissingShowtimeIDs)))

	return result, nil
}
