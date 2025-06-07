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
		genres = append(genres, ToGenreResponse(genre))
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
	ID        uint    `json:"id"`
	Title     string  `json:"title"`
	PosterURL string  `json:"poster_url"`
	Rating    float32 `json:"rating"`
	// GenreNames []string `json:"genre_names"`
	// ReleaseDate time.Time `json:"release_date"`
	// AgeRating   string    `json:"age_rating"`
	// CreatedAt   time.Time `json:"created_at"`
	// UpdatedAt   time.Time `json:"updated_at"`
}

func ToMovieSimpleResponse(movie *movie.Movie) *MovieSimpleResponse {
	return &MovieSimpleResponse{
		ID:        uint(movie.ID),
		Title:     movie.Title,
		PosterURL: movie.PosterURL,
		Rating:    movie.Rating,
	}
}

type PaginatedMovieResponse struct {
	Pagination PaginationResponse     `json:"pagination"`
	Movies     []*MovieSimpleResponse `json:"movies"`
}

type GenreResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	// CreatedAt time.Time `json:"created_at"`
	// UpdatedAt time.Time `json:"updated_at"`
}

func ToGenreResponse(genre *movie.Genre) *GenreResponse {
	return &GenreResponse{
		ID:   uint(genre.ID),
		Name: genre.Name,
	}
}

type ListAllGenresResponse struct {
	Genres []*GenreResponse `json:"genres"`
}

func ToListAllGenresResponse(genres []*movie.Genre) *ListAllGenresResponse {
	genreResponses := make([]*GenreResponse, 0, len(genres))
	for _, genre := range genres {
		genreResponses = append(genreResponses, ToGenreResponse(genre))
	}
	return &ListAllGenresResponse{Genres: genreResponses}
}
