package booking

type BookingRepository interface {
	CreateBooking(booking *Booking) (*Booking, error)
	GetBookingByID(id uint) (*Booking, error)
	GetBookingsByUserID(userID uint) ([]*Booking, error)
	GetBookingsByShowtimeID(showtimeID uint) ([]*Booking, error)
	UpdateBooking(booking *Booking) (*Booking, error)
	DeleteBooking(id uint) error
}
