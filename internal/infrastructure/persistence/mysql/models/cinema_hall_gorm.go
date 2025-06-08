package models

import (
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared/vo"

	"gorm.io/gorm"
)

// 影厅表
type CinemaHallGorm struct {
	gorm.Model
	Name        string `gorm:"type:varchar(50);uniqueIndex;not null;"` // 影厅名称 (例如: "1号厅", "IMAX厅")。
	ScreenType  string `gorm:"type:varchar(50)"`                       // 屏幕类型，如 "2D", "3D", "IMAX"
	SoundSystem string `gorm:"type:varchar(100)"`                      // 音响系统

	Seats []SeatGorm `gorm:"foreignKey:CinemaHallID;OnDelete:CASCADE"`
}

// TableName 指定表名
func (CinemaHallGorm) TableName() string {
	return "cinema_halls"
}

// ToDomain 将CinemaHallGorm转换为CinemaHall（开销过大，需要在业务层处理）
func (c *CinemaHallGorm) ToDomain() *cinema.CinemaHall {
	seats := make([]*cinema.Seat, len(c.Seats))
	if c.Seats != nil {
		for i, seat := range c.Seats {
			seats[i] = seat.ToDomain()
		}
	}
	return &cinema.CinemaHall{
		ID:          vo.CinemaHallID(c.ID),
		Name:        c.Name,
		ScreenType:  c.ScreenType,
		SoundSystem: c.SoundSystem,
		Seats:       seats,
	}
}

// CinemaHallGormFromDomain 将CinemaHall转换为CinemaHallGorm（开销过大，需要在业务层处理）
func CinemaHallGormFromDomain(c *cinema.CinemaHall) *CinemaHallGorm {
	return &CinemaHallGorm{
		Model:       gorm.Model{ID: uint(c.ID)},
		Name:        c.Name,
		ScreenType:  c.ScreenType,
		SoundSystem: c.SoundSystem,
	}
}
