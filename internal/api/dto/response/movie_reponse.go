package response

import (
	"mrs/internal/domain/movie"
	"time"
)

// 电影详情
type MovieResponse struct {
	ID              uint             `json:"id"`
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	ReleaseDate     time.Time        `json:"release_date"`
	DurationMinutes int              `json:"duration_minutes"`
	Rating          float64          `json:"rating"`
	PosterURL       string           `json:"poster_url"`
	AgeRating       string           `json:"age_rating"`
	Cast            string           `json:"cast"`
	Genres          []*GenreResponse `json:"genres"`
	// CreatedAt       time.Time        `json:"created_at"`
	// UpdatedAt       time.Time        `json:"updated_at"`
}

func ToMovieResponse(movie *movie.Movie) *MovieResponse {
	genres := make([]*GenreResponse, 0, len(movie.Genres))
	for _, genre := range movie.Genres {
		genres = append(genres, &GenreResponse{
			ID:   uint(genre.ID),
			Name: genre.Name,
		})
	}
	return &MovieResponse{
		ID:              uint(movie.ID),
		Title:           movie.Title,
		Description:     movie.Description,
		ReleaseDate:     movie.ReleaseDate,
		DurationMinutes: movie.DurationMinutes,
		Rating:          float64(movie.Rating),
		PosterURL:       movie.PosterURL,
		AgeRating:       movie.AgeRating,
		Cast:            movie.Cast,
		Genres:          genres,
	}
}

type MovieSimpleResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"release_date"`
	Rating      float64   `json:"rating"`
	PosterURL   string    `json:"poster_url"`
	AgeRating   string    `json:"age_rating"`
	GenreNames  []string  `json:"genre_names"`
	// CreatedAt   time.Time `json:"created_at"`
	// UpdatedAt   time.Time `json:"updated_at"`
}

type PaginatedMovieResponse struct {
	Pagination PaginationResponse     `json:"pagination"`
	Movies     []*MovieSimpleResponse `json:"movies"`
}
