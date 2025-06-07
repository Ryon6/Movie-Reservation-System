package models

import (
	"mrs/internal/domain/movie"
	"mrs/internal/domain/shared/vo"

	"gorm.io/gorm"
)

type GenreGrom struct {
	gorm.Model
	Name string `gorm:"type:varchar(100);uniqueIndex;not null"` // 类型名称：科幻...
	// 只需一方显式定义多对多关系，另一方会自动创建连接表
}

func (g *GenreGrom) ToDomain() *movie.Genre {
	return &movie.Genre{
		ID:   vo.GenreID(g.ID),
		Name: g.Name,
	}
}

func GenreGromFromDomain(g *movie.Genre) *GenreGrom {
	return &GenreGrom{
		Model: gorm.Model{ID: uint(g.ID)},
		Name:  g.Name,
	}
}
