package models

import (
	"mrs/internal/domain/booking"
	"mrs/internal/domain/shared/vo"

	"gorm.io/gorm"
)

type BookedSeatGorm struct {
	gorm.Model
	BookingID uint `gorm:"not null;index;foreignKey:BookingID,references:ID"`
	// 联合唯一索引(唯一约束保证每个座位只能被预订一次)
	ShowtimeID uint         `gorm:"not null;uniqueIndex:idx_showtime_seat;foreignKey:ShowtimeID,references:ID"`
	SeatID     uint         `gorm:"not null;uniqueIndex:idx_showtime_seat;foreignKey:SeatID,references:ID"`
	Price      float64      `gorm:"not null"`
	Booking    BookingGorm  `gorm:"foreignKey:BookingID"`
	Showtime   ShowtimeGorm `gorm:"foreignKey:ShowtimeID"`
	Seat       SeatGorm     `gorm:"foreignKey:SeatID"`
}

func (BookedSeatGorm) TableName() string {
	return "booked_seats"
}

func (b *BookedSeatGorm) ToDomain() *booking.BookedSeat {
	return &booking.BookedSeat{
		ID:         vo.BookedSeatID(b.ID),
		BookingID:  vo.BookingID(b.BookingID),
		ShowtimeID: vo.ShowtimeID(b.ShowtimeID),
		SeatID:     vo.SeatID(b.SeatID),
		Price:      b.Price,
	}
}

func BookedSeatGormFromDomain(b *booking.BookedSeat) *BookedSeatGorm {
	return &BookedSeatGorm{
		Model:      gorm.Model{ID: uint(b.ID)},
		BookingID:  uint(b.BookingID),
		ShowtimeID: uint(b.ShowtimeID),
		SeatID:     uint(b.SeatID),
		Price:      b.Price,
	}
}
