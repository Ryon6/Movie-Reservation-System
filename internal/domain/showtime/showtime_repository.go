package showtime

import (
	"context"
	"time"
)

// 场次仓库接口
type ShowtimeRepository interface {
	Create(ctx context.Context, showtime *Showtime) (*Showtime, error)
	FindByID(ctx context.Context, id uint) (*Showtime, error)
	FindByIDs(ctx context.Context, ids []uint) ([]*Showtime, error)
	// 分页查询支持过滤条件（如电影ID/影厅ID/日期范围）
	List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*Showtime, int64, error)
	Update(ctx context.Context, showtime *Showtime) error
	Delete(ctx context.Context, id uint) error

	// 检查指定时间段内是否存在与给定影厅冲突的场次（排除指定ID的场次）
	CheckOverlap(ctx context.Context, hallID uint, startTime, endTime time.Time, excludeShowtimeID ...uint) (bool, error)
	// 查询指定电影在日期范围内的所有场次
	FindShowtimesByMovieAndDateRanges(ctx context.Context, movieID uint,
		startDate, endDate time.Time) ([]*Showtime, error)
	// 查询指定影厅在日期范围内的所有场次
	FindShowtimesByHallAndDateRanges(ctx context.Context, hallID uint,
		startDate, endDate time.Time) ([]*Showtime, error)
}
