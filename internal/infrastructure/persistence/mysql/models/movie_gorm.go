package models

import (
	"mrs/internal/domain/movie"
	"mrs/internal/domain/shared/vo"
	"time"

	"gorm.io/gorm"
)

// 电影表(必选：Title，类型，上映日期)
type MovieGorm struct {
	gorm.Model
	Title           string    `gorm:"type:varchar(255);not null;uniqueIndex"` //  电影标题
	ReleaseDate     time.Time `gorm:"not null"`                               // 上映日期
	Description     string    `gorm:"type:text"`                              // 电影剧情简介或描述（可空）
	PosterURL       string    `gorm:"type:varchar(500)"`                      // 电影海报图片的URL地址（可空）
	DurationMinutes int       // 电影时长，单位为分钟
	Rating          float32   // 评分
	AgeRating       string    `gorm:"type:varchar(50)"` // 年龄分级 (例如 PG-13)
	Cast            string    `gorm:"type:text"`        // 主要演员 (简单起见用文本，复杂系统可设计为关联表)

	// 关系
	Genres    []*GenreGorm   `gorm:"many2many:movies_genres;joinForeignKey:movie_id;joinReferences:genre_id;constraint:OnDelete:CASCADE;"` // 多对多：GORM会自动创建名为movies_genres的连接表
	Showtimes []ShowtimeGorm `gorm:"foreignKey:MovieID;constraint:OnDelete:CASCADE;"`                                                      // 一对多
}

// TableName 指定表名
func (MovieGorm) TableName() string {
	return "movies"
}

func (m *MovieGorm) ToDomain() *movie.Movie {
	genres := make([]*movie.Genre, len(m.Genres))
	for i, genre := range m.Genres {
		genres[i] = genre.ToDomain()
	}
	return &movie.Movie{
		ID:              vo.MovieID(m.ID),
		Title:           m.Title,
		Description:     m.Description,
		PosterURL:       m.PosterURL,
		DurationMinutes: m.DurationMinutes,
		Genres:          genres,
		ReleaseDate:     m.ReleaseDate,
		Rating:          m.Rating,
		AgeRating:       m.AgeRating,
		Cast:            m.Cast,
	}
}

// 通常在创建时使用，不需要预加载关联数据
// 添加电影类型需要额外调用ReplaceGenresForMovie
func MovieGormFromDomain(m *movie.Movie) *MovieGorm {
	return &MovieGorm{
		Model:           gorm.Model{ID: uint(m.ID)},
		Title:           m.Title,
		Description:     m.Description,
		PosterURL:       m.PosterURL,
		DurationMinutes: m.DurationMinutes,
		ReleaseDate:     m.ReleaseDate,
		Rating:          m.Rating,
		AgeRating:       m.AgeRating,
		Cast:            m.Cast,
	}
}
