package booking

import (
	"mrs/internal/domain/shared/vo"
	"time"
)

// 订单状态枚举
type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCanceled  BookingStatus = "canceled"
)

// Booking 表示一个电影票订单
type Booking struct {
	ID          vo.BookingID
	UserID      vo.UserID
	ShowtimeID  vo.ShowtimeID
	BookedSeats []*BookedSeat
	TotalAmount float64
	BookingTime time.Time
	Status      BookingStatus
}

func NewBooking(userID vo.UserID, showtimeID vo.ShowtimeID, bookedSeats []*BookedSeat, totalAmount float64) *Booking {
	return &Booking{
		UserID:      userID,
		ShowtimeID:  showtimeID,
		BookedSeats: bookedSeats,
		TotalAmount: totalAmount,
		BookingTime: time.Now(),
		Status:      BookingStatusPending,
	}
}

// 确认订单
func (b *Booking) Confirm() {
	b.Status = BookingStatusConfirmed
}

// 取消订单
func (b *Booking) Cancel() {
	b.Status = BookingStatusCanceled
}

// 索引: 在(user_id, booking_time) 和 (showtime_id) 上创建索引
