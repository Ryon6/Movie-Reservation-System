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

type gormMovieRepository struct {
	db     *gorm.DB
	logger applog.Logger
}

func NewGormMovieRepository(db *gorm.DB, logger applog.Logger) movie.MovieRepository {
	return &gormMovieRepository{
		db:     db,
		logger: logger.With(applog.String("Repository", "gormMovieRepository")),
	}
}

func (r *gormMovieRepository) Create(ctx context.Context, mv *movie.Movie) (*movie.Movie, error) {
	logger := r.logger.With(applog.String("Method", "Create"),
		applog.Uint("movie_id", uint(mv.ID)), applog.String("title", mv.Title))
	movieGorm := models.MovieGormFromDomain(mv)
	if err := r.db.WithContext(ctx).Create(movieGorm).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			logger.Warn("movie already eixsts", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", movie.ErrMovieAlreadyExists, err)
		}
		logger.Error("failed to create movie", applog.Error(err))
		return nil, fmt.Errorf("failed to create movie: %w", err)

	}
	logger.Info("create movie successfully")
	return movieGorm.ToDomain(), nil
}

// 应支持预加载Genres
func (r *gormMovieRepository) FindByID(ctx context.Context, id uint) (*movie.Movie, error) {
	logger := r.logger.With(applog.String("Method", "FindByID"), applog.Uint("movie_id", id))
	var mvGorm models.MovieGorm
	if err := r.db.WithContext(ctx).Preload("Genres").First(&mvGorm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("movie id not found", applog.Error(err))
			return nil, fmt.Errorf("%w(id): %w", movie.ErrMovieNotFound, err)
		}
		logger.Error("failed to find movie by id", applog.Error(err))
		return nil, fmt.Errorf("failed to find movie by id: %w", err)
	}
	logger.Info("find movie by id successfully")
	return mvGorm.ToDomain(), nil
}

func (r *gormMovieRepository) FindByIDs(ctx context.Context, ids []uint) ([]*movie.Movie, error) {
	logger := r.logger.With(applog.String("Method", "FindByIDs"), applog.Int("count", len(ids)))
	var mvGorms []*models.MovieGorm
	if err := r.db.WithContext(ctx).Preload("Genres").Where("id IN (?)", ids).Find(&mvGorms).Error; err != nil {
		logger.Error("failed to find movies by ids", applog.Error(err))
		return nil, fmt.Errorf("failed to find movies by ids: %w", err)
	}
	movies := make([]*movie.Movie, len(mvGorms))
	for i, mvGorm := range mvGorms {
		movies[i] = mvGorm.ToDomain()
	}
	logger.Info("find movies by ids successfully")
	return movies, nil
}

// 应支持预加载Genres
func (r *gormMovieRepository) FindByTitle(ctx context.Context, title string) (*movie.Movie, error) {
	logger := r.logger.With(applog.String("Method", "FindByTitle"), applog.String("title", title))
	var mvGorm models.MovieGorm
	if err := r.db.WithContext(ctx).Preload("Genres").Where("title = ?", title).First(&mvGorm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("movie title not found", applog.Error(err))
			return nil, fmt.Errorf("%w(title): %w", movie.ErrMovieNotFound, err)
		}
		logger.Error("failed to find movie by title", applog.Error(err))
		return nil, fmt.Errorf("failed to find movie by title: %w", err)
	}
	logger.Info("find movie by title successfully")
	return mvGorm.ToDomain(), nil
}

// List 从数据库中获取电影列表，支持分页和过滤。
func (r *gormMovieRepository) List(ctx context.Context, options *movie.MovieQueryOptions) ([]*movie.Movie, int64, error) {
	logger := r.logger.With(applog.String("Method", "List"), applog.Int("page", options.Page), applog.Int("pageSize", options.PageSize))
	var moviesGorms []*models.MovieGorm
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&models.MovieGorm{})
	countQuery := r.db.WithContext(ctx).Model(&models.MovieGorm{}) // 用于计数的独立查询

	// 标题过滤（模糊查询）
	if options.Title != "" {
		searchTerm := fmt.Sprintf("%%%s%%", options.Title) // 模糊搜索
		query = query.Where("title LIKE ?", searchTerm)
		countQuery = countQuery.Where("title LIKE ?", searchTerm)
		logger = logger.With(applog.String("query_title", options.Title))
	}

	// 上映年份过滤
	if options.ReleaseYear != 0 {
		query = query.Where("YEAR(release_date) = ?", options.ReleaseYear)
		countQuery = countQuery.Where("YEAR(release_date) = ?", options.ReleaseYear)
		logger = logger.With(applog.Int("query_release_year", options.ReleaseYear))
	}

	// 类型过滤
	if options.GenreName != "" {
		query = query.Joins("JOIN movies_genres ON movies_genres.movie_id = movies.id").
			Joins("JOIN genres ON genres.id = movies_genres.genre_id").
			Where("genres.name = ?", options.GenreName)
		countQuery = countQuery.Joins("JOIN movies_genres ON movies_genres.movie_id = movies.id").
			Joins("JOIN genres ON genres.id = movies_genres.genre_id").
			Where("genres.name = ?", options.GenreName)
		logger = logger.With(applog.String("query_genre", options.GenreName))
	}
	// 获取总数
	if err := countQuery.Count(&totalCount).Error; err != nil {
		logger.Error("Failed to count movies", applog.Error(err))
		return nil, 0, err
	}

	if totalCount == 0 {
		logger.Info("No movies found matching criteria")
		return nil, 0, nil // 返回空列表和0计数
	}

	// 应用排序和分页，并预加载类型
	offset := (options.Page - 1) * options.PageSize
	if err := query.Order("release_date DESC, title ASC").Offset(offset).Limit(options.PageSize).Preload("Genres").Find(&moviesGorms).Error; err != nil {
		logger.Error("Failed to list movies", applog.Error(err))
		return nil, 0, err
	}

	logger.Info("Movies listed successfully", applog.Int("count", len(moviesGorms)), applog.Int64("total_count", totalCount))
	movies := make([]*movie.Movie, len(moviesGorms))
	for i, movieGorm := range moviesGorms {
		movies[i] = movieGorm.ToDomain()
	}
	return movies, totalCount, nil
}

func (r *gormMovieRepository) Update(ctx context.Context, mv *movie.Movie) error {
	logger := r.logger.With(applog.String("Method", "Update"),
		applog.Uint("movie_id", uint(mv.ID)), applog.String("title", mv.Title))

	movieGorm := models.MovieGormFromDomain(mv)
	if err := r.db.WithContext(ctx).First(&models.MovieGorm{}, movieGorm.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("movie not found", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrMovieNotFound, err)
		}
		logger.Error("failed to find movie by id", applog.Error(err))
		return fmt.Errorf("failed to find movie by id: %w", err)
	}

	result := r.db.WithContext(ctx).Model(&models.MovieGorm{}).Where("id = ?", movieGorm.ID).Updates(movieGorm)
	if err := result.Error; err != nil {
		logger.Error("failed to update movie", applog.Error(err))
		return fmt.Errorf("failed to update movie: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("no rows affected during update")
		return shared.ErrNoRowsAffected
	}

	logger.Info("update movie successfully")
	return nil
}

func (r *gormMovieRepository) Delete(ctx context.Context, id uint) error {
	logger := r.logger.With(applog.String("Method", "Delete"), applog.Uint("movie_id", id))

	result := r.db.WithContext(ctx).Delete(&models.MovieGorm{}, id)
	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("movie to delete not found", applog.Error(err))
			return fmt.Errorf("%w: %w", movie.ErrMovieNotFound, err)
		}
		logger.Error("failed to delete movie", applog.Error(err))
		return fmt.Errorf("failed to delete movie: %w", err)
	}

	if result.RowsAffected == 0 {
		logger.Warn("movie to delete not found or already deleted")
		return shared.ErrNoRowsAffected
	}

	logger.Info("delete movie successfully")
	return nil
}

// 为电影增加、删除和修改类型
func (r *gormMovieRepository) AddGenreToMovie(ctx context.Context, mv *movie.Movie, genre *movie.Genre) error {
	logger := r.logger.With(applog.String("Method", "AddGenreToMovie"),
		applog.Uint("movie_id", uint(mv.ID)), applog.String("movie_title", mv.Title),
		applog.Uint("genre_id", uint(genre.ID)), applog.String("genre_name", genre.Name))
	movieGorm := models.MovieGormFromDomain(mv)
	genreGorm := models.GenreGormFromDomain(genre)
	if err := r.db.WithContext(ctx).Model(movieGorm).Association("Genres").Append(genreGorm); err != nil {
		logger.Error("failed to add genre to movie", applog.Error(err))
		return fmt.Errorf("failed to add genre to movie: %w", err)
	}

	logger.Info("add genre to movie successfully")
	return nil
}

func (r *gormMovieRepository) RemoveGenreToMovie(ctx context.Context, mv *movie.Movie, genre *movie.Genre) error {
	logger := r.logger.With(applog.String("Method", "RemoveGenreToMovie"),
		applog.Uint("movie_id", uint(mv.ID)), applog.String("movie_title", mv.Title),
		applog.Uint("genre_id", uint(genre.ID)), applog.String("genre_name", genre.Name))

	movieGorm := models.MovieGormFromDomain(mv)
	genreGorm := models.GenreGormFromDomain(genre)
	if err := r.db.WithContext(ctx).Model(movieGorm).Association("Genres").Delete(genreGorm); err != nil {
		logger.Error("failed to remove genre to movie", applog.Error(err))
		return fmt.Errorf("failed to remove genre to movie: %w", err)
	}
	logger.Info("remove genre to movie successfully")
	return nil

}

func (r *gormMovieRepository) ReplaceGenresForMovie(ctx context.Context, mv *movie.Movie, genres []*movie.Genre) error {
	logger := r.logger.With(applog.String("Method", "ReplaceGenresForMovie"),
		applog.Uint("movie_id", uint(mv.ID)), applog.String("movie_title", mv.Title))

	genresGorms := make([]*models.GenreGorm, len(genres))
	for i, genre := range genres {
		genresGorms[i] = models.GenreGormFromDomain(genre)
	}
	movieGorm := models.MovieGormFromDomain(mv)
	if err := r.db.WithContext(ctx).Model(movieGorm).Association("Genres").Replace(genresGorms); err != nil {
		logger.Error("failed to replace genre to movie", applog.Error(err))
		return fmt.Errorf("failed to replace genre to movie: %w", err)
	}
	logger.Info("replace genre to movie successfully")
	return nil
}
