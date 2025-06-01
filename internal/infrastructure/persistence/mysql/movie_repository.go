package mysql

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/domain/movie"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type gormMovieRepository struct {
	db     *gorm.DB
	logger applog.Logger
}

func NewGormMovieRepository(db *gorm.DB, logger applog.Logger) movie.MovieRepository {
	return &gormMovieRepository{
		db:     db,
		logger: logger,
	}
}

func (r *gormMovieRepository) Create(ctx context.Context, mv *movie.Movie) error {
	logger := r.logger.With(applog.String("Method", "gormMovieRepository.Create"),
		applog.Uint("movie_id", mv.ID), applog.String("title", mv.Title))
	if err := r.db.WithContext(ctx).Create(mv).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			logger.Warn("movie already eixsts", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrMovieAlreadyExists, err)
		}
		logger.Error("failed to create movie", applog.Error(err))
		return fmt.Errorf("failed to create movie: %w", err)

	}
	logger.Info("create movie successfully")
	return nil
}

// 应支持预加载Genres
func (r *gormMovieRepository) FindByID(ctx context.Context, id uint) (*movie.Movie, error) {
	logger := r.logger.With(applog.String("Method", "gormMovieRepository.FindByID"), applog.Uint("movie_id", id))
	var mv movie.Movie
	if err := r.db.WithContext(ctx).Preload("Genres").First(&mv, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("movie id not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %w", movie.ErrMovieNotFound, err)
		}
		logger.Error("failed to find movie by id", applog.Error(err))
		return nil, fmt.Errorf("failed to find movie by id: %w", err)
	}
	logger.Info("find movie by id successfully")
	return &mv, nil
}

// List 从数据库中获取电影列表，支持分页和过滤。
func (r *gormMovieRepository) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*movie.Movie, int64, error) {
	logger := r.logger.With(applog.String("method", "List"), applog.Int("page", page), applog.Int("pageSize", pageSize))
	var movies []*movie.Movie
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&movie.Movie{})
	countQuery := r.db.WithContext(ctx).Model(&movie.Movie{}) // 用于计数的独立查询

	// 应用过滤器
	if title, ok := filters["title"].(string); ok && title != "" {
		searchTerm := fmt.Sprintf("%%%s%%", title) // 模糊搜索
		query = query.Where("title LIKE ?", searchTerm)
		countQuery = countQuery.Where("title LIKE ?", searchTerm)
		logger = logger.With(applog.String("filter_title", title))
	}

	if genreID, ok := filters["genre_id"].(uint); ok && genreID > 0 {
		// 通过 movie_genres 连接表进行过滤
		query = query.Joins("JOIN movie_genres ON movie_genres.movie_id = movies.id").Where("movie_genres.genre_id = ?", genreID)
		countQuery = countQuery.Joins("JOIN movie_genres ON movie_genres.movie_id = movies.id").Where("movie_genres.genre_id = ?", genreID)
		logger = logger.With(applog.Uint("filter_genre_id", genreID))
	}
	// 可以添加更多过滤器，例如按上映日期范围等

	// 获取总数
	if err := countQuery.Count(&totalCount).Error; err != nil {
		logger.Error("Failed to count movies", applog.Error(err))
		return nil, 0, err
	}

	if totalCount == 0 {
		logger.Info("No movies found matching criteria")
		return movies, 0, nil // 返回空列表和0计数
	}

	// 应用排序和分页，并预加载类型
	offset := (page - 1) * pageSize
	if err := query.Order("release_date DESC, title ASC").Offset(offset).Limit(pageSize).Preload("Genres").Find(&movies).Error; err != nil {
		logger.Error("Failed to list movies", applog.Error(err))
		return nil, 0, err
	}

	logger.Info("Movies listed successfully", applog.Int("count", len(movies)), applog.Int64("total_count", totalCount))
	return movies, totalCount, nil
}

func (r *gormMovieRepository) Update(ctx context.Context, mv *movie.Movie) error {
	logger := r.logger.With(applog.String("Method", "gormMovieRepository.Update"),
		applog.Uint("movie_id", mv.ID), applog.String("title", mv.Title))

	if err := r.db.WithContext(ctx).First(&movie.Movie{}, mv.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("movie not found", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrMovieNotFound, err)
		}
		logger.Error("failed to find movie by id", applog.Error(err))
		return fmt.Errorf("failed to find movie by id: %w", err)
	}

	result := r.db.WithContext(ctx).Model(&movie.Movie{}).Where("id = ?", mv.ID).Updates(mv)
	if err := result.Error; err != nil {
		logger.Error("failed to update movie", applog.Error(err))
		return fmt.Errorf("failed to update movie: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("no rows affected during update")
		return movie.ErrNoRowsAffected
	}

	logger.Info("update movie successfully")
	return nil
}

func (r *gormMovieRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "gormMovieRepository.Delete"), applog.Uint("movie_id", id))

	result := r.db.WithContext(ctx).Delete(&movie.Movie{}, id)
	if err := result.Error; err != nil {
		logger.Error("failed to delete movie", applog.Error(err))
		return fmt.Errorf("failed to delete movie: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("movie to delete not found or already deleted")
		return movie.ErrNoRowsAffected
	}

	logger.Info("delete movie successfully")
	return nil
}

// 为电影增加、删除和修改类型
func (r *gormMovieRepository) AddGenreToMovie(ctx context.Context, mv *movie.Movie, genre *movie.Genre) error {
	logger := r.logger.With(applog.String("Method", "gormMovieRepository.AddGenreToMovie"),
		applog.Uint("movie_id", mv.ID), applog.String("movie_title", mv.Title),
		applog.Uint("genre_id", genre.ID), applog.String("genre_name", genre.Name))
	if err := r.db.WithContext(ctx).Model(mv).Association("Genres").Append(genre); err != nil {
		logger.Error("failed to add genre to movie", applog.Error(err))
		return fmt.Errorf("failed to add genre to movie: %w", err)
	}

	logger.Info("add genre to movie successfully")
	return nil
}

func (r *gormMovieRepository) RemoveGenreToMovie(ctx context.Context, mv *movie.Movie, genre *movie.Genre) error {
	logger := r.logger.With(applog.String("Method", "gormMovieRepository.RemoveGenreToMovie"),
		applog.Uint("movie_id", mv.ID), applog.String("movie_title", mv.Title),
		applog.Uint("genre_id", genre.ID), applog.String("genre_name", genre.Name))

	if err := r.db.WithContext(ctx).Model(mv).Association("Genres").Delete(genre); err != nil {
		logger.Error("failed to remove genre to movie", applog.Error(err))
		return fmt.Errorf("failed to remove genre to movie: %w", err)
	}
	logger.Info("remove genre to movie successfully")
	return nil

}

func (r *gormMovieRepository) ReplaceGenresForMovie(ctx context.Context, mv *movie.Movie, genres []*movie.Genre) error {
	logger := r.logger.With(applog.String("Method", "gormMovieRepository.ReplaceGenreForMovie"),
		applog.Uint("movie_id", mv.ID), applog.String("movie_title", mv.Title))

	if err := r.db.WithContext(ctx).Model(mv).Association("Genres").Replace(genres); err != nil {
		logger.Error("failed to replace genre to movie", applog.Error(err))
		return fmt.Errorf("failed to replace genre to movie: %w", err)
	}
	logger.Info("replace genre to movie successfully")
	return nil
}
