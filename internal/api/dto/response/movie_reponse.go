package response

import "time"

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
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type MovieSimpleResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"release_date"`
	Rating      float64   `json:"rating"`
	PosterURL   string    `json:"poster_url"`
	AgeRating   string    `json:"age_rating"`
	GenreNames  []string  `json:"genre_names"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PaginatedMovieResponse struct {
	Pagination PaginationResponse     `json:"pagination"`
	Movies     []*MovieSimpleResponse `json:"movies"`
}
