package repository

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared/vo"
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

	cinemaHallGorm := models.CinemaHallGormFromDomain(hall)
	if err := r.db.WithContext(ctx).Create(cinemaHallGorm).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			logger.Warn("cinema hall already eixsts", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", cinema.ErrCinemaHallAlreadyExists, err)
		}
		logger.Error("database create cinema hall error", applog.Error(err))
		return nil, fmt.Errorf("database create cinema hall error: %w", err)
	}

	logger.Info("create cinema hall successfully")
	return cinemaHallGorm.ToDomain(), nil
}

func (r *gormCinemaHallRepository) FindByID(ctx context.Context, id vo.CinemaHallID) (*cinema.CinemaHall, error) {
	logger := r.logger.With(applog.String("Method", "FindByID"), applog.Uint("hall_id", uint(id)))
	var hallGorm models.CinemaHallGorm
	if err := r.db.WithContext(ctx).
		Preload("Seats").
		First(&hallGorm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("cinema hall id not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %w", cinema.ErrCinemaHallNotFound, err)
		}
		logger.Error("database find cinema hall by id error", applog.Error(err))
		return nil, fmt.Errorf("database find cinema hall by id error: %w", err)
	}

	logger.Info("find cinema hall by id successfully")
	return hallGorm.ToDomain(), nil
}

func (r *gormCinemaHallRepository) FindByName(ctx context.Context, name string) (*cinema.CinemaHall, error) {
	logger := r.logger.With(applog.String("Method", "FindByName"), applog.String("name", name))
	var hallGorm models.CinemaHallGorm
	if err := r.db.WithContext(ctx).
		Preload("Seats").
		Where("name = ?", name).
		First(&hallGorm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("cinema hall name not found", applog.Error(err))
			return nil, fmt.Errorf("%w(name): %w", cinema.ErrCinemaHallNotFound, err)
		}
		logger.Error("database find cinema hall by name error", applog.Error(err))
		return nil, fmt.Errorf("database find cinema hall by name error: %w", err)
	}

	logger.Info("find cinema hall by name successfully")
	return hallGorm.ToDomain(), nil
}

func (r *gormCinemaHallRepository) ListAll(ctx context.Context) ([]*cinema.CinemaHall, error) {
	logger := r.logger.With(applog.String("Method", "ListAll"))
	var hallsGorms []*models.CinemaHallGorm
	if err := r.db.WithContext(ctx).
		Preload("Seats").
		Find(&hallsGorms).Error; err != nil {
		logger.Error("database list all cinema halls error", applog.Error(err))
		return nil, fmt.Errorf("database list all cinema halls error: %w", err)
	}

	logger.Info("list all cinema hall successfully")
	halls := make([]*cinema.CinemaHall, len(hallsGorms))
	for i, hallGorm := range hallsGorms {
		halls[i] = hallGorm.ToDomain()
	}
	return halls, nil
}

// 更新影厅
func (r *gormCinemaHallRepository) Update(ctx context.Context, hall *cinema.CinemaHall) error {
	logger := r.logger.With(applog.String("Method", "Update"),
		applog.Uint("hall_id", uint(hall.ID)), applog.String("name", hall.Name))

	cinemaHallGorm := models.CinemaHallGormFromDomain(hall)

	// 基础设施层只负责数据访问，不负责业务逻辑
	// 因此，这里不进行业务逻辑的判断，直接更新数据库
	// 如果需要业务逻辑的判断，应该在服务层进行

	// 先执行一个轻量级查询
	var exist int64
	if err := r.db.WithContext(ctx).Model(&models.CinemaHallGorm{}).Where("id = ?", cinemaHallGorm.ID).Count(&exist).Error; err != nil {
		logger.Error("database check cinema hall exist error", applog.Error(err))
		return fmt.Errorf("database check cinema hall exist error: %w", err)
	}

	if exist == 0 {
		logger.Warn("cinema hall not found")
		return fmt.Errorf("%w(id): %v", cinema.ErrCinemaHallNotFound, cinemaHallGorm.ID)
	}

	if err := r.db.WithContext(ctx).Model(&models.CinemaHallGorm{}).Where("id = ?", cinemaHallGorm.ID).Updates(cinemaHallGorm).Error; err != nil {
		logger.Error("database update cinema hall error", applog.Error(err))
		return fmt.Errorf("database update cinema hall error: %w", err)
	}

	// 无论是否真正造成更新，都返回成功
	logger.Info("update cinema hall successfully")
	return nil
}

// 删除影厅
func (r *gormCinemaHallRepository) Delete(ctx context.Context, id vo.CinemaHallID) error {
	logger := r.logger.With(applog.String("Method", "Delete"), applog.Uint("hall_id", uint(id)))

	result := r.db.WithContext(ctx).Delete(&models.CinemaHallGorm{}, id)
	if err := result.Error; err != nil {
		// 无需判断外键约束错误，业务上座位归属于影厅，影厅删除时，座位也会被删除，数据模型上外键约束为CASCADE
		logger.Error("database delete cinema hall error", applog.Error(err))
		return fmt.Errorf("database delete cinema hall error: %w", err)
	}

	// 是否不存在或已删除
	if result.RowsAffected == 0 {
		logger.Warn("cinema hall not found")
		return fmt.Errorf("%w(id): %v", cinema.ErrCinemaHallNotFound, id)
	}

	logger.Info("delete cinema hall successfully")
	return nil
}
