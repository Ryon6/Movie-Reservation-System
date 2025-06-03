package movie

import "errors"

// Genre 相关错误
var (
	ErrGenreNotFound      = errors.New("genre not found")
	ErrGenreAlreadyExists = errors.New("genre already exists")
	ErrGenreReferenced    = errors.New("genre is referenced by other records, cannot delete")
)

// Movie 相关错误
var (
	ErrMovieNotFound        = errors.New("movie not found")
	ErrMovieAlreadyExists   = errors.New("movie already exists")
	ErrInvalidMovieDuration = errors.New("invalid movie duration")
	ErrInvalidReleaseDate   = errors.New("invalid release date")
	ErrInvalidAgeRating     = errors.New("invalid age rating")
)
