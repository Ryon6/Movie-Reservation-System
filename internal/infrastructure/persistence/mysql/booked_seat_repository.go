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

type gormBookedSeatRepository struct {
	db     *gorm.DB
	logger applog.Logger
}

func NewGormBookedSeatRepository(db *gorm.DB, logger applog.Logger) booking.BookedSeatRepository {
	return &gormBookedSeatRepository{db: db, logger: logger.With(applog.String("Repository", "gormBookedSeatRepository"))}
}

// CreateBookedSeats 创建已预订的座位
func (r *gormBookedSeatRepository) CreateBookedSeats(ctx context.Context, bookedSeats []*booking.BookedSeat) ([]*booking.BookedSeat, error) {
	logger := r.logger.With(applog.String("Method", "CreateBookedSeats"))

	bookedSeatGorms := make([]*models.BookedSeatGorm, len(bookedSeats))
	for i, bookedSeat := range bookedSeats {
		bookedSeatGorms[i] = models.BookedSeatGormFromDomain(bookedSeat)
	}

	if err := r.db.WithContext(ctx).Create(&bookedSeatGorms).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			logger.Warn("booked seats already locked", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", booking.ErrBookedSeatAlreadyLocked, err)
		}
		logger.Error("database create booked seats error", applog.Error(err))
		return nil, fmt.Errorf("database create booked seats error: %w", err)
	}

	domainBookedSeats := make([]*booking.BookedSeat, len(bookedSeatGorms))
	for i, bookedSeatGorm := range bookedSeatGorms {
		domainBookedSeats[i] = bookedSeatGorm.ToDomain()
	}

	logger.Info("create booked seats successfully")
	return domainBookedSeats, nil
}

// GetBookedSeatByID 根据ID获取已预订的座位
func (r *gormBookedSeatRepository) GetBookedSeatByID(ctx context.Context, id vo.BookedSeatID) (*booking.BookedSeat, error) {
	logger := r.logger.With(applog.String("Method", "GetBookedSeatByID"), applog.Uint("booked_seat_id", uint(id)))

	var bookedSeatGorm models.BookedSeatGorm
	if err := r.db.WithContext(ctx).First(&bookedSeatGorm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("booked seat not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %v", booking.ErrBookedSeatNotFound, id)
		}
		logger.Error("database get booked seat by id error", applog.Error(err))
		return nil, fmt.Errorf("database get booked seat by id error: %w", err)
	}

	logger.Info("get booked seat by id successfully")
	return bookedSeatGorm.ToDomain(), nil
}

// GetBookedSeatsByBookingID 根据bookingID获取已预订的座位
func (r *gormBookedSeatRepository) GetBookedSeatsByBookingID(ctx context.Context, bookingID vo.BookingID) ([]*booking.BookedSeat, error) {
	logger := r.logger.With(applog.String("Method", "GetBookedSeatsByBookingID"), applog.Uint("booking_id", uint(bookingID)))

	var bookedSeatGorms []models.BookedSeatGorm
	if err := r.db.WithContext(ctx).Where("booking_id = ?", bookingID).Find(&bookedSeatGorms).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("booked seats not found by booking id", applog.Error(err))
			return nil, fmt.Errorf("%w(booking_id): %v", booking.ErrBookedSeatNotFound, bookingID)
		}
		logger.Error("database get booked seats by booking id error", applog.Error(err))
		return nil, fmt.Errorf("database get booked seats by booking id error: %w", err)
	}

	domainBookedSeats := make([]*booking.BookedSeat, len(bookedSeatGorms))
	for i, bookedSeatGorm := range bookedSeatGorms {
		domainBookedSeats[i] = bookedSeatGorm.ToDomain()
	}

	logger.Info("get booked seats by booking id successfully")
	return domainBookedSeats, nil
}

// UpdateBookedSeat 更新已预订的座位
func (r *gormBookedSeatRepository) UpdateBookedSeat(ctx context.Context, bookedSeat *booking.BookedSeat) error {
	logger := r.logger.With(applog.String("Method", "UpdateBookedSeat"))

	bookedSeatGorm := models.BookedSeatGormFromDomain(bookedSeat)

	result := r.db.WithContext(ctx).Model(&models.BookedSeatGorm{}).
		Where("id = ?", bookedSeatGorm.ID).
		Updates(bookedSeatGorm)

	if err := result.Error; err != nil {
		logger.Error("database update booked seat error", applog.Error(err))
		return fmt.Errorf("database update booked seat error: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("booked seat not found", applog.Error(booking.ErrBookedSeatNotFound))
		return fmt.Errorf("%w", booking.ErrBookedSeatNotFound)
	}

	logger.Info("update booked seat successfully")
	return nil
}

// DeleteBookedSeat 删除已预订的座位
func (r *gormBookedSeatRepository) DeleteBookedSeat(ctx context.Context, id vo.BookedSeatID) error {
	logger := r.logger.With(applog.String("Method", "DeleteBookedSeat"), applog.Uint("booked_seat_id", uint(id)))

	result := r.db.WithContext(ctx).Delete(&models.BookedSeatGorm{}, id)
	if err := result.Error; err != nil {
		logger.Error("database delete booked seat error", applog.Error(err))
		return fmt.Errorf("database delete booked seat error: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("booked seat not found", applog.Error(booking.ErrBookedSeatNotFound))
		return fmt.Errorf("%w", booking.ErrBookedSeatNotFound)
	}

	logger.Info("delete booked seat successfully")
	return nil
}
