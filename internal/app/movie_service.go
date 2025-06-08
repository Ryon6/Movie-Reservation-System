// TODO: 添加排序功能. 修改movie字段
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
	applog "mrs/pkg/log"
)

type MovieService interface {
	CreateMovie(ctx context.Context, req *request.CreateMovieRequest) (*response.MovieResponse, error)
	UpdateMovie(ctx context.Context, req *request.UpdateMovieRequest) error
	GetMovie(ctx context.Context, req *request.GetMovieRequest) (*response.MovieResponse, error)
	DeleteMovie(ctx context.Context, req *request.DeleteMovieRequest) error
	ListMovies(ctx context.Context, req *request.ListMovieRequest) (*response.PaginatedMovieResponse, error)
	CreateGenre(ctx context.Context, req *request.CreateGenreRequest) (*response.GenreResponse, error)
	ListAllGenres(ctx context.Context) (*response.ListAllGenresResponse, error)
	UpdateGenre(ctx context.Context, req *request.UpdateGenreRequest) (*response.GenreResponse, error)
	DeleteGenre(ctx context.Context, req *request.DeleteGenreRequest) error
}

type movieService struct {
	uow        shared.UnitOfWork
	movieRepo  movie.MovieRepository
	genreRepo  movie.GenreRepository
	movieCache movie.MovieCache
	logger     applog.Logger
}

func NewMovieService(
	uow shared.UnitOfWork,
	movieRepo movie.MovieRepository,
	genreRepo movie.GenreRepository,
	movieCache movie.MovieCache,
	logger applog.Logger,
) MovieService {
	return &movieService{
		uow:        uow,
		movieRepo:  movieRepo,
		genreRepo:  genreRepo,
		movieCache: movieCache,
		logger:     logger.With(applog.String("Service", "MovieService")),
	}
}

// 创建电影
func (s *movieService) CreateMovie(ctx context.Context,
	req *request.CreateMovieRequest) (*response.MovieResponse, error) {
	logger := s.logger.With(applog.String("Method", "CreateMovie"))

	mv := req.ToMovie()

	// 开启事务
	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		// 创建电影
		var err error
		movieRepo := provider.GetMovieRepository()
		mv, err = movieRepo.Create(ctx, mv)
		if err != nil {
			logger.Error("failed to create movie", applog.Error(err))
			return fmt.Errorf("failed to create movie: %w", err)
		}

		// 检查并创建电影类型
		genres, err := provider.GetGenreRepository().FindOrCreateByNames(ctx, req.GenreNames)
		if err != nil {
			logger.Error("failed to get genres", applog.Error(err))
			return fmt.Errorf("failed to get or create genres: %w", err)
		}
		mv.Genres = genres

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
func (s *movieService) UpdateMovie(ctx context.Context, req *request.UpdateMovieRequest) error {
	logger := s.logger.With(applog.String("Method", "UpdateMovie"))

	// 转换为领域对象, gorm.Updates时只更新非空字段
	mv := req.ToDomain()
	// 是否需要更新其他字段
	hasOtherUpdate := req.Title == "" &&
		req.Description == "" &&
		req.ReleaseDate.IsZero() &&
		req.DurationMinutes == 0 &&
		req.Rating == 0 &&
		req.PosterURL == "" &&
		req.AgeRating == "" &&
		req.Cast == ""

	// 根据请求内容，存在是否更新类型字段与是否更新其他字段等四种情况
	if !hasOtherUpdate && len(req.GenreNames) == 0 {
		logger.Info("no update")
		return nil
	}

	// 如果有类型字段更新,先更新类型
	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		movieRepo := provider.GetMovieRepository()

		// 如果有类型更新
		if len(req.GenreNames) > 0 {
			// 检查并创建电影类型
			genres, err := provider.GetGenreRepository().FindOrCreateByNames(ctx, req.GenreNames)
			if err != nil {
				logger.Error("failed to get or create genres", applog.Error(err))
				return fmt.Errorf("failed to get or create genres: %w", err)
			}

			// 更新电影类型关联
			if err := movieRepo.ReplaceGenresForMovie(ctx, mv, genres); err != nil {
				logger.Error("failed to replace genres for movie", applog.Error(err))
				return fmt.Errorf("failed to replace genres for movie: %w", err)
			}

			// 如果只更新类型,则直接返回
			if !hasOtherUpdate {
				return nil
			}
		}

		// 更新电影其他字段
		if err := movieRepo.Update(ctx, mv); err != nil {
			logger.Error("failed to update movie", applog.Error(err))
			return fmt.Errorf("failed to update movie: %w", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to update movie", applog.Error(err))
		return fmt.Errorf("failed to update movie: %w", err)
	}

	if err := s.movieCache.SetMovie(ctx, mv, 0); err != nil {
		logger.Error("failed to set movie to cache", applog.Error(err))
	}

	logger.Info("update movie successfully", applog.Uint("movie_id", uint(mv.ID)))
	return nil
}

// 获取电影详情
func (s *movieService) GetMovie(ctx context.Context, req *request.GetMovieRequest) (*response.MovieResponse, error) {
	logger := s.logger.With(applog.String("Method", "GetMovie"), applog.Uint("movie_id", req.ID))

	mv, err := s.movieCache.GetMovie(ctx, req.ID)
	if err == nil {
		logger.Info("get movie from cache successfully")
		return response.ToMovieResponse(mv), nil
	}

	mv, err = s.movieRepo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, movie.ErrMovieNotFound) {
			logger.Info("movie not found")
			// 构造一个空的movie并设置到缓存中,防止缓存穿透
			emptyMovie := &movie.Movie{ID: vo.MovieID(req.ID)}
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
func (s *movieService) DeleteMovie(ctx context.Context, req *request.DeleteMovieRequest) error {
	logger := s.logger.With(applog.String("Method", "DeleteMovie"), applog.Uint("movie_id", req.ID))

	if err := s.movieRepo.Delete(ctx, req.ID); err != nil {
		logger.Error("failed to delete movie", applog.Error(err))
		return fmt.Errorf("failed to delete movie: %w", err)
	}

	if err := s.movieCache.DeleteMovie(ctx, req.ID); err != nil {
		logger.Error("failed to delete movie from cache", applog.Error(err))
	}

	logger.Info("delete movie successfully")
	return nil
}

// 获取电影列表
func (s *movieService) ListMovies(ctx context.Context, req *request.ListMovieRequest) (*response.PaginatedMovieResponse, error) {
	logger := s.logger.With(applog.String("Method", "ListMovies"))

	options := req.ToDomain()

	var movies []*movie.Movie
	// 分页函数，用于处理缓存命中和未命中两种情况
	var fn = func(movies []*movie.Movie) *response.PaginatedMovieResponse {
		total := len(movies)
		startIndex := (req.Page - 1) * req.PageSize
		endIndex := min(startIndex+req.PageSize, total)
		movies = movies[startIndex:endIndex]
		moviesResponse := make([]*response.MovieSimpleResponse, 0, len(movies))
		for _, movie := range movies {
			moviesResponse = append(moviesResponse, response.ToMovieSimpleResponse(movie))
		}
		return &response.PaginatedMovieResponse{
			Pagination: response.PaginationResponse{
				Page:       req.Page,
				PageSize:   req.PageSize,
				TotalCount: int(total),
				TotalPages: int(math.Ceil(float64(total) / float64(req.PageSize))),
			},
			Movies: moviesResponse,
		}
	}

	// 列表缓存命中(err == nil)有三种情况：
	// 1. 列表缓存为空（即无数据对应过滤条件） -> 直接返回空列表
	// 2. 列表缓存中存在数据，但部分movie记录缺失（即movie_id列表中存在但缓存中不存在） -> 进一步查询数据库
	// 3. 列表缓存中存在数据，且所有movie记录都存在（即movie_id列表中所有movie记录都存在） -> 直接返回缓存数据
	// 而列表缓存未命中，则需要进一步查询数据库
	cacheResult, err := s.movieCache.GetMovieList(ctx, options)
	if err != nil {
		logger.Warn("failed to get movie list from cache", applog.Error(err))
	} else {
		logger.Info("get movie list from cache successfully")
		if len(cacheResult.MissingMovieIDs) == 0 {
			logger.Info("all movies found in cache", applog.Int("total", len(cacheResult.Movies)))
			return fn(cacheResult.Movies), nil
		} else {
			logger.Info("some movies not found in cache", applog.Int("missing", len(cacheResult.MissingMovieIDs)))
			movies = cacheResult.Movies
		}
	}

	if len(movies) != 0 {
		missingMovies, err := s.movieRepo.FindByIDs(ctx, cacheResult.MissingMovieIDs)
		if err != nil {
			logger.Error("failed to find missing movies", applog.Error(err))
			return nil, err
		}
		movies = append(movies, missingMovies...)
	} else {
		movies, _, err = s.movieRepo.List(ctx, options)
		if err != nil {
			logger.Error("failed to list movies", applog.Error(err))
			return nil, err
		}
	}

	logger.Info("list movies successfully", applog.Int("total", len(movies)))

	if err := s.movieCache.SetMovieList(ctx, movies, options, 0); err != nil {
		logger.Error("failed to set movie list to cache", applog.Error(err))
	}
	return fn(movies), nil
}

// 创建类型
func (s *movieService) CreateGenre(ctx context.Context, req *request.CreateGenreRequest) (*response.GenreResponse, error) {
	logger := s.logger.With(applog.String("Method", "CreateGenre"))

	genre := &movie.Genre{
		Name: req.Name,
	}
	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		var err error
		genre, err = provider.GetGenreRepository().Create(ctx, genre)
		if err != nil {
			logger.Error("failed to create genre", applog.Error(err))
			return fmt.Errorf("failed to create genre: %w", err)
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to create genre", applog.Error(err))
		return nil, fmt.Errorf("failed to create genre: %w", err)
	}

	logger.Info("create genre successfully", applog.Uint("genre_id", uint(genre.ID)))
	return response.ToGenreResponse(genre), nil
}

// 更新类型
func (s *movieService) UpdateGenre(ctx context.Context, req *request.UpdateGenreRequest) (*response.GenreResponse, error) {
	logger := s.logger.With(applog.String("Method", "UpdateGenre"))

	genre, err := s.genreRepo.FindByID(ctx, req.ID)
	if err != nil {
		logger.Error("failed to get genre", applog.Error(err))
		return nil, fmt.Errorf("failed to get genre: %w", err)
	}

	if req.Name == "" || req.Name == genre.Name {
		logger.Info("genre name is the same as the original name")
		return response.ToGenreResponse(genre), nil
	}

	genre.Name = req.Name
	if err := s.genreRepo.Update(ctx, genre); err != nil {
		logger.Error("failed to update genre", applog.Error(err))
		return nil, fmt.Errorf("failed to update genre: %w", err)
	}

	logger.Info("update genre successfully", applog.Uint("genre_id", uint(genre.ID)))
	return response.ToGenreResponse(genre), nil
}

// 删除类型
func (s *movieService) DeleteGenre(ctx context.Context, req *request.DeleteGenreRequest) error {
	logger := s.logger.With(applog.String("Method", "DeleteGenre"), applog.Uint("genre_id", req.ID))

	if err := s.genreRepo.Delete(ctx, req.ID); err != nil {
		if errors.Is(err, movie.ErrGenreReferenced) {
			logger.Warn("genre is referenced by movie", applog.Uint("genre_id", req.ID))
			return fmt.Errorf("genre is referenced by movie: %w", err)
		}
		logger.Error("failed to delete genre", applog.Error(err))
		return fmt.Errorf("failed to delete genre: %w", err)
	}

	logger.Info("delete genre successfully", applog.Uint("genre_id", req.ID))
	return nil
}

// 获取类型列表
func (s *movieService) ListAllGenres(ctx context.Context) (*response.ListAllGenresResponse, error) {
	logger := s.logger.With(applog.String("Method", "ListGenres"))

	genres, err := s.genreRepo.ListAll(ctx)
	if err != nil {
		logger.Error("failed to list genres", applog.Error(err))
		return nil, fmt.Errorf("failed to list genres: %w", err)
	}

	logger.Info("list genres successfully", applog.Int("total", len(genres)))
	return response.ToListAllGenresResponse(genres), nil
}
