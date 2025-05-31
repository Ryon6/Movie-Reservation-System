package mysql

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/movie"
	applog "mrs/pkg/log"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type gormGenreRepository struct {
	db     *gorm.DB
	logger applog.Logger
}

func NewGormGenreRepository(db *gorm.DB, logger applog.Logger) movie.GenreRepository {
	return &gormGenreRepository{
		db:     db,
		logger: logger,
	}
}

func (r *gormGenreRepository) Create(ctx context.Context, genre *movie.Genre) error {
	logger := r.logger.With(applog.String("Method", "gormGenreRepository.Create"),
		applog.Uint("genre_id", genre.ID), applog.String("name", genre.Name))
	if err := r.db.WithContext(ctx).Create(genre).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			logger.Warn("genre name already exists", applog.String("name", genre.Name), applog.Error(err))
			return fmt.Errorf("genre name already exists: %w", movie.ErrGenreAlreadyExists)
		}
		logger.Error("failed to create genre", applog.Uint("genre_id", genre.ID), applog.Error(err))
		return fmt.Errorf("failed to create genre: %w", err)
	}
	logger.Info("create genre successfully", applog.Uint("genre_id", genre.ID))
	return nil
}

func (r *gormGenreRepository) FindByID(ctx context.Context, id uint) (*movie.Genre, error) {
	logger := r.logger.With(applog.String("Method", "gormGenreRepository.FindByID"), applog.Uint("genre_id", id))
	var genre movie.Genre
	if err := r.db.WithContext(ctx).First(&genre, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("genre id not found", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", movie.ErrGenreNotFound, err)
		}
		logger.Error("failed to find genre by id", applog.Error(err))
		return nil, fmt.Errorf("failed to find genre by id: %w", err)
	}
	logger.Info("find genre successfully", applog.String("name", genre.Name))
	return &genre, nil
}

func (r *gormGenreRepository) FindByName(ctx context.Context, name string) (*movie.Genre, error) {
	logger := r.logger.With(applog.String("Method", "gormGenreRepository.FindByName"),
		applog.String("name", name))
	var genre movie.Genre
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&genre).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("genre name not found", applog.Error(err))
			return nil, fmt.Errorf("genre name not found: %w", movie.ErrGenreNotFound)
		}
		logger.Error("failed to find genre by name", applog.Error(err))
		return nil, fmt.Errorf("failed to find genre by name: %w", err)
	}
	logger.Info("find genre successfully", applog.Uint("genre_id", genre.ID))
	return &genre, nil
}

func (r *gormGenreRepository) ListAll(ctx context.Context) ([]*movie.Genre, error) {
	logger := r.logger.With(applog.String("Method", "gormGenreRepository.ListAll"))
	var genres []*movie.Genre
	if err := r.db.WithContext(ctx).Find(&genres).Error; err != nil {
		logger.Error("failed to list all genres", applog.Error(err))
		return nil, fmt.Errorf("failed to list all genres: %w", err)
	}
	logger.Info("list all genres successfully", applog.Int("count", len(genres)))
	return genres, nil
}

func (r *gormGenreRepository) Update(ctx context.Context, genre *movie.Genre) error {
	logger := r.logger.With(applog.String("Method", "gormGenreRepository.Update"),
		applog.Uint("genre_id", genre.ID), applog.String("name", genre.Name))

	// 检查存在性
	if _, err := r.FindByID(ctx, genre.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("update failed: genre not found")
			return movie.ErrGenreNotFound
		}
		return fmt.Errorf("failed to find by id: %w", err)
	}

	// 使用Updates只更新非零值字段
	result := r.db.WithContext(ctx).Model(&movie.Genre{}).
		Where("id = ?", genre.ID).
		Updates(genre)

	if result.Error != nil {
		logger.Error("failed to update genre", applog.Error(result.Error))
		return fmt.Errorf("failed to update genre: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		logger.Warn("no rows affected during update")
		return movie.ErrGenreNotFound
	}

	logger.Info("update genre successfully")
	return nil
}

func (r *gormGenreRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "gormGenreRepository.Delete"), applog.Uint("genre_id", id))

	result := r.db.WithContext(ctx).Delete(&movie.Genre{}, id)

	if err := result.Error; err != nil {
		// 是否为外键约束错误，是则返回哨兵错误，有服务层进一步处理
		if isForeignKeyConstraintError(err) {
			logger.Warn("cannot delete genre due to foreign key constraint", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrGenreReferenced, err)
		}
		logger.Error("failed to delete genre", applog.Error(err))
		return fmt.Errorf("failed to delete genre: %w", err)
	}

	// 是否不存在或已删除
	if result.RowsAffected == 0 {
		logger.Warn("Genre to delete not found or already deleted")
		return movie.ErrGenreNotFound
	}
	logger.Info("delete genre successfully")
	return nil
}

// isForeignKeyConstraintError 检查错误是否是MySQL外键约束错误（错误码1451）
func isForeignKeyConstraintError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1451
	}
	return false
}

// FindOrCreateByName 查找指定名称的类型，如果不存在则创建它。
func (r *gormGenreRepository) FindOrCreateByName(ctx context.Context, name string) (*movie.Genre, error) {
	log := r.logger.With(applog.String("method", "gormGenreRepository.FindOrCreateByName"), applog.String("genre_name", name))
	var genre movie.Genre

	// GORM 的 FirstOrCreate 会原子地执行此操作（如果数据库支持）
	// 它首先尝试查找，如果未找到，则使用提供的属性创建。
	if err := r.db.WithContext(ctx).FirstOrCreate(&genre, movie.Genre{Name: name}).Error; err != nil {
		log.Error("failed to find or create genre by name", applog.Error(err))
		return nil, err
	}
	// RowsAffected 可以用来判断是找到还是创建，但这依赖于GORM的具体行为和版本
	// 通常可以直接返回 genre 实例
	log.Info("find or create genre successfully", applog.Uint("genre_id", genre.ID))
	return &genre, nil
}
