package cinema

import (
	"context"
	"mrs/internal/domain/shared/vo"
)

type SeatRepository interface {
	CreateBatch(ctx context.Context, seats []*Seat) ([]*Seat, error)

	FindByID(ctx context.Context, id vo.SeatID) (*Seat, error)
	FindByHallID(ctx context.Context, hallID vo.CinemaHallID) ([]*Seat, error)
	GetSeatsByIDs(ctx context.Context, seatIDs []vo.SeatID) ([]*Seat, error)

	Update(ctx context.Context, seat *Seat) error

	Delete(ctx context.Context, id vo.SeatID) error
	DeleteByHallID(ctx context.Context, hallID vo.CinemaHallID) error
}
