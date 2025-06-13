package booking

import (
	"context"
	"mrs/internal/domain/shared/vo"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *Booking) (*Booking, error)
	GetBookingByID(ctx context.Context, id vo.BookingID) (*Booking, error)
	GetBookingsByUserID(ctx context.Context, userID vo.UserID) ([]*Booking, error)
	GetBookingsByShowtimeID(ctx context.Context, showtimeID vo.ShowtimeID) ([]*Booking, error)
	ListBookings(ctx context.Context, options *BookingQueryOptions) ([]*Booking, int64, error)
	UpdateBooking(ctx context.Context, booking *Booking) error
	DeleteBooking(ctx context.Context, id vo.BookingID) error
}

// BookingQueryOptions 表示查询订单的选项
type BookingQueryOptions struct {
	UserID   vo.UserID
	Status   BookingStatus
	Page     int
	PageSize int
}
