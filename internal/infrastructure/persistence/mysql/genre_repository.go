package mysql

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/movie"
	"mrs/internal/domain/shared"
	"mrs/internal/infrastructure/persistence/mysql/models"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type gormGenreRepository struct {
	db     *gorm.DB
	logger applog.Logger
}

func NewGormGenreRepository(db *gorm.DB, logger applog.Logger) movie.GenreRepository {
	return &gormGenreRepository{
		db:     db,
		logger: logger.With(applog.String("Repository", "gormGenreRepository")),
	}
}

func (r *gormGenreRepository) Create(ctx context.Context, genre *movie.Genre) (*movie.Genre, error) {
	logger := r.logger.With(applog.String("Method", "Create"),
		applog.Uint("genre_id", uint(genre.ID)), applog.String("name", genre.Name))
	genreGorm := models.GenreGormFromDomain(genre)
	if err := r.db.WithContext(ctx).Create(genreGorm).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			logger.Warn("genre name already exists", applog.String("name", genre.Name), applog.Error(err))
			return nil, fmt.Errorf("%w: %w", movie.ErrGenreAlreadyExists, err)
		}
		logger.Error("failed to create genre", applog.Uint("genre_id", uint(genre.ID)), applog.Error(err))
		return nil, fmt.Errorf("failed to create genre: %w", err)
	}
	logger.Info("create genre successfully", applog.Uint("genre_id", uint(genre.ID)))
	return genreGorm.ToDomain(), nil
}

func (r *gormGenreRepository) FindByID(ctx context.Context, id uint) (*movie.Genre, error) {
	logger := r.logger.With(applog.String("Method", "FindByID"), applog.Uint("genre_id", id))
	var genreGorm models.GenreGorm
	if err := r.db.WithContext(ctx).First(&genreGorm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("genre id not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %w", movie.ErrGenreNotFound, err)
		}
		logger.Error("failed to find genre by id", applog.Error(err))
		return nil, fmt.Errorf("failed to find genre by id: %w", err)
	}
	logger.Info("find genre successfully", applog.String("name", genreGorm.Name))
	return genreGorm.ToDomain(), nil
}

func (r *gormGenreRepository) FindByName(ctx context.Context, name string) (*movie.Genre, error) {
	logger := r.logger.With(applog.String("Method", "FindByName"),
		applog.String("name", name))
	var genreGorm models.GenreGorm
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&genreGorm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("genre name not found", applog.Error(err))
			return nil, fmt.Errorf("%w(name): %w", movie.ErrGenreNotFound, err)
		}
		logger.Error("failed to find genre by name", applog.Error(err))
		return nil, fmt.Errorf("failed to find genre by name: %w", err)
	}
	logger.Info("find genre successfully", applog.Uint("genre_id", uint(genreGorm.ID)))
	return genreGorm.ToDomain(), nil
}

func (r *gormGenreRepository) ListAll(ctx context.Context) ([]*movie.Genre, error) {
	logger := r.logger.With(applog.String("Method", "ListAll"))
	var genresGorms []*models.GenreGorm
	if err := r.db.WithContext(ctx).Find(&genresGorms).Error; err != nil {
		logger.Error("failed to list all genres", applog.Error(err))
		return nil, fmt.Errorf("failed to list all genres: %w", err)
	}
	logger.Info("list all genres successfully", applog.Int("count", len(genresGorms)))
	genres := make([]*movie.Genre, len(genresGorms))
	for i, genreGorm := range genresGorms {
		genres[i] = genreGorm.ToDomain()
	}
	return genres, nil
}

func (r *gormGenreRepository) Update(ctx context.Context, genre *movie.Genre) error {
	logger := r.logger.With(applog.String("Method", "Update"),
		applog.Uint("genre_id", uint(genre.ID)), applog.String("name", genre.Name))

	// 检查存在性
	genreGorm := models.GenreGormFromDomain(genre)
	if _, err := r.FindByID(ctx, genreGorm.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("genre id not found", applog.Error(err))
			return fmt.Errorf("%w(id): %w", movie.ErrGenreNotFound, err)
		}
		return fmt.Errorf("failed to find by id: %w", err)
	}

	// 使用Updates只更新非零值字段
	result := r.db.WithContext(ctx).Model(&models.GenreGorm{}).
		Where("id = ?", genreGorm.ID).
		Updates(genreGorm)

	if result.Error != nil {
		logger.Error("failed to update genre", applog.Error(result.Error))
		return fmt.Errorf("failed to update genre: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		logger.Warn("no rows affected during update")
		return shared.ErrNoRowsAffected
	}

	logger.Info("update genre successfully")
	return nil
}

func (r *gormGenreRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "Delete"), applog.Uint("genre_id", id))

	result := r.db.WithContext(ctx).Delete(&models.GenreGorm{}, id)

	if err := result.Error; err != nil {
		// 是否为外键约束错误，是则返回哨兵错误，有服务层进一步处理
		if isForeignKeyConstraintError(err) {
			logger.Warn("cannot delete genre due to foreign key constraint", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrGenreReferenced, err)
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("genre to delete not found", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrGenreNotFound, err)
		}
		logger.Error("failed to delete genre", applog.Error(err))
		return fmt.Errorf("failed to delete genre: %w", err)
	}

	// 是否不存在或已删除
	if result.RowsAffected == 0 {
		logger.Warn("genre to delete not found or already deleted")
		return shared.ErrNoRowsAffected
	}
	logger.Info("delete genre successfully")
	return nil
}

// FindOrCreateByNames 查找指定名称的类型，如果不存在则创建它。
func (r *gormGenreRepository) FindOrCreateByNames(ctx context.Context, names []string) ([]*movie.Genre, error) {
	logger := r.logger.With(applog.String("method", "FindOrCreateByNames"), applog.Any("genre_names", names))

	var genreGorms []*models.GenreGorm
	// 先尝试查找所有已存在的类型
	if err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&genreGorms).Error; err != nil {
		logger.Error("failed to find genres by names", applog.Error(err))
		return nil, fmt.Errorf("failed to find genres by names: %w", err)
	}

	// 找出需要创建的类型名称
	existingNames := make(map[string]bool)
	for _, g := range genreGorms {
		existingNames[g.Name] = true
	}

	var newNames []string
	for _, name := range names {
		if !existingNames[name] {
			newNames = append(newNames, name)
		}
	}

	// 批量创建不存在的类型
	if len(newNames) > 0 {
		newGenres := make([]*models.GenreGorm, len(newNames))
		for i, name := range newNames {
			newGenres[i] = &models.GenreGorm{Name: name}
		}
		if err := r.db.WithContext(ctx).Create(&newGenres).Error; err != nil {
			logger.Error("failed to create new genres", applog.Error(err))
			return nil, fmt.Errorf("failed to create new genres: %w", err)
		}
		genreGorms = append(genreGorms, newGenres...)
	}

	// 转换为领域模型
	genres := make([]*movie.Genre, len(genreGorms))
	for i, g := range genreGorms {
		genres[i] = g.ToDomain()
	}

	logger.Info("find or create genres successfully", applog.Int("total_count", len(genres)))
	return genres, nil
}
