package showtime

import "errors"

// Showtime 相关错误
var (
	ErrShowtimeNotFound         = errors.New("showtime not found")
	ErrShowtimeOverlap          = errors.New("showtime overlaps with existing showtimes")
	ErrShowtimeInPast           = errors.New("showtime cannot be scheduled in the past")
	ErrShowtimeInvalidTimeRange = errors.New("invalid showtime start/end time range")
	ErrShowtimeNoSeatsAvailable = errors.New("no seats available for this showtime")
	ErrShowtimeEnded            = errors.New("showtime has ended")
)
