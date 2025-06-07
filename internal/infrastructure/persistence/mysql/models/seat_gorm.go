package models

import (
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared/vo"

	"gorm.io/gorm"
)

type SeatGorm struct {
	gorm.Model
	// 对 CinemaHallID 单独建立索引: 应对高频查询影厅全部座位的需求
	CinemaHallID  uint           `gorm:"not null;index;uniqueIndex:idx_hall_row_number"` // 联合唯一索引
	CinemaHall    CinemaHallGorm `gorm:"foreignKey:CinemaHallID"`
	RowIdentifier string         `gorm:"type:varchar(10);not null;uniqueIndex:idx_hall_row_number"` // 联合唯一索引
	SeatNumber    string         `gorm:"type:varchar(10);not null;uniqueIndex:idx_hall_row_number"`
	Type          string         `gorm:"type:varchar(50);default:'STANDARD'"`
}

func (s *SeatGorm) ToDomain() *cinema.Seat {
	return &cinema.Seat{
		ID:            vo.SeatID(s.ID),
		CinemaHallID:  vo.CinemaHallID(s.CinemaHallID),
		RowIdentifier: s.RowIdentifier,
		SeatNumber:    s.SeatNumber,
		Type:          cinema.SeatType(s.Type),
	}
}

func SeatGormFromDomain(s *cinema.Seat) *SeatGorm {
	return &SeatGorm{
		Model:         gorm.Model{ID: uint(s.ID)},
		CinemaHallID:  uint(s.CinemaHallID),
		RowIdentifier: s.RowIdentifier,
		SeatNumber:    s.SeatNumber,
		Type:          string(s.Type),
	}
}
