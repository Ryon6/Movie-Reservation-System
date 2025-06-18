package models

import (
	"mrs/internal/domain/movie"
	"mrs/internal/domain/shared/vo"

	"gorm.io/gorm"
)

// 类型表
type GenreGorm struct {
	gorm.Model
	Name   string       `gorm:"type:varchar(100);uniqueIndex;not null"` // 类型名称：科幻...
	Movies []*MovieGorm `gorm:"many2many:movies_genres;joinForeignKey:genre_id;joinReferences:movie_id;constraint:OnDelete:RESTRICT;"`
}

// TableName 指定表名
func (GenreGorm) TableName() string {
	return "genres"
}

func (g *GenreGorm) ToDomain() *movie.Genre {
	return &movie.Genre{
		ID:   vo.GenreID(g.ID),
		Name: g.Name,
	}
}

func GenreGormFromDomain(g *movie.Genre) *GenreGorm {
	return &GenreGorm{
		Model: gorm.Model{ID: uint(g.ID)},
		Name:  g.Name,
	}
}
