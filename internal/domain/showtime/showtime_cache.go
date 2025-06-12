package showtime

import (
	"context"
	"mrs/internal/domain/shared/vo"
	"time"
)

const (
	DefaultExpiration = 10 * time.Minute
)

// ShowtimeCache 放映缓存接口
type ShowtimeCache interface {
	GetShowtime(ctx context.Context, showtimeID vo.ShowtimeID) (*Showtime, error)
	SetShowtime(ctx context.Context, showtime *Showtime, expiration time.Duration) error
	DeleteShowtime(ctx context.Context, showtimeID vo.ShowtimeID) error
	GetShowtimeList(ctx context.Context, options *ShowtimeQueryOptions) (*ShowtimeListResult, error)
	SetShowtimeList(ctx context.Context, showtimes []*Showtime, options *ShowtimeQueryOptions, expiration time.Duration) error
}

// ShowtimeListResult 放映列表的查询结果
type ShowtimeListResult struct {
	Showtimes          []*Showtime     // 成功获取的放映记录
	AllShowtimeIDs     []vo.ShowtimeID // 列表中所有的放映ID
	MissingShowtimeIDs []vo.ShowtimeID // 缓存中未找到的放映ID
}
