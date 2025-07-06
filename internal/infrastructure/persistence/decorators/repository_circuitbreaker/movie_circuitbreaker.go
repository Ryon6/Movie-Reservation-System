package repository_circuitbreaker

import (
	"context"
	"mrs/internal/domain/movie"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/shared/vo"
	applog "mrs/pkg/log"

	"github.com/afex/hystrix-go/hystrix"
)

const (
	cmdMovieRead  = "movie.Read"
	cmdMovieWrite = "movie.Write"
)

// 配置MovieRepository的熔断器
func ConfigMovieRepositoryBreakers() {
	hystrix.ConfigureCommand(cmdMovieRead, hystrix.CommandConfig{
		Timeout:                1000, // 读操作超时1秒
		MaxConcurrentRequests:  200,  // 允许更多并发读
		RequestVolumeThreshold: 20,
		SleepWindow:            5000,
		ErrorPercentThreshold:  50,
	})
	hystrix.ConfigureCommand(cmdMovieWrite, hystrix.CommandConfig{
		Timeout:                2000, // 写操作允许稍长超时
		MaxConcurrentRequests:  100,  // 限制并发写，保护数据库
		RequestVolumeThreshold: 10,
		SleepWindow:            5000,
		ErrorPercentThreshold:  50,
	})
}

// 熔断降级的MovieRepository，实现movie.MovieRepository接口
type movieRepositoryWithCircuitBreaker struct {
	repo   movie.MovieRepository
	logger applog.Logger
}

func NewMovieRepositoryWithCircuitBreaker(repo movie.MovieRepository, logger applog.Logger) movie.MovieRepository {
	return &movieRepositoryWithCircuitBreaker{
		repo:   repo,
		logger: logger.With(applog.String("Repository", "movieRepositoryWithCircuitBreaker")),
	}
}

func (r *movieRepositoryWithCircuitBreaker) execute(
	ctx context.Context,
	cmd string,
	run func(ctx context.Context) error,
	fallback func(ctx context.Context, err error) error,
) error {
	return hystrix.DoC(ctx, cmd, run, fallback)
}

func (r *movieRepositoryWithCircuitBreaker) Create(ctx context.Context, mv *movie.Movie) (*movie.Movie, error) {
	logger := r.logger.With(applog.String("Method", "Create"))

	var createdMovie *movie.Movie

	run := func(ctx context.Context) error {
		mv, err := r.repo.Create(ctx, mv)
		if err != nil {
			return err
		}
		createdMovie = mv
		return err
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitWriteOperationBusy
	}

	err := r.execute(ctx, cmdMovieWrite, run, fallback)
	if err != nil {
		logger.Error("create movie circuit breaker fallback", applog.Error(err))
		return nil, err
	}

	return createdMovie, nil
}

func (r *movieRepositoryWithCircuitBreaker) FindByID(ctx context.Context, id vo.MovieID) (*movie.Movie, error) {
	logger := r.logger.With(applog.String("Method", "FindByID"))

	var movieResult *movie.Movie

	run := func(ctx context.Context) error {
		mv, err := r.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		movieResult = mv
		return err
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitReadOperationBusy
	}

	err := r.execute(ctx, cmdMovieRead, run, fallback)
	if err != nil {
		logger.Warn("find movie by id circuit breaker fallback", applog.Error(err))
		return nil, err
	}

	return movieResult, nil
}

func (r *movieRepositoryWithCircuitBreaker) FindByIDs(ctx context.Context, ids []vo.MovieID) ([]*movie.Movie, error) {
	logger := r.logger.With(applog.String("Method", "FindByIDs"))

	var movieResults []*movie.Movie

	run := func(ctx context.Context) error {
		mvs, err := r.repo.FindByIDs(ctx, ids)
		if err != nil {
			return err
		}
		movieResults = mvs
		return err
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitReadOperationBusy
	}

	err := r.execute(ctx, cmdMovieRead, run, fallback)
	if err != nil {
		logger.Warn("find movies by ids circuit breaker fallback", applog.Error(err))
		return nil, err
	}

	return movieResults, nil
}

func (r *movieRepositoryWithCircuitBreaker) FindByTitle(ctx context.Context, title string) (*movie.Movie, error) {
	logger := r.logger.With(applog.String("Method", "FindByTitle"))

	var movieResult *movie.Movie

	run := func(ctx context.Context) error {
		mv, err := r.repo.FindByTitle(ctx, title)
		if err != nil {
			return err
		}
		movieResult = mv
		return err
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitReadOperationBusy
	}

	err := r.execute(ctx, cmdMovieRead, run, fallback)
	if err != nil {
		logger.Warn("find movie by title circuit breaker fallback", applog.Error(err))
		return nil, err
	}

	return movieResult, nil
}

func (r *movieRepositoryWithCircuitBreaker) Update(ctx context.Context, mv *movie.Movie) error {
	logger := r.logger.With(applog.String("Method", "Update"))

	run := func(ctx context.Context) error {
		return r.repo.Update(ctx, mv)
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitWriteOperationBusy
	}

	err := r.execute(ctx, cmdMovieWrite, run, fallback)
	if err != nil {
		logger.Error("update movie circuit breaker fallback", applog.Error(err))
		return err
	}

	return nil
}

func (r *movieRepositoryWithCircuitBreaker) CheckGenreReferenced(ctx context.Context, genreID vo.GenreID) (bool, error) {
	logger := r.logger.With(applog.String("Method", "CheckGenreReferenced"))

	var isReferenced bool

	run := func(ctx context.Context) error {
		reference, err := r.repo.CheckGenreReferenced(ctx, genreID)
		if err != nil {
			return err
		}
		isReferenced = reference
		return err
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitReadOperationBusy
	}

	err := r.execute(ctx, cmdMovieRead, run, fallback)
	if err != nil {
		logger.Error("check genre referenced circuit breaker fallback", applog.Error(err))
		return false, err
	}

	return isReferenced, nil
}

func (r *movieRepositoryWithCircuitBreaker) List(ctx context.Context, options *movie.MovieQueryOptions) ([]*movie.Movie, int64, error) {
	logger := r.logger.With(applog.String("Method", "List"))

	var movieResults []*movie.Movie
	var total int64

	run := func(ctx context.Context) error {
		mvs, listTotal, err := r.repo.List(ctx, options)
		if err != nil {
			return err
		}
		movieResults = mvs
		total = listTotal
		return err
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitReadOperationBusy
	}

	err := r.execute(ctx, cmdMovieRead, run, fallback)
	if err != nil {
		logger.Warn("list movies circuit breaker fallback", applog.Error(err))
		return nil, 0, err
	}

	return movieResults, total, nil
}

func (r *movieRepositoryWithCircuitBreaker) Delete(ctx context.Context, id vo.MovieID) error {
	logger := r.logger.With(applog.String("Method", "Delete"))

	run := func(ctx context.Context) error {
		return r.repo.Delete(ctx, id)
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitWriteOperationBusy
	}

	err := r.execute(ctx, cmdMovieWrite, run, fallback)
	if err != nil {
		logger.Error("delete movie circuit breaker fallback", applog.Error(err))
		return err
	}

	return nil
}

func (r *movieRepositoryWithCircuitBreaker) ReplaceGenresForMovie(ctx context.Context, movie *movie.Movie, genres []*movie.Genre) error {
	logger := r.logger.With(applog.String("Method", "ReplaceGenresForMovie"))

	run := func(ctx context.Context) error {
		return r.repo.ReplaceGenresForMovie(ctx, movie, genres)
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitWriteOperationBusy
	}

	err := r.execute(ctx, cmdMovieWrite, run, fallback)
	if err != nil {
		logger.Error("replace genres for movie circuit breaker fallback", applog.Error(err))
		return err
	}

	return nil
}
