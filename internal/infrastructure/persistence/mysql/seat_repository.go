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

	seatGorms := make([]*models.SeatGorm, len(seats))
	for i, seat := range seats {
		seatGorms[i] = models.SeatGormFromDomain(seat)
	}
	if err := r.db.WithContext(ctx).Create(&seatGorms).Error; err != nil {
		logger.Error("database create seats error", applog.Error(err))
		return nil, fmt.Errorf("database create seats error: %w", err)
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
	var seatGorm models.SeatGorm
	if err := r.db.WithContext(ctx).First(&seatGorm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("seat id not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %w", cinema.ErrSeatNotFound, err)
		}
		logger.Error("database find seat by id error", applog.Error(err))
		return nil, fmt.Errorf("database find seat by id error: %w", err)
	}

	logger.Info("find seat by id successfully")
	return seatGorm.ToDomain(), nil
}

func (r *gormSeatRepository) FindByHallID(ctx context.Context, hallID uint) ([]*cinema.Seat, error) {
	logger := r.logger.With(applog.String("Method", "FindByHallID"), applog.Uint("hall_id", hallID))
	var seatsGorms []*models.SeatGorm

	if err := r.db.WithContext(ctx).Where("cinema_hall_id = ?", hallID).Find(&seatsGorms).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("no seats found for hall", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", cinema.ErrCinemaHallNotFound, err)
		}
		logger.Error("database find seats by hall id error", applog.Error(err))
		return nil, fmt.Errorf("database find seats by hall id error: %w", err)
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
	var seatsGorms []*models.SeatGorm
	if err := r.db.WithContext(ctx).Where("id IN ?", seatIDs).Find(&seatsGorms).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("seats not found", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", cinema.ErrSeatNotFound, err)
		}
		logger.Error("database get seats by ids error", applog.Error(err))
		return nil, fmt.Errorf("database get seats by ids error: %w", err)
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

	seatGorm := models.SeatGormFromDomain(seat)
	// 基础设施层只负责数据访问，不负责业务逻辑
	// 因此，这里不进行业务逻辑的判断，直接更新数据库
	// 如果需要业务逻辑的判断，应该在服务层进行

	var exist int64
	if err := r.db.WithContext(ctx).Model(&models.SeatGorm{}).Where("id = ?", seatGorm.ID).Count(&exist).Error; err != nil {
		logger.Error("database check seat exist error", applog.Error(err))
		return fmt.Errorf("database check seat exist error: %w", err)
	}

	if exist == 0 {
		logger.Warn("seat not found")
		return fmt.Errorf("%w(id): %v", cinema.ErrSeatNotFound, seatGorm.ID)
	}

	if err := r.db.WithContext(ctx).Model(&models.SeatGorm{}).Where("id = ?", seatGorm.ID).Updates(&seatGorm).Error; err != nil {
		logger.Error("database update seat error", applog.Error(err))
		return fmt.Errorf("database update seat error: %w", err)
	}

	// 无论是否真正造成更新，都返回成功
	logger.Info("update seat successfully")
	return nil
}

// 删除座位
func (r *gormSeatRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "Delete"), applog.Uint("seat_id", id))

	result := r.db.WithContext(ctx).Delete(&models.SeatGorm{}, id)
	if result.Error != nil {
		logger.Error("database delete seat error", applog.Error(result.Error))
		return fmt.Errorf("database delete seat error: %w", result.Error)
	}

	// 是否不存在或已删除
	if result.RowsAffected == 0 {
		logger.Warn("seat not found")
		return fmt.Errorf("%w(id): %v", cinema.ErrSeatNotFound, id)
	}

	logger.Info("delete seat successfully")
	return nil
}

func (r *gormSeatRepository) DeleteByHallID(ctx context.Context, hallID uint) error {
	logger := r.logger.With(applog.String("Method", "DeleteByHallID"), applog.Uint("hall_id", hallID))

	result := r.db.WithContext(ctx).Where("cinema_hall_id = ?", hallID).Delete(&models.SeatGorm{})
	if err := result.Error; err != nil {
		logger.Error("database delete seats error", applog.Error(err))
		return fmt.Errorf("database delete seats error: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("no seats found to delete")
		return shared.ErrNoRowsAffected
	}

	logger.Info("delete seats by hall id successfully")
	return nil
}
