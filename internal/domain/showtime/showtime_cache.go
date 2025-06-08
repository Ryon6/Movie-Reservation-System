package showtime

import (
	"context"
	"time"
)

// ShowtimeCache 放映缓存接口
type ShowtimeCache interface {
	GetShowtime(ctx context.Context, showtimeID uint) (*Showtime, error)
	SetShowtime(ctx context.Context, showtime *Showtime, expiration time.Duration) error
	DeleteShowtime(ctx context.Context, showtimeID uint) error
	GetShowtimeList(ctx context.Context, options *ShowtimeQueryOptions) (*ShowtimeListResult, error)
	SetShowtimeList(ctx context.Context, showtimes []*Showtime, options *ShowtimeQueryOptions, expiration time.Duration) error
}

// ShowtimeListResult 放映列表的查询结果
type ShowtimeListResult struct {
	Showtimes          []*Showtime // 成功获取的放映记录
	AllShowtimeIDs     []uint      // 列表中所有的放映ID
	MissingShowtimeIDs []uint      // 缓存中未找到的放映ID
}
