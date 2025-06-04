package app

import (
	"context"
	"errors"
	"fmt"
	"math"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"

	"mrs/internal/domain/movie"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/shared/vo"
	"mrs/internal/infrastructure/cache"
	applog "mrs/pkg/log"
)

type MovieService struct {
	uow        shared.UnitOfWork
	movieRepo  movie.MovieRepository
	genreRepo  movie.GenreRepository
	movieCache cache.MovieCache
	logger     applog.Logger
}

func NewMovieService(
	uow shared.UnitOfWork,
	movieRepo movie.MovieRepository,
	movieCache cache.MovieCache,
	logger applog.Logger,
) *MovieService {
	return &MovieService{
		uow:        uow,
		movieRepo:  movieRepo,
		movieCache: movieCache,
		logger:     logger.With(applog.String("Component", "MovieService")),
	}
}

// 创建电影
func (s *MovieService) CreateMovie(ctx context.Context,
	req *request.CreateMovieRequest) (*response.MovieResponse, error) {
	logger := s.logger.With(applog.String("Method", "CreateMovie"))

	mv := req.ToMovie()

	// 开启事务
	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		// 检查并创建电影类型
		genres, err := provider.GetGenreRepository().FindOrCreateByNames(ctx, req.GenreNames)
		if err != nil {
			logger.Error("failed to get genres", applog.Error(err))
			return fmt.Errorf("failed to get or create genres: %w", err)
		}

		// 创建电影
		movieRepo := provider.GetMovieRepository()
		mv, err = movieRepo.Create(ctx, mv)
		if err != nil {
			logger.Error("failed to create movie", applog.Error(err))
			return fmt.Errorf("failed to create movie: %w", err)
		}

		// 电影创建时并不会自动关联类型，需要手动替换
		if err := movieRepo.ReplaceGenresForMovie(ctx, mv, genres); err != nil {
			logger.Error("failed to replace genres for movie", applog.Error(err))
			return fmt.Errorf("failed to replace genres for movie: %w", err)
		}
		return nil
	})

	if err != nil {
		logger.Error("failed to create movie", applog.Error(err))
		return nil, fmt.Errorf("failed to create movie: %w", err)
	}

	if err := s.movieCache.SetMovie(ctx, mv, 0); err != nil {
		logger.Error("failed to set movie to cache", applog.Error(err))
	}

	logger.Info("create movie successfully", applog.Uint("movie_id", uint(mv.ID)))
	return response.ToMovieResponse(mv), nil
}

// 更新电影
func (s *MovieService) UpdateMovie(ctx context.Context, req *request.UpdateMovieRequest) error {
	logger := s.logger.With(applog.String("Method", "UpdateMovie"))

	mv, err := s.movieRepo.FindByTitle(ctx, req.Title)
	tag := false
	if err != nil {
		logger.Error("failed to get movie", applog.Error(err))
		return err
	}

	if req.Description != "" {
		mv.Description = req.Description
		tag = true
	}

	if !req.ReleaseDate.IsZero() {
		mv.ReleaseDate = req.ReleaseDate
		tag = true
	}

	if req.DurationMinutes != 0 {
		mv.DurationMinutes = req.DurationMinutes
		tag = true
	}

	if req.Rating != 0 {
		mv.Rating = float32(req.Rating)
		tag = true
	}

	if req.PosterURL != "" {
		mv.PosterURL = req.PosterURL
		tag = true
	}

	if req.AgeRating != "" {
		mv.AgeRating = req.AgeRating
		tag = true
	}

	if req.Cast != "" {
		mv.Cast = req.Cast
		tag = true
	}

	if len(req.GenreNames) > 0 {
		err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
			genres, err := provider.GetGenreRepository().FindOrCreateByNames(ctx, req.GenreNames)
			if err != nil {
				logger.Error("failed to get genres", applog.Error(err))
				return fmt.Errorf("failed to get genres: %w", err)
			}
			if err := provider.GetMovieRepository().ReplaceGenresForMovie(ctx, mv, genres); err != nil {
				logger.Error("failed to replace genres for movie", applog.Error(err))
				return fmt.Errorf("failed to replace genres for movie: %w", err)
			}
			return nil
		})
		if err != nil {
			logger.Error("failed to get genres", applog.Error(err))
			return fmt.Errorf("failed to get genres: %w", err)
		}
	}

	if tag {
		// 更新单条记录且不涉及多表，无需事务
		if err := s.movieRepo.Update(ctx, mv); err != nil {
			logger.Error("failed to update movie", applog.Error(err))
			return fmt.Errorf("failed to update movie: %w", err)
		}
	}

	if err := s.movieCache.SetMovie(ctx, mv, 0); err != nil {
		logger.Error("failed to set movie to cache", applog.Error(err))
	}

	logger.Info("update movie successfully", applog.Uint("movie_id", uint(mv.ID)))
	return nil
}

// 获取电影详情
func (s *MovieService) GetMovieByID(ctx context.Context, id uint) (*response.MovieResponse, error) {
	logger := s.logger.With(applog.String("Method", "GetMovieByID"), applog.Uint("movie_id", id))

	mv, err := s.movieCache.GetMovie(ctx, id)
	if err == nil {
		logger.Info("get movie from cache successfully")
		return response.ToMovieResponse(mv), nil
	}

	mv, err = s.movieRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, movie.ErrMovieNotFound) {
			logger.Info("movie not found")
			// 构造一个空的movie并设置到缓存中,防止缓存穿透
			emptyMovie := &movie.Movie{ID: vo.MovieID(id)}
			if err := s.movieCache.SetMovie(ctx, emptyMovie, 0); err != nil {
				logger.Error("failed to set empty movie to cache", applog.Error(err))
			}
			logger.Info("set empty movie to cache successfully")
			return nil, fmt.Errorf("%w: %w", movie.ErrMovieNotFound, err)
		}
		logger.Error("failed to get movie", applog.Error(err))
		return nil, fmt.Errorf("failed to get movie: %w", err)
	}

	if err := s.movieCache.SetMovie(ctx, mv, 0); err != nil {
		logger.Error("failed to set movie to cache", applog.Error(err))
	}

	logger.Info("get movie by id successfully")
	return response.ToMovieResponse(mv), nil
}

// 删除电影
func (s *MovieService) DeleteMovie(ctx context.Context, id uint) error {
	logger := s.logger.With(applog.String("Method", "DeleteMovie"), applog.Uint("movie_id", id))

	if err := s.movieRepo.Delete(ctx, id); err != nil {
		logger.Error("failed to delete movie", applog.Error(err))
		return fmt.Errorf("failed to delete movie: %w", err)
	}

	if err := s.movieCache.DeleteMovie(ctx, id); err != nil {
		logger.Error("failed to delete movie from cache", applog.Error(err))
	}

	logger.Info("delete movie successfully")
	return nil
}

// 获取电影列表
func (s *MovieService) ListMovies(ctx context.Context, req *request.ListMovieRequest) (*response.PaginatedMovieResponse, error) {
	logger := s.logger.With(applog.String("Method", "ListMovies"))

	filters := make(map[string]interface{})
	if req.Title != "" {
		filters["title"] = req.Title
	}

	if req.GenreName != "" {
		filters["genre_name"] = req.GenreName
	}

	if req.ReleaseYear != 0 {
		filters["release_year"] = req.ReleaseYear
	}

	if req.SortBy != "" {
		filters["sort_by"] = req.SortBy
	}

	if req.SortOrder != "" {
		filters["sort_order"] = req.SortOrder
	}

	movies, total, err := s.movieRepo.List(ctx, req.Page, req.PageSize, filters)
	if err != nil {
		logger.Error("failed to list movies", applog.Error(err))
		return nil, fmt.Errorf("failed to list movies: %w", err)
	}

	logger.Info("list movies successfully", applog.Int("total", int(total)))

	simpleMovies := make([]*response.MovieSimpleResponse, 0, len(movies))
	for _, movie := range movies {
		genreNames := make([]string, 0, len(movie.Genres))
		for _, genre := range movie.Genres {
			genreNames = append(genreNames, genre.Name)
		}
		simpleMovies = append(simpleMovies, &response.MovieSimpleResponse{
			ID:          uint(movie.ID),
			Title:       movie.Title,
			ReleaseDate: movie.ReleaseDate,
			Rating:      float64(movie.Rating),
			PosterURL:   movie.PosterURL,
			AgeRating:   movie.AgeRating,
			GenreNames:  genreNames,
		})
	}

	pagination := response.PaginationResponse{
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalCount: int(total),
		TotalPages: int(math.Ceil(float64(total) / float64(req.PageSize))),
	}

	logger.Info("list movies successfully",
		applog.Int("total", int(total)),
		applog.Int("page", req.Page),
		applog.Int("page_size", req.PageSize),
	)

	if err := s.movieCache.SetMovieList(ctx, movies, filters, 0); err != nil {
		logger.Error("failed to set movie list to cache", applog.Error(err))
	}

	return &response.PaginatedMovieResponse{
		Pagination: pagination,
		Movies:     simpleMovies,
	}, nil
}
