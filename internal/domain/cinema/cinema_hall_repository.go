package cinema

import (
	"context"
)

type CinemaHallRepository interface {
	Create(ctx context.Context, hall *CinemaHall) error
	FindByID(ctx context.Context, id uint) (*CinemaHall, error)
	FindByName(ctx context.Context, name string) (*CinemaHall, error)
	ListAll(ctx context.Context) ([]*CinemaHall, error)
	Update(ctx context.Context, hall *CinemaHall) error
	Delete(ctx context.Context, id uint) error
}
