package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mrs/internal/domain/booking"
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/shared/vo"
	applog "mrs/pkg/log"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisSeatCache struct {
	client            RedisClient
	logger            applog.Logger
	defaultExpiration time.Duration
}

func NewRedisSeatCache(client RedisClient, logger applog.Logger, defaultExpiration time.Duration) cinema.SeatCache {
	return &RedisSeatCache{
		client:            client,
		logger:            logger.With(applog.String("Component", "RedisSeatCache")),
		defaultExpiration: defaultExpiration,
	}
}

func (c *RedisSeatCache) getHallLayoutAndMapping(ctx context.Context, showtimeID vo.ShowtimeID) ([]*cinema.Seat, map[vo.SeatID]uint, error) {
	logger := c.logger.With(applog.String("Method", "getHallLayoutAndMapping"), applog.Uint("ShowtimeID", uint(showtimeID)))
	seatInfoKey := cinema.GetShowtimeSeatsInfoKey(showtimeID)

	seatInfo, err := c.client.Get(ctx, seatInfoKey).Result()
	if err != nil {
		logger.Warn("redis get seat info error", applog.Error(shared.ErrCacheMissing), applog.String("key", seatInfoKey))
		return nil, nil, fmt.Errorf("redis get seat info error: %w", shared.ErrCacheMissing)
	}

	var hallLayout []*cinema.Seat
	if err := json.Unmarshal([]byte(seatInfo), &hallLayout); err != nil {
		logger.Error("json unmarshal seat info error", applog.Error(err))
		return nil, nil, fmt.Errorf("json unmarshal seat info error: %w", err)
	}

	idToOffset := make(map[vo.SeatID]uint)
	for i, seat := range hallLayout {
		idToOffset[seat.ID] = uint(i)
	}

	return hallLayout, idToOffset, nil
}

// 初始化座位表（需要上层提前获取分布式锁）
func (c *RedisSeatCache) InitSeatMap(
	ctx context.Context,
	showtimeID vo.ShowtimeID,
	hallLayout []*cinema.Seat,
	bookedSeatIDs []vo.SeatID,
	expireTime time.Duration) error {

	logger := c.logger.With(applog.String("Method", "InitSeatMap"), applog.Uint("ShowtimeID", uint(showtimeID)))

	// ID归一化
	sort.Slice(hallLayout, func(i, j int) bool {
		return hallLayout[i].ID < hallLayout[j].ID
	})
	idToOffset := make(map[vo.SeatID]uint)
	for i, seat := range hallLayout {
		idToOffset[seat.ID] = uint(i)
	}

	staticJson, err := json.Marshal(hallLayout)
	if err != nil {
		logger.Error("json marshal hall layout error", applog.Error(err))
		return fmt.Errorf("json marshal hall layout error: %w", err)
	}

	seatBitmapKey := cinema.GetShowtimeSeatsBitmapKey(showtimeID)
	seatInfoKey := cinema.GetShowtimeSeatsInfoKey(showtimeID)

	pipe := c.client.Pipeline()
	pipe.Set(ctx, seatInfoKey, staticJson, expireTime)
	pipe.Del(ctx, seatBitmapKey) // 确保从干净的位图开始
	// 预先置位图，保证bookedSeatIDs为空时，该位图仍存在
	pipe.SetBit(ctx, seatBitmapKey, 0, 0)
	for _, id := range bookedSeatIDs {
		if offset, ok := idToOffset[id]; ok {
			pipe.SetBit(ctx, seatBitmapKey, int64(offset), 1)
		}
	}
	pipe.Expire(ctx, seatBitmapKey, expireTime)

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Error("redis exec pipe error", applog.Error(err))
		return fmt.Errorf("redis exec pipe error: %w", err)
	}

	logger.Info("init seat map success")
	return nil
}

// 获取座位表
func (c *RedisSeatCache) GetSeatMap(ctx context.Context, showtimeID vo.ShowtimeID) ([]*cinema.SeatInfo, error) {
	logger := c.logger.With(applog.String("Method", "GetSeatMap"), applog.Uint("ShowtimeID", uint(showtimeID)))
	seatInfoKey := cinema.GetShowtimeSeatsInfoKey(showtimeID)
	seatBitmapKey := cinema.GetShowtimeSeatsBitmapKey(showtimeID)

	hallLayout, _, err := c.getHallLayoutAndMapping(ctx, showtimeID)
	if err != nil {
		if errors.Is(err, shared.ErrCacheMissing) {
			logger.Info("seat map not found in redis", applog.String("key", seatInfoKey))
			return nil, fmt.Errorf("%w: %w", shared.ErrCacheMissing, err)
		}
		logger.Error("get hall layout and mapping error", applog.Error(err))
		return nil, fmt.Errorf("get hall layout and mapping error: %w", err)
	}

	bitmapBytes, err := c.client.Get(ctx, seatBitmapKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.Info("seat bitmap not found in redis", applog.String("key", seatBitmapKey))
			return nil, fmt.Errorf("seat bitmap not found in redis: %w", shared.ErrCacheMissing)
		}
		logger.Error("redis get seat bitmap error", applog.Error(err))
		return nil, fmt.Errorf("redis get seat bitmap error: %w", err)
	}

	seatInfos := make([]*cinema.SeatInfo, len(hallLayout))
	for i, seat := range hallLayout {
		status := cinema.SeatStatusAvailable
		offset := uint(i)
		byteIndex := offset / 8
		bitIndex := 7 - (offset % 8)

		if byteIndex < uint(len(bitmapBytes)) && (bitmapBytes[byteIndex]>>bitIndex)&1 == 1 {
			status = cinema.SeatStatusLocked
		}

		seatInfos[i] = &cinema.SeatInfo{
			ID:            seat.ID,
			RowIdentifier: seat.RowIdentifier,
			SeatNumber:    seat.SeatNumber,
			Type:          seat.Type,
			Status:        status,
		}
	}
	logger.Info("get seat map success")
	return seatInfos, nil
}

// 检查座位是否已被锁定
var lockSeatScript = redis.NewScript(`
	for i=1, #ARGV do
		if redis.call("getbit", KEYS[1], ARGV[i]) == 1 then
			return 0
		end
	end
	for i=1, #ARGV do
		redis.call("setbit", KEYS[1], ARGV[i], 1)
	end
	return 1
`)

// 锁定座位
func (c *RedisSeatCache) LockSeats(ctx context.Context, showtimeID vo.ShowtimeID, seatIDs []vo.SeatID) error {
	logger := c.logger.With(applog.String("Method", "LockSeats"), applog.Uint("ShowtimeID", uint(showtimeID)))
	seatBitmapKey := cinema.GetShowtimeSeatsBitmapKey(showtimeID)

	// 获取座位表及座位ID到偏移量的映射
	_, idToOffset, err := c.getHallLayoutAndMapping(ctx, showtimeID)
	if err != nil {
		logger.Error("get hall layout and mapping error", applog.Error(err))
		return fmt.Errorf("get hall layout and mapping error: %w", err)
	}

	// 检查座位ID是否存在，不存在则返回错误
	offsetArgs := make([]interface{}, 0, len(seatIDs))
	for _, seatID := range seatIDs {
		offset, ok := idToOffset[seatID]
		if !ok {
			logger.Error("seat id not found in redis", applog.Uint("SeatID", uint(seatID)))
			return fmt.Errorf("seat id not found in redis: %w", shared.ErrCacheMissing)
		}
		offsetArgs = append(offsetArgs, offset)
	}

	res, err := lockSeatScript.Run(ctx, c.client.(*redis.Client), []string{seatBitmapKey}, offsetArgs...).Int()
	if err != nil {
		logger.Error("redis eval error", applog.Error(err))
		return fmt.Errorf("redis eval error: %w", err)
	}
	if res == 0 {
		logger.Error("redis lock seat error", applog.Error(booking.ErrBookedSeatAlreadyLocked))
		return fmt.Errorf("redis lock seat error: %w", booking.ErrBookedSeatAlreadyLocked)
	}

	logger.Info("lock seats successfully")
	return nil
}

// 释放座位
func (c *RedisSeatCache) ReleaseSeats(ctx context.Context, showtimeID vo.ShowtimeID, seatIDs []vo.SeatID) error {
	logger := c.logger.With(applog.String("Method", "ReleaseSeats"), applog.Uint("ShowtimeID", uint(showtimeID)))
	seatBitmapKey := cinema.GetShowtimeSeatsBitmapKey(showtimeID)

	// 获取座位表及座位ID到偏移量的映射
	_, idToOffset, err := c.getHallLayoutAndMapping(ctx, showtimeID)
	if err != nil {
		logger.Error("get hall layout and mapping error", applog.Error(err))
		return fmt.Errorf("get hall layout and mapping error: %w", err)
	}

	pipe := c.client.Pipeline()
	for _, seatID := range seatIDs {
		offset, ok := idToOffset[seatID]
		if !ok {
			logger.Warn("seat id not found in redis", applog.Uint("SeatID", uint(seatID)))
			continue // 座位不存在，忽略
		}
		pipe.SetBit(ctx, seatBitmapKey, int64(offset), 0)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Error("redis exec pipe error", applog.Error(err))
		return fmt.Errorf("redis exec pipe error: %w", err)
	}

	logger.Info("release seats successfully")
	return nil
}

// 失效座位表
func (c *RedisSeatCache) InvalidateSeatMap(ctx context.Context, showtimeID vo.ShowtimeID) error {
	logger := c.logger.With(applog.String("Method", "InvalidateSeatMap"), applog.Uint("ShowtimeID", uint(showtimeID)))
	seatBitmapKey := cinema.GetShowtimeSeatsBitmapKey(showtimeID)
	seatInfoKey := cinema.GetShowtimeSeatsInfoKey(showtimeID)

	if err := c.client.Del(ctx, seatBitmapKey, seatInfoKey).Err(); err != nil {
		logger.Error("redis del error", applog.Error(err))
		return fmt.Errorf("redis del error: %w", err)
	}

	logger.Info("invalidate seat map successfully")
	return nil
}
