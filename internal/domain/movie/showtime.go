package movie

import (
	"time"

	"gorm.io/gorm"
)

type Showtime struct {
	gorm.Model
	MovieID      uint       `gorm:"not null;index"`          // 关联的电影ID 已在Movie中定义外键
	Movie        Movie      `gorm:"foreignKey:MovieID"`      // 关联的电影实体
	CinemaHallID uint       `gorm:"not null;index"`          // 关联的影厅ID 已在CinemaHall中定义外键
	CinemaHall   CinemaHall `gorm:"foreignKey:CinemaHallID"` // 关联的影厅实体

	StartTime time.Time `gorm:"not null;index"` // 放映开始时间 (包含日期和时间)
	EndTime   time.Time `gorm:"not null"`       // 放映结束时间 (可以根据电影时长和开始时间计算)
	Price     float64   `gorm:"not null"`       // 该场次的票价 (可以更复杂，比如不同座位类型不同价格)
}
