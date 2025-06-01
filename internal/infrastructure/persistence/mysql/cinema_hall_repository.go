package mysql

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/movie"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type gormCinemaHallRepository struct {
	db     *gorm.DB
	logger applog.Logger
}

func NewGormCinemaHallRepository(db *gorm.DB, logger applog.Logger) movie.CinemaHallRepository {
	return &gormCinemaHallRepository{
		db:     db,
		logger: logger,
	}
}

func (r *gormCinemaHallRepository) Create(ctx context.Context, hall *movie.CinemaHall) error {
	logger := r.logger.With(applog.String("Method", "gormCinemaHallRepository.Create"),
		applog.Uint("hall_id", hall.ID), applog.String("name", hall.Name))

	if err := r.db.WithContext(ctx).Create(hall).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			logger.Warn("cinema hall already eixsts", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrCinemaHallAlreadyExists, err)
		}
		logger.Error("failed to create cinema hall", applog.Error(err))
		return fmt.Errorf("failed to create cinema hall: %w", err)

	}

	logger.Info("create cinema hall successfully")
	return nil
}

func (r *gormCinemaHallRepository) FindByID(ctx context.Context, id uint) (*movie.CinemaHall, error) {
	logger := r.logger.With(applog.String("Method", "gormCinemaHallRepository.FindByID"), applog.Uint("hall_id", id))
	var hall movie.CinemaHall
	if err := r.db.WithContext(ctx).
		Preload("Seats").
		Preload("Showtimes").
		First(&hall, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("cinema hall id not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %w", movie.ErrCinemaHallNotFound, err)
		}
		logger.Error("failed to find cinema hall by id", applog.Error(err))
		return nil, fmt.Errorf("failed to find cinema hall by id: %w", err)
	}

	logger.Info("find cinema hall by id successfully")
	return &hall, nil
}

func (r *gormCinemaHallRepository) FindByName(ctx context.Context, name string) (*movie.CinemaHall, error) {
	logger := r.logger.With(applog.String("Method", "gormCinemaHallRepository.FindByName"), applog.String("name", name))
	var hall movie.CinemaHall
	if err := r.db.WithContext(ctx).
		Preload("Seats").
		Preload("Showtimes").
		Where("name = ?", name).
		First(&hall).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("cinema hall name not found", applog.Error(err))
			return nil, fmt.Errorf("%w(name): %w", movie.ErrCinemaHallNotFound, err)
		}
		logger.Error("failed to find cinema hall by name", applog.Error(err))
		return nil, fmt.Errorf("failed to find cinema hall by name: %w", err)
	}

	logger.Info("find cinema hall by name successfully")
	return &hall, nil
}

func (r *gormCinemaHallRepository) ListAll(ctx context.Context) ([]*movie.CinemaHall, error) {
	logger := r.logger.With(applog.String("Method", "gormCinemaHallRepository.ListAll"))
	var halls []*movie.CinemaHall
	if err := r.db.WithContext(ctx).
		Preload("Seats").
		Find(&halls).Error; err != nil {
		logger.Error("failed to list all cinema halls", applog.Error(err))
		return nil, fmt.Errorf("failed to list all cinema halls: %w", err)
	}

	logger.Info("list all cinema hall successfully")
	return halls, nil
}

func (r *gormCinemaHallRepository) Update(ctx context.Context, hall *movie.CinemaHall) error {
	logger := r.logger.With(applog.String("Method", "gormCinemaHallRepository.Update"),
		applog.Uint("hall_id", hall.ID), applog.String("name", hall.Name))

	if err := r.db.WithContext(ctx).First(&movie.CinemaHall{}, hall.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("cinema hall not found", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrCinemaHallNotFound, err)
		}
		logger.Error("failed to find cinema hall by id", applog.Error(err))
		return fmt.Errorf("failed to find cinema hall by id: %w", err)
	}

	result := r.db.WithContext(ctx).
		Model(&movie.CinemaHall{}).
		Where("id = ?", hall.ID).
		Updates(hall)
	if err := result.Error; err != nil {
		logger.Error("failed to update cinema hall", applog.Error(err))
		return fmt.Errorf("failed to update cinema hall: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("no rows affected during update")
		return movie.ErrNoRowsAffected
	}

	logger.Info("Update cinema hall successfully")
	return nil
}

func (r *gormCinemaHallRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "gormCinemaHallRepository.Delete"), applog.Uint("hall_id", id))

	result := r.db.WithContext(ctx).Delete(&movie.CinemaHall{}, id)
	if err := result.Error; err != nil {
		// 是否为外键约束错误，是则返回哨兵错误，有服务层进一步处理
		if isForeignKeyConstraintError(err) {
			logger.Warn("cannot delete cinema hall due to foreign key constraint", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrCinemaHallReferenced, err)
		}
		logger.Error("failed to delete cinema hall", applog.Error(err))
		return fmt.Errorf("failed to delete cinema hall: %w", err)
	}

	// 是否不存在或已删除
	if result.RowsAffected == 0 {
		logger.Warn("cinema hall to delete not found or already deleted")
		return movie.ErrCinemaHallNotFound
	}

	logger.Info("delete cinema hall successfully")
	return nil
}
