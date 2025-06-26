package cinema

import (
	"context"
	"fmt"
	"mrs/internal/domain/shared/vo"
	"time"
)

// CinemaHallCache 影厅缓存接口
type CinemaHallCache interface {
	GetCinemaHall(ctx context.Context, id vo.CinemaHallID) (*CinemaHall, error)
	SetCinemaHall(ctx context.Context, cinemaHall *CinemaHall, expiration time.Duration) error
	DeleteCinemaHall(ctx context.Context, id vo.CinemaHallID) error
	GetAllCinemaHalls(ctx context.Context) (*CinemaHallCacheListResult, error)
	SetAllCinemaHalls(ctx context.Context, cinemaHalls []*CinemaHall, expiration time.Duration) error
	DeleteAllCinemaHallIDs(ctx context.Context) error
}

const (
	DefaultCinemaHallExpiration     = 1 * time.Hour
	DefaultCinemaHallListExpiration = 30 * time.Minute
)

const (
	CinemaHallCacheKeyFormat = "cinema_hall:%d"
	CinemaHallAllIDsKey      = "cinema_hall:all_ids"
)

func GetCinemaHallCacheKey(id vo.CinemaHallID) string {
	return fmt.Sprintf(CinemaHallCacheKeyFormat, id)
}

func GetCinemaHallAllIDsKey() string {
	return CinemaHallAllIDsKey
}

// CinemaHallCacheListResult 影厅列表的查询结果
type CinemaHallCacheListResult struct {
	CinemaHalls          []*CinemaHall     // 成功获取的影厅记录
	AllCinemaHallIDs     []vo.CinemaHallID // 列表中所有的影厅ID
	MissingCinemaHallIDs []vo.CinemaHallID // 缓存中未找到的影厅ID
}
