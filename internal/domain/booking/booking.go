package booking

import "time"

// Booking 表示一个电影票订单
type Booking struct {
	ID          uint
	UserID      uint
	ShowtimeID  uint
	TotalAmount float64
	BookingTime time.Time
	Status      string
}

// 索引: 在(user_id, booking_time) 和 (showtime_id) 上创建索引
