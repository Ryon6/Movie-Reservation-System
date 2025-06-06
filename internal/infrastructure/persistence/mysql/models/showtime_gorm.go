package models

import (
	"mrs/internal/domain/shared/vo"
	"mrs/internal/domain/showtime"
	"time"

	"gorm.io/gorm"
)

type ShowtimeGrom struct {
	gorm.Model
	// 关联的电影ID 已在Movie中定义外键 电影ID，联合索引1
	MovieID uint `gorm:"not null;index;index:idx_movie_start_time,priority:1"`
	// 关联的电影实体
	Movie MovieGrom `gorm:"foreignKey:MovieID"`
	// 关联的影厅ID 已在CinemaHall中定义外键 影厅ID，联合索引2
	CinemaHallID uint `gorm:"not null;index;index:idx_hall_start_time,priority:1"`
	// 关联的影厅实体
	CinemaHall CinemaHallGrom `gorm:"foreignKey:CinemaHallID"`

	// 放映开始时间 (包含日期和时间) 放映开始时间，联合索引1和2
	StartTime time.Time `gorm:"not null;index;index:idx_movie_start_time,priority:2;index:idx_hall_start_time,priority:2"`
	// 放映结束时间 (可以根据电影时长和开始时间计算)
	EndTime time.Time `gorm:"not null"`
	// 该场次的票价 (可以更复杂，比如不同座位类型不同价格)
	Price float64 `gorm:"not null"`
}

func (s *ShowtimeGrom) ToDomain() *showtime.Showtime {
	return &showtime.Showtime{
		ID:           vo.ShowtimeID(s.ID),
		MovieID:      vo.MovieID(s.MovieID),
		CinemaHallID: vo.CinemaHallID(s.CinemaHallID),
		Movie:        s.Movie.ToDomain(),
		CinemaHall:   s.CinemaHall.ToDomain(),
		StartTime:    s.StartTime,
		EndTime:      s.EndTime,
		Price:        s.Price,
	}
}

func ShowtimeGromFromDomain(s *showtime.Showtime) *ShowtimeGrom {
	return &ShowtimeGrom{
		Model:        gorm.Model{ID: uint(s.ID)},
		MovieID:      uint(s.MovieID),
		CinemaHallID: uint(s.CinemaHallID),
		StartTime:    s.StartTime,
		EndTime:      s.EndTime,
		Price:        s.Price,
	}
}
