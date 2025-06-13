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

// 查询订单请求
type ListBookingsRequest struct {
	PaginationRequest
	UserID uint
	Status string `json:"status" binding:"omitempty,min=1,max=255"`
}

func (r *ListBookingsRequest) ToDomain() *booking.BookingQueryOptions {
	return &booking.BookingQueryOptions{
		UserID:   vo.UserID(r.UserID),
		Status:   booking.BookingStatus(r.Status),
		Page:     r.Page,
		PageSize: r.PageSize,
	}
}
