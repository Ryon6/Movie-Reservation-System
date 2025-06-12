package response

import (
	"mrs/internal/domain/booking"
	"time"
)

// BookingResponse 表示一个订单的响应
type BookingResponse struct {
	ID          uint      `json:"id"`
	TotalAmount float64   `json:"total_amount"`
	BookingTime time.Time `json:"booking_time"`
	Status      string    `json:"status"`
}

func ToBookingResponse(booking *booking.Booking) *BookingResponse {
	return &BookingResponse{
		ID:          uint(booking.ID),
		TotalAmount: booking.TotalAmount,
		BookingTime: booking.BookingTime,
		Status:      string(booking.Status),
	}
}
