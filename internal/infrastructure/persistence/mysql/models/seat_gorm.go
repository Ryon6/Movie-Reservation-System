package models

import "gorm.io/gorm"

type SeatGrom struct {
	gorm.Model
	// 对 CinemaHallID 单独建立索引: 应对高频查询影厅全部座位的需求
	CinemaHallID uint           `gorm:"not null;index;uniqueIndex:idx_hall_row_number"` // 联合唯一索引
	CinemaHall   CinemaHallGrom `gorm:"foreignKey:CinemaHallID"`
	Row          string         `gorm:"type:varchar(10);not null;uniqueIndex:idx_hall_row_number"` // 联合唯一索引
	RowNumber    string         `gorm:"type:varchar(10);not null;uniqueIndex:idx_hall_row_number"`
	Type         string         `gorm:"type:varchar(50);default:'STANDARD'"`
}
