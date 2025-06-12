package response

import (
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

type SeatInfo struct {
	ID            uint   `json:"id"`
	SeatType      string `json:"seat_type"`
	RowIdentifier string `json:"row_identifier"` // 座位所在排的标识,如 A、B、C
	SeatNumber    string `json:"seat_number"`    // 座位在该排中的编号,如 1、2、3
	Status        int    `json:"status"`         // 座位状态,0: 可用,1: 已预订,2: 已售出
}

// 座位表
type SeatMapResponse struct {
	ShowtimeID uint       `json:"showtime_id"`
	HallName   string     `json:"hall_name"`
	MovieTitle string     `json:"movie_title"`
	Seats      []SeatInfo `json:"seats"`
}
