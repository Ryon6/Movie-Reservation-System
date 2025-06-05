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

type gormCinemaHallRepository struct {
	db     *gorm.DB
	logger applog.Logger
}

func NewGormCinemaHallRepository(db *gorm.DB, logger applog.Logger) cinema.CinemaHallRepository {
	return &gormCinemaHallRepository{
		db:     db,
		logger: logger.With(applog.String("Repository", "gormCinemaHallRepository")),
	}
}

func (r *gormCinemaHallRepository) Create(ctx context.Context, hall *cinema.CinemaHall) (*cinema.CinemaHall, error) {
	logger := r.logger.With(applog.String("Method", "Create"),
		applog.Uint("hall_id", uint(hall.ID)), applog.String("name", hall.Name))

	cinemaHallGorm := models.CinemaHallGromFromDomain(hall)
	if err := r.db.WithContext(ctx).Create(cinemaHallGorm).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			logger.Warn("cinema hall already eixsts", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", cinema.ErrCinemaHallAlreadyExists, err)
		}
		logger.Error("failed to create cinema hall", applog.Error(err))
		return nil, fmt.Errorf("failed to create cinema hall: %w", err)
	}

	logger.Info("create cinema hall successfully")
	return cinemaHallGorm.ToDomain(), nil
}

func (r *gormCinemaHallRepository) FindByID(ctx context.Context, id uint) (*cinema.CinemaHall, error) {
	logger := r.logger.With(applog.String("Method", "FindByID"), applog.Uint("hall_id", id))
	var hallGorm models.CinemaHallGrom
	if err := r.db.WithContext(ctx).
		Preload("Seats").
		First(&hallGorm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("cinema hall id not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %w", cinema.ErrCinemaHallNotFound, err)
		}
		logger.Error("failed to find cinema hall by id", applog.Error(err))
		return nil, fmt.Errorf("failed to find cinema hall by id: %w", err)
	}

	logger.Info("find cinema hall by id successfully")
	return hallGorm.ToDomain(), nil
}

func (r *gormCinemaHallRepository) FindByName(ctx context.Context, name string) (*cinema.CinemaHall, error) {
	logger := r.logger.With(applog.String("Method", "FindByName"), applog.String("name", name))
	var hallGorm models.CinemaHallGrom
	if err := r.db.WithContext(ctx).
		Preload("Seats").
		Where("name = ?", name).
		First(&hallGorm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("cinema hall name not found", applog.Error(err))
			return nil, fmt.Errorf("%w(name): %w", cinema.ErrCinemaHallNotFound, err)
		}
		logger.Error("failed to find cinema hall by name", applog.Error(err))
		return nil, fmt.Errorf("failed to find cinema hall by name: %w", err)
	}

	logger.Info("find cinema hall by name successfully")
	return hallGorm.ToDomain(), nil
}

func (r *gormCinemaHallRepository) ListAll(ctx context.Context) ([]*cinema.CinemaHall, error) {
	logger := r.logger.With(applog.String("Method", "ListAll"))
	var hallsGorms []*models.CinemaHallGrom
	if err := r.db.WithContext(ctx).
		Preload("Seats").
		Find(&hallsGorms).Error; err != nil {
		logger.Error("failed to list all cinema halls", applog.Error(err))
		return nil, fmt.Errorf("failed to list all cinema halls: %w", err)
	}

	logger.Info("list all cinema hall successfully")
	halls := make([]*cinema.CinemaHall, len(hallsGorms))
	for i, hallGorm := range hallsGorms {
		halls[i] = hallGorm.ToDomain()
	}
	return halls, nil
}
func (r *gormCinemaHallRepository) Update(ctx context.Context, hall *cinema.CinemaHall) error {
	logger := r.logger.With(applog.String("Method", "Update"),
		applog.Uint("hall_id", uint(hall.ID)), applog.String("name", hall.Name))

	cinemaHallGorm := models.CinemaHallGromFromDomain(hall)
	if err := r.db.WithContext(ctx).First(&models.CinemaHallGrom{}, cinemaHallGorm.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("cinema hall not found", applog.Error(err))
			return fmt.Errorf("%w: %w", cinema.ErrCinemaHallNotFound, err)
		}
		logger.Error("failed to find cinema hall by id", applog.Error(err))
		return fmt.Errorf("failed to find cinema hall by id: %w", err)
	}

	result := r.db.WithContext(ctx).
		Model(&models.CinemaHallGrom{}).
		Where("id = ?", cinemaHallGorm.ID).
		Updates(cinemaHallGorm)
	if err := result.Error; err != nil {
		logger.Error("failed to update cinema hall", applog.Error(err))
		return fmt.Errorf("failed to update cinema hall: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("no rows affected during update")
		return shared.ErrNoRowsAffected
	}

	logger.Info("Update cinema hall successfully")
	return nil
}

func (r *gormCinemaHallRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "Delete"), applog.Uint("hall_id", id))

	result := r.db.WithContext(ctx).Delete(&models.CinemaHallGrom{}, id)
	if err := result.Error; err != nil {
		// 是否为外键约束错误，是则返回哨兵错误，有服务层进一步处理
		if isForeignKeyConstraintError(err) {
			logger.Warn("cannot delete cinema hall due to foreign key constraint", applog.Error(err))
			return fmt.Errorf("%w: %w", cinema.ErrCinemaHallReferenced, err)
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("cinema hall to delete not found", applog.Error(err))
			return fmt.Errorf("%w: %w", cinema.ErrCinemaHallNotFound, err)
		}
		logger.Error("failed to delete cinema hall", applog.Error(err))
		return fmt.Errorf("failed to delete cinema hall: %w", err)
	}

	// 是否不存在或已删除
	if result.RowsAffected == 0 {
		logger.Warn("cinema hall to delete not found or already deleted")
		return shared.ErrNoRowsAffected
	}

	logger.Info("delete cinema hall successfully")
	return nil
}
