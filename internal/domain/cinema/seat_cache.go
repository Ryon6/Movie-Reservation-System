package cinema

import (
	"context"
	"fmt"
	"mrs/internal/domain/shared/vo"
	"time"
)

// 关于座位缓存：将动态信息座位状态存储为位图，静态信息存储为哈希表

// 座位缓存接口
type SeatCache interface {
	LockSeats(ctx context.Context, showtimeID vo.ShowtimeID, seatIDs []vo.SeatID) error
	ReleaseSeats(ctx context.Context, showtimeID vo.ShowtimeID, seatIDs []vo.SeatID) error
	GetSeatMap(ctx context.Context, showtimeID vo.ShowtimeID) ([]*SeatInfo, error)
	InitSeatMap(ctx context.Context, showtimeID vo.ShowtimeID, hallLayout []*Seat, bookedSeatIDs []vo.SeatID, expireTime time.Duration) error
	InvalidateSeatMap(ctx context.Context, showtimeID vo.ShowtimeID) error // 失效座位表（大多数情况下，座位表会自动过期，但若修改座位时需要手动失效）
}

const (
	ShowtimeSeatsBitmapKeyFormat   = "showtime:%d:seats:bitmap"    // 场次座位状态位图
	ShowtimeSeatsInfoKeyFormat     = "showtime:%d:seats:info"      // 场次座位静态信息
	ShowtimeSeatsLockKeyFormat     = "showtime:%d:seats:locks"     // 座位临时锁定
	ShowtimeSeatsInitLockKeyFormat = "showtime:%d:seats:init:lock" // 初始化座位表的锁，防止并发初始化座位表
)

// 生成座位临时锁定的缓存键
func GetShowtimeSeatsLockKey(showtimeID vo.ShowtimeID) string {
	return fmt.Sprintf(ShowtimeSeatsLockKeyFormat, showtimeID)
}

// 生成座位初始化锁定的缓存键
func GetShowtimeSeatsInitLockKey(showtimeID vo.ShowtimeID) string {
	return fmt.Sprintf(ShowtimeSeatsInitLockKeyFormat, showtimeID)
}

// 生成座位状态位图的缓存键
func GetShowtimeSeatsBitmapKey(showtimeID vo.ShowtimeID) string {
	return fmt.Sprintf(ShowtimeSeatsBitmapKeyFormat, showtimeID)
}

// 生成座位静态信息的缓存键
func GetShowtimeSeatsInfoKey(showtimeID vo.ShowtimeID) string {
	return fmt.Sprintf(ShowtimeSeatsInfoKeyFormat, showtimeID)
}
