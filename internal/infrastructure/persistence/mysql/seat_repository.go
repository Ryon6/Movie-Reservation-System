package mysql

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared"
	"mrs/internal/infrastructure/persistence/mysql/models"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type gormSeatRepository struct {
	logger applog.Logger
	db     *gorm.DB
}

func NewGormSeatRepository(db *gorm.DB, logger applog.Logger) cinema.SeatRepository {
	return &gormSeatRepository{
		logger: logger.With(applog.String("Repository", "gormSeatRepository")),
		db:     db,
	}
}

func (r *gormSeatRepository) CreateBatch(ctx context.Context, seats []*cinema.Seat) ([]*cinema.Seat, error) {
	logger := r.logger.With(applog.String("Method", "CreateBatch"), applog.Int("seat count", len(seats)))

	seatGorms := make([]*models.SeatGrom, len(seats))
	for i, seat := range seats {
		seatGorms[i] = models.SeatGromFromDomain(seat)
	}
	if err := r.db.WithContext(ctx).Create(&seatGorms).Error; err != nil {
		logger.Error("failed to create seats", applog.Error(err))
		return nil, err
	}

	logger.Info("create seats successfully")
	domainSeats := make([]*cinema.Seat, len(seatGorms))
	for i, seatGorm := range seatGorms {
		domainSeats[i] = seatGorm.ToDomain()
	}
	return domainSeats, nil
}

func (r *gormSeatRepository) FindByID(ctx context.Context, id uint) (*cinema.Seat, error) {
	logger := r.logger.With(applog.String("Method", "FindByID"), applog.Uint("seat_id", id))
	var seatGorm models.SeatGrom
	if err := r.db.WithContext(ctx).First(&seatGorm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("seat id not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %w", cinema.ErrSeatNotFound, err)
		}
		logger.Error("failed to find seat by id", applog.Error(err))
		return nil, fmt.Errorf("failed to find seat by id: %w", err)
	}

	logger.Info("find seat by id successfully")
	return seatGorm.ToDomain(), nil
}

func (r *gormSeatRepository) FindByHallID(ctx context.Context, hallID uint) ([]*cinema.Seat, error) {
	logger := r.logger.With(applog.String("Method", "FindByHallID"), applog.Uint("hall_id", hallID))
	var seatsGorms []*models.SeatGrom

	if err := r.db.WithContext(ctx).Where("cinema_hall_id = ?", hallID).Find(&seatsGorms).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("no seats found for hall", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", cinema.ErrCinemaHallNotFound, err)
		}
		logger.Error("failed to find seats by hall id", applog.Error(err))
		return nil, fmt.Errorf("failed to find seats by hall id: %w", err)
	}

	logger.Info("find seats by hall id successfully")
	seats := make([]*cinema.Seat, len(seatsGorms))
	for i, seatGorm := range seatsGorms {
		seats[i] = seatGorm.ToDomain()
	}
	return seats, nil
}

func (r *gormSeatRepository) GetSeatsByIDs(ctx context.Context, seatIDs []uint) ([]*cinema.Seat, error) {
	logger := r.logger.With(applog.String("Method", "GetSeatsByIDs"), applog.Int("seat count", len(seatIDs)))
	var seatsGorms []*models.SeatGrom
	if err := r.db.WithContext(ctx).Where("id IN ?", seatIDs).Find(&seatsGorms).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("seats not found", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", cinema.ErrSeatNotFound, err)
		}
		logger.Error("failed to get seats by ids", applog.Error(err))
		return nil, fmt.Errorf("failed to get seats by ids: %w", err)
	}

	logger.Info("get seats by ids successfully")
	seats := make([]*cinema.Seat, len(seatsGorms))
	for i, seatGorm := range seatsGorms {
		seats[i] = seatGorm.ToDomain()
	}
	return seats, nil
}

func (r *gormSeatRepository) Update(ctx context.Context, seat *cinema.Seat) error {
	logger := r.logger.With(applog.String("Method", "Update"),
		applog.Uint("seat_id", uint(seat.ID)), applog.Uint("hall_id", uint(seat.CinemaHallID)))

	seatGorm := models.SeatGromFromDomain(seat)
	if err := r.db.WithContext(ctx).First(&models.SeatGrom{}, seatGorm.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("seat not found", applog.Error(err))
			return fmt.Errorf("%w: %w", cinema.ErrSeatNotFound, err)
		}
		logger.Error("failed to find seat by id", applog.Error(err))
		return fmt.Errorf("failed to find seat by id: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", seatGorm.ID).Updates(&seatGorm)
	if err := result.Error; err != nil {
		logger.Error("failed to update seat", applog.Error(err))
		return fmt.Errorf("failed to update seat: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("no rows affected during update")
		return shared.ErrNoRowsAffected
	}

	logger.Info("update seat successfully")
	return nil
}

func (r *gormSeatRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "Delete"), applog.Uint("seat_id", id))

	result := r.db.WithContext(ctx).Delete(&models.SeatGrom{}, id)
	if err := result.Error; err != nil {
		logger.Error("failed to delete seat", applog.Error(err))
		return fmt.Errorf("failed to delete seat: %w", err)
	}
	// 是否不存在或已删除
	if result.RowsAffected == 0 {
		logger.Warn("seat to delete not found or already deleted")
		return shared.ErrNoRowsAffected
	}

	logger.Info("delete seat successfully")
	return nil
}

func (r *gormSeatRepository) DeleteByHallID(ctx context.Context, hallID uint) error {
	logger := r.logger.With(applog.String("Method", "DeleteByHallID"), applog.Uint("hall_id", hallID))

	result := r.db.WithContext(ctx).Where("cinema_hall_id = ?", hallID).Delete(&models.SeatGrom{})
	if err := result.Error; err != nil {
		logger.Error("failed to delete seats", applog.Error(err))
		return fmt.Errorf("failed to delete seats: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("no seats found to delete")
		return shared.ErrNoRowsAffected
	}

	logger.Info("delete seats by hall id successfully")
	return nil
}
