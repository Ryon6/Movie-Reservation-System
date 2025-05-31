package movie

import (
	"time"

	"gorm.io/gorm"
)

type Movie struct {
	gorm.Model
	Title           string `gorm:"type:varchar(255);not null;"` //  电影标题
	Description     string `gorm:"type:text"`                   // 电影剧情简介或描述（可空）
	PosterURL       string `gorm:"type:varchar(500)"`           // 电影海报图片的URL地址（可空）
	DurationMinutes int    // 电影时长，单位为分钟

	// 关系
	Genres    []*Genre   `gorm:"many2many:movies_genres"` // 多对多：GORM会自动创建名为movie_genres的连接表。
	Showtimes []Showtime `gorm:"foreignKey:MovieID"`      // 一对多

	// 可选
	ReleaseDate time.Time // 上映日期（可空）
	Rating      float32   // 评分
	AgeRating   string    `gorm:"type:varchar(50)"` // 年龄分级 (例如 PG-13)
	Cast        string    `gorm:"type:text"`        // 主要演员 (简单起见用文本，复杂系统可设计为关联表)
}
