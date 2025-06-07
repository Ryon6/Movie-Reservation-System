package movie

import (
	"mrs/internal/domain/shared/vo"
	"time"
)

// 电影实体(必选：Title，类型，上映日期)
type Movie struct {
	ID              vo.MovieID // 电影ID
	Title           string     // 标题
	PosterURL       string     // 海报URL
	DurationMinutes int        // 时长
	ReleaseDate     time.Time  // 上映日期
	Description     string     // 描述
	Rating          float32    // 评分
	AgeRating       string     // 年龄分级 (例如 PG-13)
	Cast            string     // 主要演员 (简单起见用文本，复杂系统可设计为关联表)
	Genres          []*Genre   // 类型（多对多关系）
}
