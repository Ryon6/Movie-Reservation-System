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

// Create 创建一个 booking
func (r *gormBookingRepository) Create(ctx context.Context, booking *booking.Booking) (*booking.Booking, error) {
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

// FindByID 根据 booking id 获取 booking
func (r *gormBookingRepository) FindByID(ctx context.Context, id vo.BookingID) (*booking.Booking, error) {
	logger := r.logger.With(applog.String("Method", "GetBookingByID"),
		applog.Uint("booking_id", uint(id)))

	var bookingGorm models.BookingGorm
	if err := r.db.WithContext(ctx).Preload("BookedSeats").First(&bookingGorm, id).Error; err != nil {
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

// FindByUserID 根据 user id 获取 booking
func (r *gormBookingRepository) FindByUserID(ctx context.Context, userID vo.UserID) ([]*booking.Booking, error) {
	logger := r.logger.With(applog.String("Method", "GetBookingsByUserID"),
		applog.Uint("user_id", uint(userID)))

	var bookingGorms []models.BookingGorm
	if err := r.db.WithContext(ctx).Preload("BookedSeats").Where("user_id = ?", userID).Find(&bookingGorms).Error; err != nil {
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

// FindByShowtimeID 根据 showtime id 获取 booking
func (r *gormBookingRepository) FindByShowtimeID(ctx context.Context, showtimeID vo.ShowtimeID) ([]*booking.Booking, error) {
	logger := r.logger.With(applog.String("Method", "GetBookingsByShowtimeID"),
		applog.Uint("showtime_id", uint(showtimeID)))

	var bookingGorms []models.BookingGorm
	if err := r.db.WithContext(ctx).Preload("BookedSeats").Where("showtime_id = ?", showtimeID).Find(&bookingGorms).Error; err != nil {
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

// List 查询订单
func (r *gormBookingRepository) List(ctx context.Context, options *booking.BookingQueryOptions) ([]*booking.Booking, int64, error) {
	logger := r.logger.With(applog.String("Method", "ListBookings"))

	var bookingGorms []models.BookingGorm
	var totalCount int64
	query := r.db.WithContext(ctx)
	countQuery := r.db.WithContext(ctx)

	// 用户ID过滤
	if options.UserID != 0 {
		query = query.Where("user_id = ?", options.UserID)
		countQuery = countQuery.Where("user_id = ?", options.UserID)
		logger = logger.With(applog.Uint("query_user_id", uint(options.UserID)))
	}

	// 状态过滤
	if options.Status != "" {
		query = query.Where("status = ?", options.Status)
		countQuery = countQuery.Where("status = ?", options.Status)
		logger = logger.With(applog.String("query_status", string(options.Status)))
	}

	// 分页
	offset := (options.Page - 1) * options.PageSize
	query = query.Offset(offset).Limit(options.PageSize)

	if err := countQuery.Count(&totalCount).Error; err != nil {
		logger.Error("database count bookings error", applog.Error(err))
		return nil, 0, fmt.Errorf("database count bookings error: %w", err)
	}

	if totalCount == 0 {
		logger.Info("No bookings found matching criteria")
		return nil, 0, nil // 返回空列表和0计数
	}

	if err := query.Find(&bookingGorms).Error; err != nil {
		logger.Error("database list bookings error", applog.Error(err))
		return nil, 0, fmt.Errorf("database list bookings error: %w", err)
	}

	bks := make([]*booking.Booking, len(bookingGorms))
	for i, bookingGorm := range bookingGorms {
		bks[i] = bookingGorm.ToDomain()
	}

	logger.Info("list bookings successfully")
	return bks, totalCount, nil
}

// Update 更新 booking
func (r *gormBookingRepository) Update(ctx context.Context, bk *booking.Booking) error {
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

// Delete 删除 booking
func (r *gormBookingRepository) Delete(ctx context.Context, id vo.BookingID) error {
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
