package booking

import "errors"

var (
	ErrBookingNotFound          = errors.New("booking not found")
	ErrBookingAlreadyExists     = errors.New("booking already exists")
	ErrBookingSeatNotFound      = errors.New("booking seat not found")
	ErrBookingSeatAlreadyLocked = errors.New("booking seat already locked")
)
