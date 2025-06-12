package models

import (
	"mrs/internal/domain/booking"
	"mrs/internal/domain/shared/vo"

	"gorm.io/gorm"
)

type BookedSeatGorm struct {
	gorm.Model
	BookingID uint        `gorm:"not null;index;foreignKey:BookingID,references:ID"`
	SeatID    uint        `gorm:"not null;index;foreignKey:SeatID,references:ID"`
	Price     float64     `gorm:"not null"`
	Booking   BookingGorm `gorm:"foreignKey:BookingID"`
	Seat      SeatGorm    `gorm:"foreignKey:SeatID"`
}

func (BookedSeatGorm) TableName() string {
	return "booked_seats"
}

func (b *BookedSeatGorm) ToDomain() *booking.BookedSeat {
	return &booking.BookedSeat{
		ID:        vo.BookedSeatID(b.ID),
		BookingID: vo.BookingID(b.BookingID),
		SeatID:    vo.SeatID(b.SeatID),
		Price:     b.Price,
	}
}

func BookedSeatGormFromDomain(b *booking.BookedSeat) *BookedSeatGorm {
	return &BookedSeatGorm{
		Model:     gorm.Model{ID: uint(b.ID)},
		BookingID: uint(b.BookingID),
		SeatID:    uint(b.SeatID),
		Price:     b.Price,
	}
}
