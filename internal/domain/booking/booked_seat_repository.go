package booking

import (
	"context"
	"mrs/internal/domain/shared/vo"
)

type BookedSeatRepository interface {
	CreateBookedSeats(ctx context.Context, bookedSeats []*BookedSeat) ([]*BookedSeat, error)
	GetBookedSeatByID(ctx context.Context, id vo.BookedSeatID) (*BookedSeat, error)
	GetBookedSeatsByBookingID(ctx context.Context, bookingID vo.BookingID) ([]*BookedSeat, error)
	UpdateBookedSeat(ctx context.Context, bookedSeat *BookedSeat) error
	DeleteBookedSeat(ctx context.Context, id vo.BookedSeatID) error
}
