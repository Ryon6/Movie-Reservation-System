package response

import "mrs/internal/domain/cinema"

// 影厅
type CinemaHallResponse struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	ScreenType  string          `json:"screen_type"`
	SoundSystem string          `json:"sound_system"`
	Seats       []*SeatResponse `json:"seats"`
}

func ToCinemaHallResponse(hall *cinema.CinemaHall) *CinemaHallResponse {
	return &CinemaHallResponse{
		ID:          uint(hall.ID),
		Name:        hall.Name,
		ScreenType:  hall.ScreenType,
		SoundSystem: hall.SoundSystem,
		Seats:       ToSeatResponses(hall.Seats),
	}
}

// 座位
type SeatResponse struct {
	ID            uint   `json:"id"`
	RowIdentifier string `json:"row_identifier"`
	SeatNumber    string `json:"seat_number"`
	Type          string `json:"type"`
}

func ToSeatResponses(seats []*cinema.Seat) []*SeatResponse {
	seatResponses := make([]*SeatResponse, len(seats))
	for i, seat := range seats {
		seatResponses[i] = ToSeatResponse(seat)
	}
	return seatResponses
}

func ToSeatResponse(seat *cinema.Seat) *SeatResponse {
	return &SeatResponse{
		ID:            uint(seat.ID),
		RowIdentifier: seat.RowIdentifier,
		SeatNumber:    seat.SeatNumber,
		Type:          string(seat.Type),
	}
}
