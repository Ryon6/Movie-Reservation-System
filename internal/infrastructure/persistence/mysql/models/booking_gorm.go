package models

import (
	"mrs/internal/domain/booking"
	"mrs/internal/domain/shared/vo"
	"time"

	"gorm.io/gorm"
)

// 订单表
type BookingGorm struct {
	gorm.Model
	UserID      uint             `gorm:"not null;index;foreignKey:UserID,references:ID"`
	ShowtimeID  uint             `gorm:"not null;index;foreignKey:ShowtimeID,references:ID"`
	BookingTime time.Time        `gorm:"not null"`
	TotalAmount float64          `gorm:"not null"`
	Status      string           `gorm:"not null"`
	User        UserGorm         `gorm:"foreignKey:UserID"`
	Showtime    ShowtimeGorm     `gorm:"foreignKey:ShowtimeID"`
	BookedSeats []BookedSeatGorm `gorm:"foreignKey:BookingID"`
}

// TableName 指定表名
func (BookingGorm) TableName() string {
	return "bookings"
}

// ToDomain 将GORM模型转换为领域模型
func (b *BookingGorm) ToDomain() *booking.Booking {
	bookedSeats := make([]*booking.BookedSeat, len(b.BookedSeats))
	for i, bookedSeat := range b.BookedSeats {
		bookedSeats[i] = bookedSeat.ToDomain()
	}
	return &booking.Booking{
		ID:          vo.BookingID(b.ID),
		UserID:      vo.UserID(b.UserID),
		ShowtimeID:  vo.ShowtimeID(b.ShowtimeID),
		TotalAmount: b.TotalAmount,
		BookingTime: b.BookingTime,
		Status:      booking.BookingStatus(b.Status),
		BookedSeats: bookedSeats,
	}
}

// BookingGormFromDomain 将领域模型转换为GORM模型
func BookingGormFromDomain(b *booking.Booking) *BookingGorm {
	return &BookingGorm{
		Model:       gorm.Model{ID: uint(b.ID)},
		UserID:      uint(b.UserID),
		ShowtimeID:  uint(b.ShowtimeID),
		TotalAmount: b.TotalAmount,
		BookingTime: b.BookingTime,
		Status:      string(b.Status),
	}
}
