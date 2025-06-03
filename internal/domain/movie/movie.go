package movie

import (
	"mrs/internal/domain/shared/vo"
	"time"
)

// 电影
type Movie struct {
	ID              vo.MovieID
	Title           string
	Description     string
	PosterURL       string
	DurationMinutes int

	// 多对多关系
	Genres       []*Genre        // 聚合内部可以直接持有同一聚合内其他实体的引用
	ShowtimesIDs []vo.ShowtimeID // 跨聚合关系优先使用ID

	// 可选
	ReleaseDate time.Time // 上映日期（可空）
	Rating      float32   // 评分
	AgeRating   string    // 年龄分级 (例如 PG-13)
	Cast        string    // 主要演员 (简单起见用文本，复杂系统可设计为关联表)
}
