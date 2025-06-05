package request

import (
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared/vo"
)

// 创建影厅
type CreateCinemaHallRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	ScreenType  string `json:"screen_type" binding:"required,min=1,max=255"`
	SoundSystem string `json:"sound_system" binding:"required,min=1,max=255"`
}

func (r *CreateCinemaHallRequest) ToDomain() *cinema.CinemaHall {
	return &cinema.CinemaHall{
		Name:        r.Name,
		ScreenType:  r.ScreenType,
		SoundSystem: r.SoundSystem,
	}
}

type GetCinemaHallRequest struct {
	ID uint `json:"id" binding:"required,min=1"`
}

type SeatRequest struct {
	RowIdentifier string `json:"row_identifier" binding:"required,min=1,max=255"`
	SeatNumber    string `json:"seat_number" binding:"required,min=1,max=255"`
	Type          string `json:"type" binding:"required,min=1,max=255"`
}

func (r *SeatRequest) ToDomain() *cinema.Seat {
	return &cinema.Seat{
		RowIdentifier: r.RowIdentifier,
		SeatNumber:    r.SeatNumber,
		Type:          cinema.SeatType(r.Type),
	}
}

type CreateHallSeatsRequest struct {
	CinemaHallID uint           `json:"cinema_hall_id" binding:"required,min=1"`
	Seats        []*SeatRequest `json:"seats" binding:"required,min=1"`
}

func (r *CreateHallSeatsRequest) ToDomain() []*cinema.Seat {
	domainSeats := make([]*cinema.Seat, len(r.Seats))
	for i, seat := range r.Seats {
		domainSeats[i] = seat.ToDomain()
		domainSeats[i].CinemaHallID = vo.CinemaHallID(r.CinemaHallID)
	}
	return domainSeats
}
