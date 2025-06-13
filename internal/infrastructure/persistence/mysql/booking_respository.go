package mysql

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/booking"
	"mrs/internal/domain/shared/vo"
	"mrs/internal/infrastructure/persistence/mysql/models"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type gormBookingRepository struct {
	db     *gorm.DB
	logger applog.Logger
}

func NewGormBookingRepository(db *gorm.DB, logger applog.Logger) booking.BookingRepository {
	return &gormBookingRepository{db: db, logger: logger.With(applog.String("Repository", "gormBookingRepository"))}
}

// CreateBooking 创建一个 booking
func (r *gormBookingRepository) CreateBooking(ctx context.Context, booking *booking.Booking) (*booking.Booking, error) {
	logger := r.logger.With(applog.String("Method", "CreateBooking"),
		applog.Uint("booking_id", uint(booking.ID)))

	bookingGorm := models.BookingGormFromDomain(booking)
	if err := r.db.WithContext(ctx).Create(bookingGorm).Error; err != nil {
		logger.Error("database create booking error", applog.Error(err))
		return nil, fmt.Errorf("database create booking error: %w", err)
	}

	logger.Info("create booking successfully")
	return bookingGorm.ToDomain(), nil
}

// GetBookingByID 根据 booking id 获取 booking
func (r *gormBookingRepository) GetBookingByID(ctx context.Context, id vo.BookingID) (*booking.Booking, error) {
	logger := r.logger.With(applog.String("Method", "GetBookingByID"),
		applog.Uint("booking_id", uint(id)))

	var bookingGorm models.BookingGorm
	if err := r.db.WithContext(ctx).First(&bookingGorm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("booking id not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %w", booking.ErrBookingNotFound, err)
		}
		logger.Error("database get booking by id error", applog.Error(err))
		return nil, fmt.Errorf("database get booking by id error: %w", err)
	}

	logger.Info("get booking by id successfully")
	return bookingGorm.ToDomain(), nil
}

// GetBookingsByUserID 根据 user id 获取 booking
func (r *gormBookingRepository) GetBookingsByUserID(ctx context.Context, userID vo.UserID) ([]*booking.Booking, error) {
	logger := r.logger.With(applog.String("Method", "GetBookingsByUserID"),
		applog.Uint("user_id", uint(userID)))

	var bookingGorms []models.BookingGorm
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&bookingGorms).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("bookings not found", applog.Error(err))
			return nil, fmt.Errorf("%w(user_id): %w", booking.ErrBookingNotFound, err)
		}
		logger.Error("database get bookings by user id error", applog.Error(err))
		return nil, fmt.Errorf("database get bookings by user id error: %w", err)
	}

	bks := make([]*booking.Booking, len(bookingGorms))
	for i, bookingGorm := range bookingGorms {
		bks[i] = bookingGorm.ToDomain()
	}

	logger.Info("get bookings by user id successfully")
	return bks, nil
}

// GetBookingsByShowtimeID 根据 showtime id 获取 booking
func (r *gormBookingRepository) GetBookingsByShowtimeID(ctx context.Context, showtimeID vo.ShowtimeID) ([]*booking.Booking, error) {
	logger := r.logger.With(applog.String("Method", "GetBookingsByShowtimeID"),
		applog.Uint("showtime_id", uint(showtimeID)))

	var bookingGorms []models.BookingGorm
	if err := r.db.WithContext(ctx).Where("showtime_id = ?", showtimeID).Find(&bookingGorms).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("bookings not found", applog.Error(err))
			return nil, fmt.Errorf("%w(showtime_id): %w", booking.ErrBookingNotFound, err)
		}
		logger.Error("database get bookings by showtime id error", applog.Error(err))
		return nil, fmt.Errorf("database get bookings by showtime id error: %w", err)
	}

	bks := make([]*booking.Booking, len(bookingGorms))
	for i, bookingGorm := range bookingGorms {
		bks[i] = bookingGorm.ToDomain()
	}

	logger.Info("get bookings by showtime id successfully")
	return bks, nil
}

// UpdateBooking 更新 booking
func (r *gormBookingRepository) UpdateBooking(ctx context.Context, bk *booking.Booking) error {
	logger := r.logger.With(applog.String("Method", "UpdateBooking"),
		applog.Uint("booking_id", uint(bk.ID)))

	bookingGorm := models.BookingGormFromDomain(bk)
	result := r.db.WithContext(ctx).Model(&models.BookingGorm{}).Where("id = ?", bk.ID).Updates(bookingGorm)
	if result.Error != nil {
		logger.Error("database update booking error", applog.Error(result.Error))
		return fmt.Errorf("database update booking error: %w", result.Error)
	}

	// 未更新任何数据，说明 booking id 不存在
	if result.RowsAffected == 0 {
		logger.Warn("booking id not found", applog.Error(booking.ErrBookingNotFound))
		return fmt.Errorf("%w(id): %v", booking.ErrBookingNotFound, bk.ID)
	}

	logger.Info("update booking successfully")
	return nil
}

// DeleteBooking 删除 booking
func (r *gormBookingRepository) DeleteBooking(ctx context.Context, id vo.BookingID) error {
	logger := r.logger.With(applog.String("Method", "DeleteBooking"),
		applog.Uint("booking_id", uint(id)))

	result := r.db.WithContext(ctx).Delete(&models.BookingGorm{}, id)
	if result.Error != nil {
		logger.Error("database delete booking error", applog.Error(result.Error))
		return fmt.Errorf("database delete booking error: %w", result.Error)
	}

	// 未删除任何数据，说明 booking id 不存在
	if result.RowsAffected == 0 {
		logger.Warn("booking id not found", applog.Error(booking.ErrBookingNotFound))
		return fmt.Errorf("%w(id): %v", booking.ErrBookingNotFound, id)
	}

	logger.Info("delete booking successfully")
	return nil
}
