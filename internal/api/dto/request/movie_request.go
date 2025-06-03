package request

import "time"

// CreateMovieRequest 定义了创建电影请求的结构体。
type CreateMovieRequest struct {
	Title           string    `json:"title" binding:"required,min=1,max=255"`
	GenreNames      []string  `json:"genre_names" binding:"required,min=1,max=255"`
	Description     string    `json:"description" binding:"omitempty,min=1,max=1000"`
	ReleaseDate     time.Time `json:"release_date" binding:"omitempty,datetime=2006-01-02"`
	DurationMinutes int       `json:"duration_minutes" binding:"omitempty,min=1"`
	Rating          float64   `json:"rating" binding:"omitempty,min=0,max=10"`
	PosterURL       string    `json:"poster_url" binding:"omitempty,url"`
	AgeRating       string    `json:"age_rating" binding:"omitempty,min=1,max=50"`
	Cast            string    `json:"cast" binding:"omitempty,min=1,max=1000"`
}

// UpdateMovieRequest 定义了更新电影请求的结构体。
type UpdateMovieRequest struct {
	Title           string    `json:"title" binding:"required,min=1,max=255"`
	GenreNames      []string  `json:"genre_names" binding:"omitempty,min=1,max=255"`
	Description     string    `json:"description" binding:"omitempty,min=1,max=1000"`
	ReleaseDate     time.Time `json:"release_date" binding:"omitempty,datetime=2006-01-02"`
	DurationMinutes int       `json:"duration_minutes" binding:"omitempty,min=1"`
	Rating          float64   `json:"rating" binding:"omitempty,min=0,max=10"`
	PosterURL       string    `json:"poster_url" binding:"omitempty,url"`
	AgeRating       string    `json:"age_rating" binding:"omitempty,min=1,max=50"`
	Cast            string    `json:"cast" binding:"omitempty,min=1,max=1000"`
}

type ListMovieRequest struct {
	PaginationRequest
	Title       string `json:"title" binding:"omitempty,min=1,max=255"`
	GenreName   string `json:"genre_name" binding:"omitempty,min=1,max=255"`
	ReleaseYear int    `json:"release_year" binding:"omitempty,min=1900,max=2100"`                           // 按上映年份过滤
	SortBy      string `json:"sort_by" binding:"omitempty,oneof=title release_date rating duration_minutes"` // 排序字段
	SortOrder   string `json:"sort_order" binding:"omitempty,oneof=asc desc"`                                // 排序顺序
}
