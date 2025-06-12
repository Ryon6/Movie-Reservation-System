package response

import (
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/showtime"
	"time"
)

// 返回完整的showtime信息
type ShowtimeResponse struct {
	ID         uint                      `json:"id"`
	Movie      *MovieSimpleResponse      `json:"movie"`
	CinemaHall *CinemaHallSimpleResponse `json:"cinema_hall"`
	StartTime  time.Time                 `json:"start_time"`
	EndTime    time.Time                 `json:"end_time"`
	Price      float64                   `json:"price"`
}

func ToShowtimeResponse(showtime *showtime.Showtime) *ShowtimeResponse {
	return &ShowtimeResponse{
		ID:         uint(showtime.ID),
		Movie:      ToMovieSimpleResponse(showtime.Movie),
		CinemaHall: ToCinemaHallSimpleResponse(showtime.CinemaHall),
		StartTime:  showtime.StartTime,
		EndTime:    showtime.EndTime,
		Price:      showtime.Price,
	}
}

type ShowtimeSimpleResponse struct {
	ID           uint      `json:"id"`
	MovieID      uint      `json:"movie_id"`
	CinemaHallID uint      `json:"cinema_hall_id"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

func ToShowtimeSimpleResponse(showtime *showtime.Showtime) *ShowtimeSimpleResponse {
	return &ShowtimeSimpleResponse{
		ID:           uint(showtime.ID),
		MovieID:      uint(showtime.MovieID),
		CinemaHallID: uint(showtime.CinemaHallID),
		StartTime:    showtime.StartTime,
		EndTime:      showtime.EndTime,
	}
}

type PaginatedShowtimeResponse struct {
	Pagination PaginationResponse        `json:"pagination"`
	Showtimes  []*ShowtimeSimpleResponse `json:"showtimes"`
}

// 座位表
type SeatMapResponse struct {
	Seats []*cinema.SeatInfo `json:"seats"`
}
