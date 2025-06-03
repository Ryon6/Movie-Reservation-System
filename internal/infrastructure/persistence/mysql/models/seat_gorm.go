package models

import (
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared/vo"

	"gorm.io/gorm"
)

type SeatGrom struct {
	gorm.Model
	// 对 CinemaHallID 单独建立索引: 应对高频查询影厅全部座位的需求
	CinemaHallID uint           `gorm:"not null;index;uniqueIndex:idx_hall_row_number"` // 联合唯一索引
	CinemaHall   CinemaHallGrom `gorm:"foreignKey:CinemaHallID"`
	Row          string         `gorm:"type:varchar(10);not null;uniqueIndex:idx_hall_row_number"` // 联合唯一索引
	RowNumber    string         `gorm:"type:varchar(10);not null;uniqueIndex:idx_hall_row_number"`
	Type         string         `gorm:"type:varchar(50);default:'STANDARD'"`
}

func (s *SeatGrom) ToDomain() *cinema.Seat {
	return &cinema.Seat{
		ID:         vo.SeatID(s.ID),
		CinemaHall: s.CinemaHall.ToDomain(),
		Row:        s.Row,
		RowNumber:  s.RowNumber,
		Type:       cinema.SeatType(s.Type),
	}
}

func SeatGromFromDomain(s *cinema.Seat) *SeatGrom {
	return &SeatGrom{
		Model:        gorm.Model{ID: uint(s.ID)},
		CinemaHallID: uint(s.CinemaHall.ID),
		CinemaHall:   *CinemaHallGromFromDomain(s.CinemaHall),
		Row:          s.Row,
		RowNumber:    s.RowNumber,
		Type:         string(s.Type),
	}
}
