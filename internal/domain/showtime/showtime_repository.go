package showtime

import (
	"context"
	"mrs/internal/domain/shared/vo"
	"time"
)

// 场次仓库接口
type ShowtimeRepository interface {
	Create(ctx context.Context, showtime *Showtime) (*Showtime, error)
	FindByID(ctx context.Context, id vo.ShowtimeID) (*Showtime, error)
	FindByIDs(ctx context.Context, ids []vo.ShowtimeID) ([]*Showtime, error)
	// 分页查询支持过滤条件（如电影ID/影厅ID/日期范围）
	List(ctx context.Context, options *ShowtimeQueryOptions) ([]*Showtime, int64, error)
	Update(ctx context.Context, showtime *Showtime) error
	Delete(ctx context.Context, id vo.ShowtimeID) error

	// 检查指定时间段内是否存在与给定影厅冲突的场次（排除指定ID的场次）
	CheckOverlap(ctx context.Context, hallID vo.CinemaHallID, startTime, endTime time.Time, excludeShowtimeID ...vo.ShowtimeID) (bool, error)
	// 查询指定电影在日期范围内的所有场次
	FindShowtimesByMovieAndDateRanges(ctx context.Context, movieID vo.MovieID,
		startDate, endDate time.Time) ([]*Showtime, error)
	// 查询指定影厅在日期范围内的所有场次
	FindShowtimesByHallAndDateRanges(ctx context.Context, hallID vo.CinemaHallID,
		startDate, endDate time.Time) ([]*Showtime, error)
}

// 放映查询选项
type ShowtimeQueryOptions struct {
	MovieID      vo.MovieID      // 电影ID
	CinemaHallID vo.CinemaHallID // 影厅ID
	Date         time.Time       // 日期
	Page         int             // 页码（从1开始）
	PageSize     int             // 每页数量
}
