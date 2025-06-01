package mysql

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/movie"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type gormSeatRepository struct {
	logger applog.Logger
	db     *gorm.DB
}

func NewGormSeatRepository(db *gorm.DB, logger applog.Logger) movie.SeatRepository {
	return &gormSeatRepository{
		logger: logger,
		db:     db,
	}
}

func (r *gormSeatRepository) CreateBatch(ctx context.Context, seats []*movie.Seat) error {
	logger := r.logger.With(applog.String("Method", "gormSeatRepository.CreateBatch"), applog.Int("seat count", len(seats)))

	if err := r.db.WithContext(ctx).Create(&seats).Error; err != nil {
		logger.Error("failed to create seats", applog.Error(err))
		return err
	}

	logger.Info("create seats successfully")
	return nil
}

func (r *gormSeatRepository) FindByID(ctx context.Context, id uint) (*movie.Seat, error) {
	logger := r.logger.With(applog.String("Method", "gormSeatRepository.FindByID"), applog.Uint("seat_id", id))
	var seat movie.Seat
	if err := r.db.WithContext(ctx).First(&seat, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("seat id not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %w", movie.ErrSeatNotFound, err)
		}
		logger.Error("failed to find seat by id", applog.Error(err))
		return nil, fmt.Errorf("failed to find seat by id: %w", err)
	}

	logger.Info("find seat by id successfully")
	return &seat, nil
}

func (r *gormSeatRepository) FindByHallID(ctx context.Context, hallID uint) ([]*movie.Seat, error) {
	logger := r.logger.With(applog.String("Method", "gormSeatRepository.FindByHallID"), applog.Uint("hall_id", hallID))
	var seats []*movie.Seat

	if err := r.db.WithContext(ctx).Where("cinema_hall_id = ?", hallID).Find(&seats).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("no seats found for hall", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", movie.ErrSeatNotFound, err)
		}
		logger.Error("failed to find seats by hall id", applog.Error(err))
		return nil, fmt.Errorf("failed to find seats by hall id: %w", err)
	}

	logger.Info("find seats by hall id successfully")
	return seats, nil
}

func (r *gormSeatRepository) GetSeatsByIDs(ctx context.Context, seatIDs []uint) ([]*movie.Seat, error) {
	logger := r.logger.With(applog.String("Method", "gormSeatRepository.GetSeatsByIDs"), applog.Int("seat count", len(seatIDs)))
	var seats []*movie.Seat
	if err := r.db.WithContext(ctx).Where("id IN ?", seatIDs).Find(&seats).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("seats not found", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", movie.ErrSeatNotFound, err)
		}
		logger.Error("failed to get seats by ids", applog.Error(err))
		return nil, fmt.Errorf("failed to get seats by ids: %w", err)
	}

	logger.Info("get seats by ids successfully")
	return seats, nil
}

func (r *gormSeatRepository) Update(ctx context.Context, seat *movie.Seat) error {
	logger := r.logger.With(applog.String("Method", "gormSeatRepository.Update"),
		applog.Uint("seat_id", seat.ID), applog.Uint("hall_id", seat.CinemaHallID))

	if err := r.db.WithContext(ctx).First(&movie.Seat{}, seat.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("seat not found", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrSeatNotFound, err)
		}
		logger.Error("failed to find seat by id", applog.Error(err))
		return fmt.Errorf("failed to find seat by id: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", seat.ID).Updates(&seat)
	if err := result.Error; err != nil {
		logger.Error("failed to update seat", applog.Error(err))
		return fmt.Errorf("failed to update seat: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("no rows affected during update")
		return movie.ErrNoRowsAffected
	}

	logger.Info("update seat successfully")
	return nil
}

func (r *gormSeatRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "gormSeatRepository.Delete"), applog.Uint("seat_id", id))

	result := r.db.WithContext(ctx).Delete(&movie.Seat{}, id)
	if err := result.Error; err != nil {
		logger.Error("failed to delete seat", applog.Error(err))
		return fmt.Errorf("failed to delete seat: %w", err)
	}
	// 是否不存在或已删除
	if result.RowsAffected == 0 {
		logger.Warn("seat to delete not found or already deleted")
		return movie.ErrSeatNotFound
	}

	logger.Info("delete seat successfully")
	return nil
}

func (r *gormSeatRepository) DeleteByHallID(ctx context.Context, hallID uint) error {
	logger := r.logger.With(applog.String("Method", "gormSeatRepository.DeleteByHallID"), applog.Uint("hall_id", hallID))

	result := r.db.WithContext(ctx).Where("cinema_hall_id = ?", hallID).Delete(&movie.Seat{})
	if err := result.Error; err != nil {
		logger.Error("failed to delete seats", applog.Error(err))
		return fmt.Errorf("failed to delete seats: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("no seats found to delete")
		return movie.ErrSeatNotFound
	}

	logger.Info("delete seats by hall id successfully")
	return nil
}
