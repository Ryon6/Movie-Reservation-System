package request

import (
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared/vo"
)

// 创建影厅
type CreateCinemaHallRequest struct {
	Name        string         `json:"name" binding:"required,min=1,max=255"`
	ScreenType  string         `json:"screen_type" binding:"required,min=1,max=255"`
	SoundSystem string         `json:"sound_system" binding:"required,min=1,max=255"`
	Seats       []*SeatRequest `json:"seats" binding:"omitempty"` // 影厅座位，如果为空，则自动生成默认座位布局
}

func (r *CreateCinemaHallRequest) ToDomain() *cinema.CinemaHall {
	seats := make([]*cinema.Seat, 0)
	if len(r.Seats) != 0 {
		seats = make([]*cinema.Seat, len(r.Seats))
		for i, seat := range r.Seats {
			seats[i] = seat.ToDomain()
		}
	}
	return &cinema.CinemaHall{
		Name:        r.Name,
		ScreenType:  r.ScreenType,
		SoundSystem: r.SoundSystem,
		Seats:       seats,
	}
}

type GetCinemaHallRequest struct {
	ID uint
}

type SeatRequest struct {
	RowIdentifier string `json:"row_identifier" binding:"required,min=1,max=255"`
	SeatNumber    string `json:"seat_number" binding:"required,min=1,max=255"`
	Type          string `json:"type" binding:"omitempty,min=1,max=255"`
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
	Seats        []*SeatRequest `json:"seats" binding:"required"`
}

func (r *CreateHallSeatsRequest) ToDomain() []*cinema.Seat {
	domainSeats := make([]*cinema.Seat, len(r.Seats))
	for i, seat := range r.Seats {
		domainSeats[i] = seat.ToDomain()
		domainSeats[i].CinemaHallID = vo.CinemaHallID(r.CinemaHallID)
	}
	return domainSeats
}

type UpdateCinemaHallRequest struct {
	ID          uint
	Name        string `json:"name" binding:"omitempty,min=1,max=255"`
	ScreenType  string `json:"screen_type" binding:"omitempty,min=1,max=255"`
	SoundSystem string `json:"sound_system" binding:"omitempty,min=1,max=255"`
}

func (r *UpdateCinemaHallRequest) ToDomain() *cinema.CinemaHall {
	return &cinema.CinemaHall{
		ID:          vo.CinemaHallID(r.ID),
		Name:        r.Name,
		ScreenType:  r.ScreenType,
		SoundSystem: r.SoundSystem,
	}
}

type DeleteCinemaHallRequest struct {
	ID uint
}
