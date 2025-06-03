package movie

import "mrs/internal/domain/shared/vo"

// 电影类型
type Genre struct {
	ID   vo.GenreID // 类型ID
	Name string     // 类型名称

	// 多对多关系
	Movies []*Movie
}
