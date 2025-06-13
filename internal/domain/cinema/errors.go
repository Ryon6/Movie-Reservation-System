package cinema

import "errors"

// CinemaHall 相关错误
var (
	ErrCinemaHallNotFound         = errors.New("cinema hall not found")
	ErrCinemaHallAlreadyExists    = errors.New("cinema hall already exists")
	ErrCinemaHallCapacityExceeded = errors.New("cinema hall capacity exceeded")
	ErrCinemaHallReferenced       = errors.New("cinema hall is referenced by other records, cannot delete")
)

// Seat 相关错误
var (
	ErrSeatNotFound          = errors.New("seat not found")
	ErrSeatRowNumberConflict = errors.New("seat row/number conflict")
	ErrInvalidSeatType       = errors.New("invalid seat type")
	ErrSeatNotAvailable      = errors.New("seat not available for showtime")
)
