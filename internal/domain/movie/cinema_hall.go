package movie

import "gorm.io/gorm"

type CinemaHall struct {
	gorm.Model
	Name        string `gorm:"type:varchar(50);uniqueIndex;not null;"` // 影厅名称 (例如: "1号厅", "IMAX厅")。
	ScreenType  string `gorm:"type:varchar(50)"`                       // 屏幕类型，如 "2D", "3D", "IMAX"
	SoundSystem string `gorm:"type:varchar(100)"`                      // 音响系统
	Capacity    int    // 总座位数

	Seats     []Seat     `gorm:"foreignKey:CinemaHallID;OnDelete:RESTRICT"`
	Showtimes []Showtime `gorm:"foreignKey:CinemaHallID;OnDelete:RESTRICT"`
}
