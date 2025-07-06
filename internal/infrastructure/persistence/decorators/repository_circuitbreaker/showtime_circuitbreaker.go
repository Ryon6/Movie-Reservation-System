package repository_circuitbreaker

import (
	"context"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/shared/vo"
	"mrs/internal/domain/showtime"
	applog "mrs/pkg/log"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

const (
	cmdShowtimeRead  = "showtime.Read"
	cmdShowtimeWrite = "showtime.Write"
)

func ConfigShowtimeRepositoryBreakers() {
	hystrix.ConfigureCommand(cmdShowtimeRead, hystrix.CommandConfig{
		Timeout:                1000, // 读操作超时1秒
		MaxConcurrentRequests:  200,  // 允许更多并发读
		RequestVolumeThreshold: 20,
		SleepWindow:            5000,
		ErrorPercentThreshold:  50,
	})
	hystrix.ConfigureCommand(cmdShowtimeWrite, hystrix.CommandConfig{
		Timeout:                2000, // 写操作允许稍长超时
		MaxConcurrentRequests:  100,  // 限制并发写，保护数据库
		RequestVolumeThreshold: 10,
		SleepWindow:            5000,
		ErrorPercentThreshold:  50,
	})
}

type showtimeRepositoryWithCircuitBreaker struct {
	repo   showtime.ShowtimeRepository
	logger applog.Logger
}

func NewShowtimeRepositoryWithCircuitBreaker(repo showtime.ShowtimeRepository, logger applog.Logger) showtime.ShowtimeRepository {
	return &showtimeRepositoryWithCircuitBreaker{
		repo:   repo,
		logger: logger.With(applog.String("Repository", "showtimeRepositoryWithCircuitBreaker")),
	}
}

func (r *showtimeRepositoryWithCircuitBreaker) execute(
	ctx context.Context,
	cmd string,
	run func(ctx context.Context) error,
	fallback func(ctx context.Context, err error) error,
) error {
	return hystrix.DoC(ctx, cmd, run, fallback)
}

func (r *showtimeRepositoryWithCircuitBreaker) Create(ctx context.Context, st *showtime.Showtime) (*showtime.Showtime, error) {
	logger := r.logger.With(applog.String("Method", "Create"))

	var createdShowtime *showtime.Showtime

	run := func(ctx context.Context) error {
		st, err := r.repo.Create(ctx, st)
		if err != nil {
			return err
		}
		createdShowtime = st
		return nil
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitWriteOperationBusy
	}

	err := r.execute(ctx, cmdShowtimeWrite, run, fallback)
	if err != nil {
		logger.Error("create showtime circuit breaker fallback", applog.Error(err))
		return nil, err
	}

	return createdShowtime, nil
}

func (r *showtimeRepositoryWithCircuitBreaker) FindByID(ctx context.Context, id vo.ShowtimeID) (*showtime.Showtime, error) {
	logger := r.logger.With(applog.String("Method", "FindByID"))

	var showtimeResult *showtime.Showtime

	run := func(ctx context.Context) error {
		st, err := r.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		showtimeResult = st
		return nil
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitReadOperationBusy
	}

	err := r.execute(ctx, cmdShowtimeRead, run, fallback)
	if err != nil {
		logger.Warn("find showtime by id circuit breaker fallback", applog.Error(err))
		return nil, err
	}

	return showtimeResult, nil
}

func (r *showtimeRepositoryWithCircuitBreaker) FindByIDs(ctx context.Context, ids []vo.ShowtimeID) ([]*showtime.Showtime, error) {
	logger := r.logger.With(applog.String("Method", "FindByIDs"))

	var showtimeResults []*showtime.Showtime

	run := func(ctx context.Context) error {
		sts, err := r.repo.FindByIDs(ctx, ids)
		if err != nil {
			return err
		}
		showtimeResults = sts
		return nil
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitReadOperationBusy
	}

	err := r.execute(ctx, cmdShowtimeRead, run, fallback)
	if err != nil {
		logger.Warn("find showtimes by ids circuit breaker fallback", applog.Error(err))
		return nil, err
	}

	return showtimeResults, nil
}

func (r *showtimeRepositoryWithCircuitBreaker) List(ctx context.Context, options *showtime.ShowtimeQueryOptions) ([]*showtime.Showtime, int64, error) {
	logger := r.logger.With(applog.String("Method", "List"))

	var showtimeResults []*showtime.Showtime
	var total int64

	run := func(ctx context.Context) error {
		sts, listTotal, err := r.repo.List(ctx, options)
		if err != nil {
			return err
		}
		showtimeResults = sts
		total = listTotal
		return nil
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitReadOperationBusy
	}

	err := r.execute(ctx, cmdShowtimeRead, run, fallback)
	if err != nil {
		logger.Warn("list showtimes circuit breaker fallback", applog.Error(err))
		return nil, 0, err
	}

	return showtimeResults, total, nil
}

func (r *showtimeRepositoryWithCircuitBreaker) Update(ctx context.Context, st *showtime.Showtime) error {
	logger := r.logger.With(applog.String("Method", "Update"))

	run := func(ctx context.Context) error {
		return r.repo.Update(ctx, st)
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitWriteOperationBusy
	}

	err := r.execute(ctx, cmdShowtimeWrite, run, fallback)
	if err != nil {
		logger.Error("update showtime circuit breaker fallback", applog.Error(err))
		return err
	}

	return nil
}

func (r *showtimeRepositoryWithCircuitBreaker) Delete(ctx context.Context, id vo.ShowtimeID) error {
	logger := r.logger.With(applog.String("Method", "Delete"))

	run := func(ctx context.Context) error {
		return r.repo.Delete(ctx, id)
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitWriteOperationBusy
	}

	err := r.execute(ctx, cmdShowtimeWrite, run, fallback)
	if err != nil {
		logger.Error("delete showtime circuit breaker fallback", applog.Error(err))
		return err
	}

	return nil
}

func (r *showtimeRepositoryWithCircuitBreaker) CheckOverlap(ctx context.Context, hallID vo.CinemaHallID, startTime, endTime time.Time, excludeShowtimeID ...vo.ShowtimeID) (bool, error) {
	logger := r.logger.With(applog.String("Method", "CheckOverlap"))

	var isOverlap bool

	run := func(ctx context.Context) error {
		overlap, err := r.repo.CheckOverlap(ctx, hallID, startTime, endTime, excludeShowtimeID...)
		if err != nil {
			return err
		}
		isOverlap = overlap
		return nil
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitReadOperationBusy
	}

	err := r.execute(ctx, cmdShowtimeRead, run, fallback)
	if err != nil {
		logger.Warn("check overlap circuit breaker fallback", applog.Error(err))
		return false, err
	}

	return isOverlap, nil
}

func (r *showtimeRepositoryWithCircuitBreaker) FindShowtimesByMovieAndDateRanges(ctx context.Context, movieID vo.MovieID, startDate, endDate time.Time) ([]*showtime.Showtime, error) {
	logger := r.logger.With(applog.String("Method", "FindShowtimesByMovieAndDateRanges"))

	var showtimeResults []*showtime.Showtime

	run := func(ctx context.Context) error {
		sts, err := r.repo.FindShowtimesByMovieAndDateRanges(ctx, movieID, startDate, endDate)
		if err != nil {
			return err
		}
		showtimeResults = sts
		return nil
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitReadOperationBusy
	}

	err := r.execute(ctx, cmdShowtimeRead, run, fallback)
	if err != nil {
		logger.Warn("find showtimes by movie and date ranges circuit breaker fallback", applog.Error(err))
		return nil, err
	}

	return showtimeResults, nil
}

func (r *showtimeRepositoryWithCircuitBreaker) FindShowtimesByHallAndDateRanges(ctx context.Context, hallID vo.CinemaHallID, startDate, endDate time.Time) ([]*showtime.Showtime, error) {
	logger := r.logger.With(applog.String("Method", "FindShowtimesByHallAndDateRanges"))

	var showtimeResults []*showtime.Showtime

	run := func(ctx context.Context) error {
		sts, err := r.repo.FindShowtimesByHallAndDateRanges(ctx, hallID, startDate, endDate)
		if err != nil {
			return err
		}
		showtimeResults = sts
		return nil
	}

	fallback := func(ctx context.Context, err error) error {
		return shared.ErrCircuitReadOperationBusy
	}

	err := r.execute(ctx, cmdShowtimeRead, run, fallback)
	if err != nil {
		logger.Warn("find showtimes by hall and date ranges circuit breaker fallback", applog.Error(err))
		return nil, err
	}

	return showtimeResults, nil
}
