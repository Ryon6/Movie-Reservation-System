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

// 影厅简易信息
type CinemaHallSimpleResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	ScreenType  string `json:"screen_type"`
	SoundSystem string `json:"sound_system"`
}

func ToCinemaHallSimpleResponse(hall *cinema.CinemaHall) *CinemaHallSimpleResponse {
	if hall == nil {
		return nil
	}
	return &CinemaHallSimpleResponse{
		ID:          uint(hall.ID),
		Name:        hall.Name,
		ScreenType:  hall.ScreenType,
		SoundSystem: hall.SoundSystem,
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
	if seat == nil {
		return nil
	}
	return &SeatResponse{
		ID:            uint(seat.ID),
		RowIdentifier: seat.RowIdentifier,
		SeatNumber:    seat.SeatNumber,
		Type:          string(seat.Type),
	}
}

type ListAllCinemaHallsResponse struct {
	CinemaHalls []*CinemaHallSimpleResponse `json:"cinema_halls"`
}

func ToListAllCinemaHallsResponse(halls []*cinema.CinemaHall) *ListAllCinemaHallsResponse {
	if halls == nil {
		return nil
	}
	cinemaHallResponses := make([]*CinemaHallSimpleResponse, len(halls))
	for i, hall := range halls {
		cinemaHallResponses[i] = ToCinemaHallSimpleResponse(hall)
	}
	return &ListAllCinemaHallsResponse{
		CinemaHalls: cinemaHallResponses,
	}
}
