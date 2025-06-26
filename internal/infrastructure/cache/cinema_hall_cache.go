package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/shared/vo"
	applog "mrs/pkg/log"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
)

type cinemaHallCache struct {
	redisClient RedisClient
	logger      applog.Logger
}

func NewCinemaHallCache(redisClient RedisClient, logger applog.Logger) cinema.CinemaHallCache {
	return &cinemaHallCache{
		redisClient: redisClient,
		logger:      logger.With(applog.String("Component", "CinemaHallCache")),
	}
}

// 设置单个影厅
func (c *cinemaHallCache) SetCinemaHall(ctx context.Context, cinemaHall *cinema.CinemaHall, expiration time.Duration) error {
	logger := c.logger.With(applog.String("Method", "SetCinemaHall"), applog.Uint("cinema_hall_id", uint(cinemaHall.ID)))
	key := cinema.GetCinemaHallCacheKey(cinemaHall.ID)
	valBytes, err := json.Marshal(cinemaHall)
	if err != nil {
		logger.Error("failed to marshal cinema hall", applog.Error(err))
		return fmt.Errorf("failed to marshal cinema hall: %w", err)
	}

	if err := c.redisClient.Set(ctx, key, valBytes, expiration).Err(); err != nil {
		logger.Error("failed to set cinema hall in redis", applog.Error(err))
		return fmt.Errorf("failed to set cinema hall in redis: %w", err)
	}
	logger.Info("set cinema hall in redis successfully", applog.String("key", key))
	return nil
}

// 获取单个影厅
func (c *cinemaHallCache) GetCinemaHall(ctx context.Context, id vo.CinemaHallID) (*cinema.CinemaHall, error) {
	logger := c.logger.With(applog.String("Method", "GetCinemaHall"), applog.Uint("cinema_hall_id", uint(id)))
	key := cinema.GetCinemaHallCacheKey(id)
	valBytes, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.Info("cinema hall not found in redis", applog.String("key", key))
			return nil, fmt.Errorf("%w: %w", shared.ErrCacheMissing, err)
		}
		logger.Error("failed to get cinema hall from redis", applog.Error(err))
		return nil, fmt.Errorf("failed to get cinema hall from redis: %w", err)
	}

	var cinemaHall cinema.CinemaHall
	if err := json.Unmarshal(valBytes, &cinemaHall); err != nil {
		logger.Error("failed to unmarshal cinema hall", applog.Error(err))
		return nil, fmt.Errorf("failed to unmarshal cinema hall: %w", err)
	}
	logger.Info("get cinema hall from redis successfully", applog.String("key", key))
	return &cinemaHall, nil
}

// 删除单个影厅
func (c *cinemaHallCache) DeleteCinemaHall(ctx context.Context, id vo.CinemaHallID) error {
	logger := c.logger.With(applog.String("Method", "DeleteCinemaHall"), applog.Uint("cinema_hall_id", uint(id)))
	key := cinema.GetCinemaHallCacheKey(id)

	if err := c.redisClient.Del(ctx, key).Err(); err != nil {
		logger.Error("failed to delete cinema hall in redis", applog.Error(err))
		return fmt.Errorf("failed to delete cinema hall in redis: %w", err)
	}
	logger.Info("delete cinema hall in redis successfully", applog.String("key", key))
	return nil
}

// 设置所有影厅
func (c *cinemaHallCache) SetAllCinemaHalls(ctx context.Context, cinemaHalls []*cinema.CinemaHall, expiration time.Duration) error {
	logger := c.logger.With(applog.String("Method", "SetAllCinemaHalls"))

	// 设置所有影厅的ID列表
	key := cinema.GetCinemaHallAllIDsKey()
	ids := make([]uint, len(cinemaHalls))
	for i, cinemaHall := range cinemaHalls {
		ids[i] = uint(cinemaHall.ID)
	}
	valBytes, err := json.Marshal(ids)
	if err != nil {
		logger.Error("failed to marshal cinema halls ids", applog.Error(err))
		return fmt.Errorf("failed to marshal cinema halls ids: %w", err)
	}
	if err := c.redisClient.Set(ctx, key, valBytes, expiration).Err(); err != nil {
		logger.Error("failed to set all cinema halls ids in redis", applog.Error(err))
		return fmt.Errorf("failed to set all cinema halls ids in redis: %w", err)
	}

	// 设置所有影厅的缓存
	pipe := c.redisClient.Pipeline()
	for _, cinemaHall := range cinemaHalls {
		valBytes, err := json.Marshal(cinemaHall)
		if err != nil {
			logger.Error("failed to marshal cinema hall", applog.Error(err))
			return fmt.Errorf("failed to marshal cinema hall: %w", err)
		}
		pipe.Set(ctx, cinema.GetCinemaHallCacheKey(cinemaHall.ID), valBytes, expiration)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Error("failed to set all cinema halls in redis", applog.Error(err))
		return fmt.Errorf("failed to set all cinema halls in redis: %w", err)
	}

	logger.Info("set all cinema halls in redis successfully")
	return nil
}

// 获取所有影厅
func (c *cinemaHallCache) GetAllCinemaHalls(ctx context.Context) (*cinema.CinemaHallCacheListResult, error) {
	logger := c.logger.With(applog.String("Method", "GetAllCinemaHalls"))
	key := cinema.GetCinemaHallAllIDsKey()
	result := &cinema.CinemaHallCacheListResult{}

	// 获取所有影厅的ID列表
	valBytes, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.Info("cinema halls ids not found in redis", applog.String("key", key))
			return nil, fmt.Errorf("%w: %w", shared.ErrCacheMissing, err)
		}
		logger.Error("failed to get cinema halls ids from redis", applog.Error(err))
		return nil, fmt.Errorf("failed to get cinema halls ids from redis: %w", err)
	}

	if err := json.Unmarshal(valBytes, &result.AllCinemaHallIDs); err != nil {
		logger.Error("failed to unmarshal cinema halls ids", applog.Error(err))
		return nil, fmt.Errorf("failed to unmarshal cinema halls ids: %w", err)
	}

	// 批量获取电影详情
	pipe := c.redisClient.Pipeline()
	cinemaHallCmds := make([]*redis.StringCmd, len(result.AllCinemaHallIDs))
	for i, id := range result.AllCinemaHallIDs {
		cinemaHallCmds[i] = pipe.Get(ctx, cinema.GetCinemaHallCacheKey(id))
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		logger.Error("failed to get cinema halls from redis", applog.Error(err))
		return nil, fmt.Errorf("failed to get cinema halls from redis: %w", err)
	}

	// 处理结果并记录缺失的电影ID
	missingIDs := make(map[uint]struct{})
	for i, cmd := range cinemaHallCmds {
		cinemaHallID := result.AllCinemaHallIDs[i]
		cinemaHallBytes, err := cmd.Bytes()
		if err != nil {
			logger.Warn("failed to get cinema hall bytes",
				applog.Error(err),
				applog.Uint("cinema_hall_id", uint(cinemaHallID)))
			missingIDs[uint(cinemaHallID)] = struct{}{}
			continue
		}

		var cinemaHall cinema.CinemaHall
		if err := json.Unmarshal(cinemaHallBytes, &cinemaHall); err != nil {
			logger.Warn("failed to unmarshal cinema hall",
				applog.Error(err),
				applog.Uint("cinema_hall_id", uint(cinemaHallID)))
			missingIDs[uint(cinemaHallID)] = struct{}{}
			continue
		}
		result.CinemaHalls = append(result.CinemaHalls, &cinemaHall)
	}

	// 将缺失的ID转换为切片并排序
	if len(missingIDs) > 0 {
		result.MissingCinemaHallIDs = make([]vo.CinemaHallID, 0, len(missingIDs))
		for id := range missingIDs {
			result.MissingCinemaHallIDs = append(result.MissingCinemaHallIDs, vo.CinemaHallID(id))
		}
		sort.Slice(result.MissingCinemaHallIDs, func(i, j int) bool {
			return result.MissingCinemaHallIDs[i] < result.MissingCinemaHallIDs[j]
		})
	}

	logger.Info("get all cinema halls successfully",
		applog.Int("total_ids", len(result.AllCinemaHallIDs)),
		applog.Int("found_cinema_halls", len(result.CinemaHalls)),
		applog.Int("missing_cinema_halls", len(result.MissingCinemaHallIDs)))

	return result, nil
}

// 删除所有影厅的ID列表
func (c *cinemaHallCache) DeleteAllCinemaHallIDs(ctx context.Context) error {
	logger := c.logger.With(applog.String("Method", "DeleteAllCinemaHallIDs"))
	key := cinema.GetCinemaHallAllIDsKey()
	if err := c.redisClient.Del(ctx, key).Err(); err != nil {
		logger.Error("failed to delete all cinema halls ids in redis", applog.Error(err))
		return fmt.Errorf("failed to delete all cinema halls ids in redis: %w", err)
	}
	logger.Info("delete all cinema halls ids in redis successfully")
	return nil
}
