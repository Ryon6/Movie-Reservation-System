package showtime

import (
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/movie"
	"mrs/internal/domain/shared/vo"
	"time"
)

// 场次
type Showtime struct {
	ID           vo.ShowtimeID      // 场次ID
	MovieID      vo.MovieID         // 电影ID
	CinemaHallID vo.CinemaHallID    // 影厅ID
	Movie        *movie.Movie       // 电影
	CinemaHall   *cinema.CinemaHall // 影厅

	StartTime time.Time // 放映开始时间
	EndTime   time.Time // 放映结束时间
	Price     float64   // 票价
}
