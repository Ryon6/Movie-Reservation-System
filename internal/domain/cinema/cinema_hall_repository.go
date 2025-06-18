package cinema

import (
	"context"
	"mrs/internal/domain/shared/vo"
)

type CinemaHallRepository interface {
	Create(ctx context.Context, hall *CinemaHall) (*CinemaHall, error)
	FindByID(ctx context.Context, id vo.CinemaHallID) (*CinemaHall, error)
	FindByName(ctx context.Context, name string) (*CinemaHall, error)
	ListAll(ctx context.Context) ([]*CinemaHall, error)
	Update(ctx context.Context, hall *CinemaHall) error
	Delete(ctx context.Context, id vo.CinemaHallID) error
}
