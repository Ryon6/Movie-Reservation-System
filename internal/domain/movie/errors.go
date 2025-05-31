package movie

import "errors"

var (
	// 通用错误
	ErrInvalidInput = errors.New("invalid input")

	// CinemaHall 相关错误
	ErrCinemaHallNotFound         = errors.New("cinema hall not found")
	ErrCinemaHallAlreadyExists    = errors.New("cinema hall already exists")
	ErrCinemaHallCapacityExceeded = errors.New("cinema hall capacity exceeded")

	// Genre 相关错误
	ErrGenreNotFound      = errors.New("genre not found")
	ErrGenreAlreadyExists = errors.New("genre already exists")

	// Movie 相关错误
	ErrMovieNotFound        = errors.New("movie not found")
	ErrInvalidMovieDuration = errors.New("invalid movie duration")
	ErrInvalidReleaseDate   = errors.New("invalid release date")
	ErrInvalidAgeRating     = errors.New("invalid age rating")

	// Seat 相关错误
	ErrSeatNotFound          = errors.New("seat not found")
	ErrSeatRowNumberConflict = errors.New("seat row/number conflict")
	ErrInvalidSeatType       = errors.New("invalid seat type")
	ErrSeatNotAvailable      = errors.New("seat not available for showtime")

	// Showtime 相关错误
	ErrShowtimeNotFound         = errors.New("showtime not found")
	ErrShowtimeOverlap          = errors.New("showtime overlaps with existing showtimes")
	ErrShowtimeInPast           = errors.New("showtime cannot be scheduled in the past")
	ErrShowtimeInvalidTimeRange = errors.New("invalid showtime start/end time range")
	ErrShowtimeNoSeatsAvailable = errors.New("no seats available for this showtime")

	// 关联错误
	ErrMovieNotAssociatedWithGenre = errors.New("movie not associated with this genre")
	ErrSeatNotInCinemaHall         = errors.New("seat does not belong to this cinema hall")
)
