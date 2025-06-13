package booking

import "errors"

var (
	ErrBookingNotFound         = errors.New("booking not found")
	ErrBookingAlreadyExists    = errors.New("booking already exists")
	ErrBookedSeatAlreadyLocked = errors.New("booked seat already locked")
	ErrBookedSeatNotFound      = errors.New("booked seat not found")
)
