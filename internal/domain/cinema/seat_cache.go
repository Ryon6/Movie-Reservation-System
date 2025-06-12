package cinema

import (
	"context"
	"mrs/internal/domain/shared/vo"
)

// 关于座位缓存：将动态信息座位状态存储为位图，静态信息存储为哈希表

// 座位缓存接口
type SeatCache interface {
	LockSeats(ctx context.Context, showtimeID vo.ShowtimeID, seatIDs []vo.SeatID) error
	ReleaseSeats(ctx context.Context, showtimeID vo.ShowtimeID, seatIDs []vo.SeatID) error
	GetSeatMap(ctx context.Context, showtimeID vo.ShowtimeID) ([]SeatWithStatus, error)
	SetSeatMap(ctx context.Context, showtimeID vo.ShowtimeID, seats []SeatWithStatus) error
	// DeleteSeatMap(ctx context.Context, showtimeID vo.ShowtimeID) error // 自动过期
}
