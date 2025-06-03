package models

import (
	"mrs/internal/domain/movie"
	"mrs/internal/domain/shared/vo"

	"gorm.io/gorm"
)

type GenreGrom struct {
	gorm.Model
	Name   string       `gorm:"type:varchar(100);uniqueIndex;not null"`              // 类型名称：科幻...
	Movies []*MovieGrom `gorm:"many2many:movie_genres;constraint:OnDelete:RESTRICT"` // 该类型下的电影 (可选，如果需要反向查询)
}

func (g *GenreGrom) ToDomain() *movie.Genre {
	return &movie.Genre{
		ID:   vo.GenreID(g.ID),
		Name: g.Name,
	}
}

func GenreGromFromDomain(g *movie.Genre) *GenreGrom {
	movies := make([]*MovieGrom, len(g.Movies))
	for i, movie := range g.Movies {
		movies[i] = MovieGromFromDomain(movie)
	}
	return &GenreGrom{
		Model:  gorm.Model{ID: uint(g.ID)},
		Name:   g.Name,
		Movies: movies,
	}
}
