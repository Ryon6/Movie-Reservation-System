package models

import (
	"mrs/internal/domain/movie"
	"mrs/internal/domain/shared/vo"

	"gorm.io/gorm"
)

// 类型表
type GenreGorm struct {
	gorm.Model
	Name string `gorm:"type:varchar(100);uniqueIndex;not null"` // 类型名称：科幻...
	// 只需一方显式定义多对多关系，另一方会自动创建连接表
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
