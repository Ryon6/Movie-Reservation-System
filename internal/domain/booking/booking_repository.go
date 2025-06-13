package booking

import (
	"context"
	"mrs/internal/domain/shared/vo"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *Booking) (*Booking, error)
	FindByID(ctx context.Context, id vo.BookingID) (*Booking, error)
	FindByUserID(ctx context.Context, userID vo.UserID) ([]*Booking, error)
	FindByShowtimeID(ctx context.Context, showtimeID vo.ShowtimeID) ([]*Booking, error)
	List(ctx context.Context, options *BookingQueryOptions) ([]*Booking, int64, error)
	Update(ctx context.Context, booking *Booking) error
	Delete(ctx context.Context, id vo.BookingID) error
}

// BookingQueryOptions 表示查询订单的选项
type BookingQueryOptions struct {
	UserID   vo.UserID
	Status   BookingStatus
	Page     int
	PageSize int
}
