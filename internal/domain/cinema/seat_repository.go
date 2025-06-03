package cinema

import "context"

type SeatRepository interface {
	CreateBatch(ctx context.Context, seats []*Seat) error

	FindByID(ctx context.Context, id uint) (*Seat, error)
	FindByHallID(ctx context.Context, hallID uint) ([]*Seat, error)
	GetSeatsByIDs(ctx context.Context, seatIDs []uint) ([]*Seat, error)

	Update(ctx context.Context, seat *Seat) error

	Delete(ctx context.Context, id uint) error
	DeleteByHallID(ctx context.Context, hallID uint) error
}
