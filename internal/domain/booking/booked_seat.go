package booking

import "mrs/internal/domain/shared/vo"

// 实现座位锁定和防止超额预订的核心逻辑

// BookedSeat 表示一个已预订的座位
type BookedSeat struct {
	ID        vo.BookedSeatID // 主键
	BookingID vo.BookingID    // 关联的预订ID
	SeatID    vo.SeatID       // 关联的座位ID
	Price     float64         // 座位价格
}

// 约束：对于(ShowtimeID, SeatID)组合应有唯一约束
