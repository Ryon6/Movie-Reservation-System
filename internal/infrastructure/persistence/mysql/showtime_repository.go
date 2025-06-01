package mysql

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/movie"
	applog "mrs/pkg/log"
	"time"

	"gorm.io/gorm"
)

type gormShowtimeRepository struct {
	logger applog.Logger
	db     *gorm.DB
}

func NewGormShowtimeRepository(db *gorm.DB, logger applog.Logger) movie.ShowtimeRepository {
	return &gormShowtimeRepository{
		db:     db,
		logger: logger,
	}
}

func (r *gormShowtimeRepository) Create(ctx context.Context, showtime *movie.Showtime) error {
	logger := r.logger.With(
		applog.String("Method", "gormShowtimeRepository.Create"),
		applog.Uint("showtime_id", showtime.ID),
		applog.Uint("movie_id", showtime.MovieID),
		applog.Uint("hall_id", showtime.CinemaHallID),
	)
	if err := r.db.WithContext(ctx).Create(showtime).Error; err != nil {
		logger.Error("failed to create showtime", applog.Error(err))
		return fmt.Errorf("failed to create showtime: %w", err)
	}
	logger.Info("create showtime successfully")
	return nil
}

// 预加载 Movie 和 CinemaHall
func (r *gormShowtimeRepository) FindByID(ctx context.Context, id uint) (*movie.Showtime, error) {
	logger := r.logger.With(
		applog.String("Method", "gormShowtimeRepository.FindByID"),
		applog.Uint("showtime_id", id))
	var showtime movie.Showtime
	if err := r.db.WithContext(ctx).
		Preload("Movie").
		Preload("CinemaHall").
		First(&showtime, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("showtime id not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %w", movie.ErrShowtimeNotFound, err)
		}
		logger.Error("failed to find showtime by id", applog.Error(err))
		return nil, fmt.Errorf("failed to find showtime by id: %w", err)
	}

	logger.Info("find showtime by id successfully")
	return &showtime, nil
}

// 分页查询支持过滤条件（如电影ID/影厅ID/日期范围）
func (r *gormShowtimeRepository) List(ctx context.Context, page, pageSize int,
	filters map[string]interface{}) ([]*movie.Showtime, int64, error) {
	logger := r.logger.With(
		applog.String("method", "gormShowtimeRepository.List"),
		applog.Int("page", page),
		applog.Int("pageSize", pageSize))
	var showtimes []*movie.Showtime
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&movie.Showtime{})
	countQuery := r.db.WithContext(ctx).Model(&movie.Showtime{})

	// 应用过滤器
	if movieID, ok := filters["movie_id"].(uint); ok && movieID > 0 {
		query = query.Where("movie_id = ?", movieID)
		countQuery = countQuery.Where("movie_id = ?", movieID)
		logger = logger.With(applog.Uint("filter_movie_id", movieID))
	}
	if cinemaHallID, ok := filters["cinema_hall_id"].(uint); ok && cinemaHallID > 0 {
		query = query.Where("cinema_hall_id = ?", cinemaHallID)
		countQuery = countQuery.Where("cinema_hall_id = ?", cinemaHallID)
		logger = logger.With(applog.Uint("filter_cinema_hall_id", cinemaHallID))
	}
	if date, ok := filters["date"].(time.Time); ok && !date.IsZero() {
		// 假设 "date" 过滤器是指某一天内的所有场次
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.AddDate(0, 0, 1)
		query = query.Where("start_time >= ? AND start_time < ?", startOfDay, endOfDay)
		countQuery = countQuery.Where("start_time >= ? AND start_time < ?", startOfDay, endOfDay)
		logger = logger.With(applog.String("filter_date", date.Format("2006-01-02")))
	}
	// 可以添加更多过滤器，例如按价格范围、未来的场次等

	// 获取总数
	if err := countQuery.Count(&totalCount).Error; err != nil {
		logger.Error("failed to count showtimes", applog.Error(err))
		return nil, 0, err
	}

	if totalCount == 0 {
		logger.Info("no showtimes found matching criteria")
		return showtimes, 0, nil
	}

	// 应用排序和分页，并预加载关联数据
	offset := (page - 1) * pageSize
	if err := query.Order("start_time ASC").
		Offset(offset).Limit(pageSize).
		Preload("Movie").
		Preload("CinemaHall").
		Find(&showtimes).Error; err != nil {
		logger.Error("failed to list showtimes", applog.Error(err))
		return nil, 0, err
	}

	logger.Info("list showtimes successfully",
		applog.Int("count", len(showtimes)),
		applog.Int64("total_count", totalCount))
	return showtimes, totalCount, nil
}

func (r *gormShowtimeRepository) Update(ctx context.Context, showtime *movie.Showtime) error {
	logger := r.logger.With(
		applog.String("Method", "gormShowtimeRepository.Update"),
		applog.Uint("showtime_id", showtime.ID),
		applog.Uint("movie_id", showtime.MovieID),
		applog.Uint("hall_id", showtime.CinemaHallID),
	)

	if err := r.db.WithContext(ctx).First(&movie.Showtime{}, showtime.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("showtime not found", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrShowtimeNotFound, err)
		}
		logger.Error("failed to find showtime by id", applog.Error(err))
		return fmt.Errorf("failed to find showtime by id: %w", err)
	}

	result := r.db.WithContext(ctx).Model(&movie.Showtime{}).Where("id = ?", showtime.ID).Updates(showtime)
	if err := result.Error; err != nil {
		logger.Error("failed to update showtime", applog.Error(err))
		return fmt.Errorf("failed to update showtime: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("no rows affected during update")
		return movie.ErrNoRowsAffected
	}

	logger.Info("update showtime successfully")
	return nil
}

func (r *gormShowtimeRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(
		applog.String("Method", "gormShowtimeRepository.Delete"),
		applog.Uint("showtime_id", id),
	)

	result := r.db.WithContext(ctx).Delete(&movie.Showtime{}, id)
	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("showtime to delete not found", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrShowtimeNotFound, err)
		}
		logger.Error("failed to delete showtime", applog.Error(err))
		return fmt.Errorf("failed to delete showtime: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("movie to delete not found or already deleted")
		return movie.ErrNoRowsAffected
	}

	logger.Info("delete showtime successfully")
	return nil
}

// CheckOverlap 检查指定影厅在给定时间段内是否存在与其他放映计划（可排除特定ID）的重叠。
// excludeShowtimeID 是一个可选参数，用于在更新场景下排除当前正在更新的放映计划自身。
func (r *gormShowtimeRepository) CheckOverlap(ctx context.Context, hallID uint,
	startTime, endTime time.Time, excludeShowtimeID ...uint) (bool, error) {
	logger := r.logger.With(
		applog.String("method", "CheckOverlap"),
		applog.Uint("hall_id", hallID),
		applog.Time("start_time", startTime),
		applog.Time("end_time", endTime),
	)
	if len(excludeShowtimeID) > 0 {
		logger = logger.With(applog.Uint("exclude_showtime_id", excludeShowtimeID[0]))
	}

	var count int64
	query := r.db.WithContext(ctx).Model(&movie.Showtime{}).
		Where("cinema_hall_id = ?", hallID).
		// 核心重叠逻辑:
		// 新场次的开始时间在新场次结束之前 AND 新场次的结束时间在现有场次开始之后
		Where("start_time < ?", endTime). // Existing showtime starts before new one ends
		Where("end_time > ?", startTime)  // Existing showtime ends after new one starts

	if len(excludeShowtimeID) > 0 && excludeShowtimeID[0] > 0 {
		query = query.Where("id != ?", excludeShowtimeID[0])
	}

	if err := query.Count(&count).Error; err != nil {
		logger.Error("failed to count overlapping showtimes", applog.Error(err))
		return false, err
	}

	isOverlapping := count > 0
	if isOverlapping {
		logger.Warn("showtime overlap detected")
	} else {
		logger.Info("no showtime overlap detected")
	}
	return isOverlapping, nil
}

// 查询指定电影在日期范围内的所有场次
func (r *gormShowtimeRepository) FindShowtimesByMovieAndDateRanges(ctx context.Context, movieID uint,
	startDate, endDate time.Time) ([]*movie.Showtime, error) {
	logger := r.logger.With(
		applog.String("method", "FindShowtimesByMovieAndDateRange"),
		applog.Uint("movie_id", movieID),
		applog.Time("start_date", startDate),
		applog.Time("end_date", endDate),
	)

	var showtimes []*movie.Showtime

	// 确保 endDate 包含一整天
	loc := startDate.Location() // 使用 startDate 的时区
	actualEndDate := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, loc)

	err := r.db.WithContext(ctx).
		Where("movie_id = ?", movieID).
		Where("start_time BETWEEN ? AND ?", startDate, actualEndDate). // BETWEEN 通常包含两端
		Order("start_time ASC").
		Preload("CinemaHall"). // 预加载影厅信息
		Find(&showtimes).Error

	if err != nil {
		logger.Error("failed to find showtimes by movie and date range", applog.Error(err))
		return nil, err
	}
	logger.Info("find showtimes by movie and date range successfully", applog.Int("count", len(showtimes)))

	return showtimes, nil
}

// 查询指定影厅在日期范围内的所有场次
func (r *gormShowtimeRepository) FindShowtimesByHallAndDateRanges(ctx context.Context, hallID uint,
	startDate, endDate time.Time) ([]*movie.Showtime, error) {
	logger := r.logger.With(
		applog.String("method", "FindShowtimesByHallAndDateRange"),
		applog.Uint("hall_id", hallID),
		applog.Time("start_date", startDate),
		applog.Time("end_date", endDate),
	)
	var showtimes []*movie.Showtime

	err := r.db.WithContext(ctx).
		Where("cinema_hall_id = ?", hallID).
		Where("start_time BETWEEN ? AND ?").
		Order("start_time ASC").
		Preload("Movie").
		Find(&showtimes).Error

	if err != nil {
		logger.Error("Failed to find showtimes by hall and date range", applog.Error(err))
		return nil, err
	}

	logger.Info("find showtimes by hall and date range successfully", applog.Int("count", len(showtimes)))
	return showtimes, nil
}
