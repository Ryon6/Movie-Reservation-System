package models

import "gorm.io/gorm"

type GenreGrom struct {
	gorm.Model
	Name   string       `gorm:"type:varchar(100);uniqueIndex;not null"`              // 类型名称：科幻...
	Movies []*MovieGrom `gorm:"many2many:movie_genres;constraint:OnDelete:RESTRICT"` // 该类型下的电影 (可选，如果需要反向查询)
}
