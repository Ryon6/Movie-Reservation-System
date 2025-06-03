package cinema

import "gorm.io/gorm"

type SeatType string

const (
	SeatTypeStandard   = "STANDARD"
	SeatTypeVIP        = "VIP"
	SeatTypeWheelchair = "WHEELCHAIR"
)

// 注意：座位的可用性通常与特定的Showtime相关，而不是座位本身的静态属性。
type Seat struct {
	gorm.Model
	// 对 CinemaHallID 单独建立索引: 应对高频查询影厅全部座位的需求
	CinemaHallID uint       `gorm:"not null;index;uniqueIndex:idx_hall_row_number"` // 联合唯一索引
	CinemaHall   CinemaHall `gorm:"foreignKey:CinemaHallID"`
	Row          string     `gorm:"type:varchar(10);not null;uniqueIndex:idx_hall_row_number"` // 联合唯一索引
	RowNumber    string     `gorm:"type:varchar(10);not null;uniqueIndex:idx_hall_row_number"`
	Type         SeatType   `gorm:"type:varchar(50);default:'STANDARD'"`
}
