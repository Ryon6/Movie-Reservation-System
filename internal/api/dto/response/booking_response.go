package response

import (
	"math"
	"mrs/internal/api/dto/request"
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

// ListBookingsResponse 表示一个订单列表的响应
type ListBookingsResponse struct {
	Bookings []*BookingResponse `json:"bookings"`
	PaginationResponse
}

// ToListBookingsResponse 将domain层的订单列表转换为API层的订单列表
func ToListBookingsResponse(bookings []*booking.Booking, totalCount int, req *request.PaginationRequest) *ListBookingsResponse {
	response := make([]*BookingResponse, len(bookings))
	for i, booking := range bookings {
		response[i] = ToBookingResponse(booking)
	}
	return &ListBookingsResponse{
		Bookings: response,
		PaginationResponse: PaginationResponse{
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalCount: totalCount,
			TotalPages: int(math.Ceil(float64(totalCount) / float64(req.PageSize))),
		},
	}
}
