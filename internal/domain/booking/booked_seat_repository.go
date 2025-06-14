package booking

import (
	"context"
	"mrs/internal/domain/shared/vo"
)

type BookedSeatRepository interface {
	CreateBatch(ctx context.Context, bookedSeats []*BookedSeat) ([]*BookedSeat, error)
	FindByID(ctx context.Context, id vo.BookedSeatID) (*BookedSeat, error)
	FindByBookingID(ctx context.Context, bookingID vo.BookingID) ([]*BookedSeat, error)
	Update(ctx context.Context, bookedSeat *BookedSeat) error
	Delete(ctx context.Context, id vo.BookedSeatID) error
	DeleteByBookingID(ctx context.Context, bookingID vo.BookingID) error
}
