package repository

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/shared/vo"
	"mrs/internal/domain/showtime"
	"mrs/internal/infrastructure/persistence/mysql/models"
	applog "mrs/pkg/log"
	"time"

	"gorm.io/gorm"
)

type gormShowtimeRepository struct {
	logger applog.Logger
	db     *gorm.DB
}

func NewGormShowtimeRepository(db *gorm.DB, logger applog.Logger) showtime.ShowtimeRepository {
	return &gormShowtimeRepository{
		db:     db,
		logger: logger.With(applog.String("Repository", "gormShowtimeRepository")),
	}
}

func (r *gormShowtimeRepository) Create(ctx context.Context, st *showtime.Showtime) (*showtime.Showtime, error) {
	logger := r.logger.With(
		applog.String("Method", "Create"),
		applog.Uint("showtime_id", uint(st.ID)),
		applog.Uint("movie_id", uint(st.MovieID)),
		applog.Uint("hall_id", uint(st.CinemaHallID)),
	)
	showtimeGorm := models.ShowtimeGormFromDomain(st)
	if err := r.db.WithContext(ctx).Create(showtimeGorm).Error; err != nil {
		logger.Error("database create showtime error", applog.Error(err))
		return nil, fmt.Errorf("database create showtime error: %w", err)
	}
	logger.Info("create showtime successfully")
	return showtimeGorm.ToDomain(), nil
}

// 预加载 Movie 和 CinemaHall
func (r *gormShowtimeRepository) FindByID(ctx context.Context, id vo.ShowtimeID) (*showtime.Showtime, error) {
	logger := r.logger.With(
		applog.String("Method", "FindByID"),
		applog.Uint("showtime_id", uint(id)))
	var showtimeGorm models.ShowtimeGorm
	if err := r.db.WithContext(ctx).
		Preload("Movie").
		Preload("CinemaHall").
		First(&showtimeGorm, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("showtime id not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %w", showtime.ErrShowtimeNotFound, err)
		}
		logger.Error("database find showtime by id error", applog.Error(err))
		return nil, fmt.Errorf("database find showtime by id error: %w", err)
	}

	logger.Info("find showtime by id successfully")
	return showtimeGorm.ToDomain(), nil
}

func (r *gormShowtimeRepository) FindByIDs(ctx context.Context, ids []vo.ShowtimeID) ([]*showtime.Showtime, error) {
	logger := r.logger.With(
		applog.String("method", "FindByIDs"),
		applog.Int("count", len(ids)),
	)
	// 将 ShowtimeID 切片转换为 uint 切片
	uintIDs := make([]uint, len(ids))
	for i, id := range ids {
		uintIDs[i] = uint(id)
	}
	var showtimesGorms []*models.ShowtimeGorm
	if err := r.db.WithContext(ctx).Where("id IN (?)", uintIDs).
		Preload("Movie").
		Preload("CinemaHall").
		Find(&showtimesGorms).Error; err != nil {
		logger.Error("database find showtimes by ids error", applog.Error(err))
		return nil, fmt.Errorf("database find showtimes by ids error: %w", err)
	}
	logger.Info("find showtimes by ids successfully", applog.Int("count", len(showtimesGorms)))
	showtimes := make([]*showtime.Showtime, len(showtimesGorms))
	for i, showtimeGorm := range showtimesGorms {
		showtimes[i] = showtimeGorm.ToDomain()
	}
	return showtimes, nil
}

// 分页查询支持过滤条件（如电影ID/影厅ID/日期范围）
func (r *gormShowtimeRepository) List(ctx context.Context, options *showtime.ShowtimeQueryOptions) ([]*showtime.Showtime, int64, error) {
	logger := r.logger.With(applog.String("method", "List"), applog.Any("options", options))
	var showtimesGorms []*models.ShowtimeGorm
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&models.ShowtimeGorm{})
	countQuery := r.db.WithContext(ctx).Model(&models.ShowtimeGorm{})

	if options.MovieID > 0 {
		query = query.Where("movie_id = ?", options.MovieID)
		countQuery = countQuery.Where("movie_id = ?", options.MovieID)
	}
	if options.CinemaHallID > 0 {
		query = query.Where("cinema_hall_id = ?", options.CinemaHallID)
		countQuery = countQuery.Where("cinema_hall_id = ?", options.CinemaHallID)
	}
	if !options.Date.IsZero() {
		query = query.Where("start_time >= ? AND start_time < ?", options.Date, options.Date.AddDate(0, 0, 1))
		countQuery = countQuery.Where("start_time >= ? AND start_time < ?", options.Date, options.Date.AddDate(0, 0, 1))
	}

	// 获取总数
	if err := countQuery.Count(&totalCount).Error; err != nil {
		logger.Error("database count showtimes error", applog.Error(err))
		return nil, 0, fmt.Errorf("database count showtimes error: %w", err)
	}

	if totalCount == 0 {
		logger.Info("no showtimes found matching criteria")
		return nil, 0, nil
	}

	// 应用排序和分页，并预加载关联数据
	offset := (options.Page - 1) * options.PageSize
	if err := query.Order("start_time ASC").
		Offset(offset).Limit(options.PageSize).
		Preload("Movie").
		Preload("CinemaHall").
		Find(&showtimesGorms).Error; err != nil {
		logger.Error("database list showtimes error", applog.Error(err))
		return nil, 0, fmt.Errorf("database list showtimes error: %w", err)
	}

	logger.Info("list showtimes successfully",
		applog.Int("count", len(showtimesGorms)),
		applog.Int64("total_count", totalCount))
	showtimes := make([]*showtime.Showtime, len(showtimesGorms))
	for i, showtimeGorm := range showtimesGorms {
		showtimes[i] = showtimeGorm.ToDomain()
	}
	return showtimes, totalCount, nil
}

func (r *gormShowtimeRepository) Update(ctx context.Context, st *showtime.Showtime) error {
	logger := r.logger.With(
		applog.String("Method", "Update"),
		applog.Uint("showtime_id", uint(st.ID)),
		applog.Uint("movie_id", uint(st.MovieID)),
		applog.Uint("hall_id", uint(st.CinemaHallID)),
	)

	// 基础设施层只负责数据访问，不负责业务逻辑
	// 因此，这里不进行业务逻辑的判断，直接更新数据库
	// 如果需要业务逻辑的判断，应该在服务层进行

	showtimeGorm := models.ShowtimeGormFromDomain(st)

	var exist int64
	if err := r.db.WithContext(ctx).Model(&models.ShowtimeGorm{}).Where("id = ?", st.ID).Count(&exist).Error; err != nil {
		logger.Error("database check showtime exist error", applog.Error(err))
		return fmt.Errorf("database check showtime exist error: %w", err)
	}

	if exist == 0 {
		logger.Warn("showtime not found")
		return fmt.Errorf("%w(id): %v", showtime.ErrShowtimeNotFound, st.ID)
	}

	if err := r.db.WithContext(ctx).Model(&models.ShowtimeGorm{}).Where("id = ?", st.ID).Updates(showtimeGorm).Error; err != nil {
		logger.Error("database update showtime error", applog.Error(err))
		return fmt.Errorf("database update showtime error: %w", err)
	}

	// 无论是否真正造成更新，都返回成功
	logger.Info("update showtime successfully")
	return nil
}

func (r *gormShowtimeRepository) Delete(ctx context.Context, id vo.ShowtimeID) error {
	logger := r.logger.With(
		applog.String("Method", "Delete"),
		applog.Uint("showtime_id", uint(id)),
	)

	result := r.db.WithContext(ctx).Delete(&models.ShowtimeGorm{}, uint(id))
	if result.Error != nil {
		logger.Error("database delete showtime error", applog.Error(result.Error))
		return fmt.Errorf("database delete showtime error: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		logger.Warn("showtime not found")
		return fmt.Errorf("%w(id): %v", showtime.ErrShowtimeNotFound, id)
	}

	logger.Info("delete showtime successfully")
	return nil
}

// CheckOverlap 检查指定影厅在给定时间段内是否存在与其他放映计划（可排除特定ID）的重叠。
// excludeShowtimeID 是一个可选参数，用于在更新场景下排除当前正在更新的放映计划自身。
func (r *gormShowtimeRepository) CheckOverlap(ctx context.Context, hallID vo.CinemaHallID,
	startTime, endTime time.Time, excludeShowtimeID ...vo.ShowtimeID) (bool, error) {
	logger := r.logger.With(
		applog.String("method", "CheckOverlap"),
		applog.Uint("hall_id", uint(hallID)),
		applog.Time("start_time", startTime),
		applog.Time("end_time", endTime),
	)
	if len(excludeShowtimeID) > 0 {
		logger = logger.With(applog.Uint("exclude_showtime_id", uint(excludeShowtimeID[0])))
	}

	uintExcludeShowtimeID := make([]uint, len(excludeShowtimeID))
	for i, id := range excludeShowtimeID {
		uintExcludeShowtimeID[i] = uint(id)
	}

	var count int64
	query := r.db.WithContext(ctx).Model(&models.ShowtimeGorm{}).
		Where("cinema_hall_id = ?", uint(hallID)).
		// 核心重叠逻辑:
		// 新场次的开始时间在新场次结束之前 AND 新场次的结束时间在现有场次开始之后
		Where("start_time < ?", endTime). // Existing showtime starts before new one ends
		Where("end_time > ?", startTime)  // Existing showtime ends after new one starts

	if len(uintExcludeShowtimeID) > 0 && uintExcludeShowtimeID[0] > 0 {
		query = query.Where("id != ?", uintExcludeShowtimeID[0])
	}

	if err := query.Count(&count).Error; err != nil {
		logger.Error("database count overlapping showtimes error", applog.Error(err))
		return false, fmt.Errorf("database count overlapping showtimes error: %w", err)
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
func (r *gormShowtimeRepository) FindShowtimesByMovieAndDateRanges(ctx context.Context, movieID vo.MovieID,
	startDate, endDate time.Time) ([]*showtime.Showtime, error) {
	logger := r.logger.With(
		applog.String("method", "FindShowtimesByMovieAndDateRange"),
		applog.Uint("movie_id", uint(movieID)),
		applog.Time("start_date", startDate),
		applog.Time("end_date", endDate),
	)

	var showtimesGorms []*models.ShowtimeGorm

	// 确保 endDate 包含一整天
	loc := startDate.Location() // 使用 startDate 的时区
	actualEndDate := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, loc)

	err := r.db.WithContext(ctx).
		Where("movie_id = ?", movieID).
		Where("start_time BETWEEN ? AND ?", startDate, actualEndDate). // BETWEEN 通常包含两端
		Order("start_time ASC").
		Preload("CinemaHall"). // 预加载影厅信息
		Find(&showtimesGorms).Error

	if err != nil {
		logger.Error("database find showtimes by movie and date range error", applog.Error(err))
		return nil, fmt.Errorf("database find showtimes by movie and date range error: %w", err)
	}
	logger.Info("find showtimes by movie and date range successfully", applog.Int("count", len(showtimesGorms)))

	showtimes := make([]*showtime.Showtime, len(showtimesGorms))
	for i, showtimeGorm := range showtimesGorms {
		showtimes[i] = showtimeGorm.ToDomain()
	}
	return showtimes, nil
}

// 查询指定影厅在日期范围内的所有场次
func (r *gormShowtimeRepository) FindShowtimesByHallAndDateRanges(ctx context.Context, hallID vo.CinemaHallID,
	startDate, endDate time.Time) ([]*showtime.Showtime, error) {
	logger := r.logger.With(
		applog.String("method", "FindShowtimesByHallAndDateRange"),
		applog.Uint("hall_id", uint(hallID)),
		applog.Time("start_date", startDate),
		applog.Time("end_date", endDate),
	)
	var showtimesGorms []*models.ShowtimeGorm

	err := r.db.WithContext(ctx).
		Where("cinema_hall_id = ?", uint(hallID)).
		Where("start_time BETWEEN ? AND ?").
		Order("start_time ASC").
		Preload("Movie").
		Find(&showtimesGorms).Error

	if err != nil {
		logger.Error("database find showtimes by hall and date range error", applog.Error(err))
		return nil, fmt.Errorf("database find showtimes by hall and date range error: %w", err)
	}

	logger.Info("find showtimes by hall and date range successfully", applog.Int("count", len(showtimesGorms)))
	showtimes := make([]*showtime.Showtime, len(showtimesGorms))
	for i, showtimeGorm := range showtimesGorms {
		showtimes[i] = showtimeGorm.ToDomain()
	}
	return showtimes, nil
}
