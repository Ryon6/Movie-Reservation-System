package request

import (
	"mrs/internal/domain/booking"
	"mrs/internal/domain/shared/vo"
)

// 创建订单请求
type CreateBookingRequest struct {
	UserID     uint
	ShowtimeID uint   `json:"showtime_id"`
	SeatIDs    []uint `json:"seat_ids"`
}

type GetBookingRequest struct {
	ID uint
}

// 查询订单请求
type ListBookingsRequest struct {
	PaginationRequest
	UserID uint
	Status string `json:"status" form:"status" binding:"omitempty,min=1,max=255"`
}

func (r *ListBookingsRequest) ToDomain() *booking.BookingQueryOptions {
	return &booking.BookingQueryOptions{
		UserID:   vo.UserID(r.UserID),
		Status:   booking.BookingStatus(r.Status),
		Page:     r.Page,
		PageSize: r.PageSize,
	}
}

// 取消订单请求
type CancelBookingRequest struct {
	ID uint
}

// 确认订单请求
type ConfirmBookingRequest struct {
	ID uint
}
