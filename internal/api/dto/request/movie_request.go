package request

import (
	"mrs/internal/domain/movie"
	"mrs/internal/domain/shared/vo"
	"time"
)

// CreateMovieRequest 定义了创建电影请求的结构体。
type CreateMovieRequest struct {
	Title           string    `json:"title" binding:"required,min=1,max=255"`
	GenreNames      []string  `json:"genre_names" binding:"required,min=1,max=255"`
	Description     string    `json:"description" binding:"omitempty,min=1,max=1000"`
	ReleaseDate     time.Time `json:"release_date" binding:"required"`
	DurationMinutes int       `json:"duration_minutes" binding:"omitempty,min=1"`
	Rating          float64   `json:"rating" binding:"omitempty,min=0,max=10"`
	PosterURL       string    `json:"poster_url" binding:"omitempty,url"`
	AgeRating       string    `json:"age_rating" binding:"omitempty,min=1,max=50"`
	Cast            string    `json:"cast" binding:"omitempty,min=1,max=1000"`
}

// Genre 需要应用层自行处理
func (r *CreateMovieRequest) ToMovie() *movie.Movie {
	return &movie.Movie{
		Title:           r.Title,
		Description:     r.Description,
		ReleaseDate:     r.ReleaseDate,
		DurationMinutes: r.DurationMinutes,
		Rating:          float32(r.Rating),
		PosterURL:       r.PosterURL,
		AgeRating:       r.AgeRating,
		Cast:            r.Cast,
	}
}

// GetMovieRequest 定义了获取电影请求的结构体。
type GetMovieRequest struct {
	ID uint
}

// UpdateMovieRequest 定义了更新电影请求的结构体。
type UpdateMovieRequest struct {
	ID              uint
	Title           string    `json:"title" binding:"omitempty,min=1,max=255"`
	GenreNames      []string  `json:"genre_names" binding:"omitempty,min=1,max=255"`
	Description     string    `json:"description" binding:"omitempty,min=1,max=1000"`
	ReleaseDate     time.Time `json:"release_date" binding:"omitempty"`
	DurationMinutes int       `json:"duration_minutes" binding:"omitempty,min=1"`
	Rating          float64   `json:"rating" binding:"omitempty,min=0,max=10"`
	PosterURL       string    `json:"poster_url" binding:"omitempty,url"`
	AgeRating       string    `json:"age_rating" binding:"omitempty,min=1,max=50"`
	Cast            string    `json:"cast" binding:"omitempty,min=1,max=1000"`
}

func (r *UpdateMovieRequest) ToDomain() *movie.Movie {
	return &movie.Movie{
		ID:              vo.MovieID(r.ID),
		Title:           r.Title,
		Description:     r.Description,
		ReleaseDate:     r.ReleaseDate,
		DurationMinutes: r.DurationMinutes,
		Rating:          float32(r.Rating),
		PosterURL:       r.PosterURL,
		AgeRating:       r.AgeRating,
		Cast:            r.Cast,
	}
}

// 删除电影
type DeleteMovieRequest struct {
	ID uint
}

type ListMovieRequest struct {
	PaginationRequest
	Title       string `json:"title" form:"title" binding:"omitempty,min=1,max=255"`
	GenreName   string `json:"genre_name" form:"genre_name" binding:"omitempty,min=1,max=255"`
	ReleaseYear int    `json:"release_year" form:"release_year" binding:"omitempty,min=1900,max=2100"` // 按上映年份过滤
}

func (r *ListMovieRequest) ToDomain() *movie.MovieQueryOptions {
	return &movie.MovieQueryOptions{
		Title:       r.Title,
		GenreName:   r.GenreName,
		ReleaseYear: r.ReleaseYear,
		Page:        r.Page,
		PageSize:    r.PageSize,
	}
}

// 创建类型
type CreateGenreRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

// 更新类型
type UpdateGenreRequest struct {
	ID   uint
	Name string `json:"name" binding:"required,min=1,max=255"`
}

func (r *UpdateGenreRequest) ToDomain() *movie.Genre {
	return &movie.Genre{
		ID:   vo.GenreID(r.ID),
		Name: r.Name,
	}
}

// 删除类型
type DeleteGenreRequest struct {
	ID uint
}
