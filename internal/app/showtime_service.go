package app

import (
	"context"
	"fmt"
	"math"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/showtime"
	"mrs/internal/infrastructure/cache"
	applog "mrs/pkg/log"
)

type ShowtimeService interface {
	CreateShowtime(ctx context.Context, req *request.CreateShowtimeRequest) (*response.ShowtimeResponse, error)
	GetShowtime(ctx context.Context, req *request.GetShowtimeRequest) (*response.ShowtimeResponse, error)
}

type showtimeService struct {
	uow       shared.UnitOfWork
	showRepo  showtime.ShowtimeRepository
	showCache cache.ShowtimeCache
	logger    applog.Logger
}

func NewShowtimeService(
	uow shared.UnitOfWork,
	showCache cache.ShowtimeCache,
	logger applog.Logger,
) ShowtimeService {
	return &showtimeService{
		uow:       uow,
		showCache: showCache,
		logger:    logger.With(applog.String("Service", "ShowtimeService")),
	}
}

// showtime查询应该分为两种情况，查列表时，返回简易结果；根据ID查询时，返回完整结果（包括电影，影厅，座位）

// 需要返回完整结果
func (s *showtimeService) CreateShowtime(ctx context.Context, req *request.CreateShowtimeRequest) (*response.ShowtimeResponse, error) {
	logger := s.logger.With(applog.String("Method", "CreateShowtime"), applog.Uint("movie_id", req.MovieID), applog.Uint("cinema_hall_id", req.CinemaHallID))
	st := req.ToDomain()
	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		var err error
		showtimeRepo := provider.GetShowtimeRepository()
		overlap, err := showtimeRepo.CheckOverlap(ctx, uint(st.CinemaHallID), st.StartTime, st.EndTime)
		if err != nil {
			logger.Error("failed to check overlap", applog.Error(err))
			return err
		}
		if overlap {
			logger.Error("showtime overlaps with existing showtime", applog.Uint("cinema_hall_id", uint(st.CinemaHallID)))
			return fmt.Errorf("ServiceError: %w", showtime.ErrShowtimeOverlap)
		}
		st, err = showtimeRepo.Create(ctx, st)
		if err != nil {
			logger.Error("failed to create showtime", applog.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to create showtime", applog.Error(err))
		return nil, err
	}
	s.showCache.SetShowtime(ctx, st, 0)

	logger.Info("create showtime successfully", applog.Uint("showtime_id", uint(st.ID)))
	return response.ToShowtimeResponse(st), nil
}

func (s *showtimeService) GetShowtime(ctx context.Context, req *request.GetShowtimeRequest) (*response.ShowtimeResponse, error) {
	logger := s.logger.With(applog.String("Method", "GetShowtime"), applog.Uint("showtime_id", req.ID))
	showtime, err := s.showCache.GetShowtime(ctx, req.ID)
	if err != nil {
		logger.Warn("failed to get showtime", applog.Error(err))
	} else {
		logger.Info("get showtime from cache successfully", applog.Uint("showtime_id", uint(showtime.ID)))
		return response.ToShowtimeResponse(showtime), nil
	}

	showtime, err = s.showRepo.FindByID(ctx, req.ID)
	if err != nil {
		logger.Error("failed to get showtime", applog.Error(err))
		return nil, err
	}

	s.showCache.SetShowtime(ctx, showtime, 0)
	logger.Info("get showtime successfully", applog.Uint("showtime_id", uint(showtime.ID)))
	return response.ToShowtimeResponse(showtime), nil
}

func (s *showtimeService) UpdateShowtime(ctx context.Context, req *request.UpdateShowtimeRequest) (*response.ShowtimeResponse, error) {
	logger := s.logger.With(applog.String("Method", "UpdateShowtime"), applog.Uint("showtime_id", req.ID))
	st := req.ToDomain()
	// 检查是否重叠
	overlap, err := s.showRepo.CheckOverlap(ctx, uint(st.CinemaHallID), st.StartTime, st.EndTime, uint(st.ID))
	if err != nil {
		logger.Error("failed to check overlap", applog.Error(err))
		return nil, err
	}
	if overlap {
		logger.Error("showtime overlaps with existing showtime", applog.Uint("cinema_hall_id", uint(st.CinemaHallID)))
		return nil, fmt.Errorf("ServiceError: %w", showtime.ErrShowtimeOverlap)
	}
	err = s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		showtimeRepo := provider.GetShowtimeRepository()
		err = showtimeRepo.Update(ctx, st)
		if err != nil {
			logger.Error("failed to update showtime", applog.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to update showtime", applog.Error(err))
		return nil, err
	}

	logger.Info("update showtime successfully", applog.Uint("showtime_id", uint(st.ID)))
	s.showCache.SetShowtime(ctx, st, 0)
	return response.ToShowtimeResponse(st), nil
}

func (s *showtimeService) DeleteShowtime(ctx context.Context, req *request.DeleteShowtimeRequest) error {
	logger := s.logger.With(applog.String("Method", "DeleteShowtime"), applog.Uint("showtime_id", req.ID))
	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		showtimeRepo := provider.GetShowtimeRepository()
		return showtimeRepo.Delete(ctx, req.ID)
	})
	if err != nil {
		logger.Error("failed to delete showtime", applog.Error(err))
		return err
	}
	logger.Info("delete showtime successfully", applog.Uint("showtime_id", req.ID))
	return nil
}

func (s *showtimeService) ListShowtimes(ctx context.Context,
	req *request.ListShowtimesRequest) (*response.PaginatedShowtimeResponse, error) {
	logger := s.logger.With(applog.String("Method", "ListShowtimes"),
		applog.Uint("movie_id", req.MovieID),
		applog.Uint("cinema_hall_id", req.CinemaHallID),
		applog.Time("date", req.Date))

	filters := make(map[string]interface{})
	if req.MovieID > 0 {
		filters["movie_id"] = req.MovieID
	}
	if req.CinemaHallID > 0 {
		filters["cinema_hall_id"] = req.CinemaHallID
	}
	if !req.Date.IsZero() {
		filters["date"] = req.Date
	}

	var showtimes []*showtime.Showtime
	var fn = func(showtimes []*showtime.Showtime) *response.PaginatedShowtimeResponse {
		responseShowtimes := make([]*response.ShowtimeSimpleResponse, 0, len(showtimes))
		total := len(showtimes)
		startIndex := (req.Page - 1) * req.PageSize
		endIndex := min(startIndex+req.PageSize, total)
		showtimes = showtimes[startIndex:endIndex]
		for _, showtime := range showtimes {
			responseShowtimes = append(responseShowtimes, response.ToShowtimeSimpleResponse(showtime))
		}
		return &response.PaginatedShowtimeResponse{
			Pagination: response.PaginationResponse{
				Page:       req.Page,
				PageSize:   req.PageSize,
				TotalPages: int(math.Ceil(float64(total) / float64(req.PageSize))),
				TotalCount: int(total),
			},
			Showtimes: responseShowtimes,
		}
	}

	// 列表缓存命中(err == nil)有三种情况：
	// 1. 列表缓存为空（即无数据对应过滤条件） -> 直接返回空列表
	// 2. 列表缓存中存在数据，但部分showtime记录缺失（即showtime_id列表中存在但缓存中不存在） -> 进一步查询数据库
	// 3. 列表缓存中存在数据，且所有showtime记录都存在（即showtime_id列表中所有showtime记录都存在） -> 直接返回缓存数据
	// 而列表缓存未命中，则需要进一步查询数据库
	cacheResult, err := s.showCache.GetShowtimeList(ctx, filters)
	if err != nil {
		logger.Warn("failed to get showtime list from cache", applog.Error(err))
	} else {
		logger.Info("get showtime list from cache successfully")
		if len(cacheResult.MissingShowtimeIDs) == 0 {
			logger.Info("all showtimes found in cache", applog.Int("total", len(cacheResult.Showtimes)))
			return fn(cacheResult.Showtimes), nil
		} else {
			logger.Info("some showtimes not found in cache", applog.Int("missing", len(cacheResult.MissingShowtimeIDs)))
			showtimes = cacheResult.Showtimes
		}
	}

	if len(showtimes) != 0 {
		missingShowtimes, err := s.showRepo.FindByIDs(ctx, cacheResult.MissingShowtimeIDs)
		if err != nil {
			logger.Error("failed to find missing showtimes", applog.Error(err))
			return nil, err
		}
		showtimes = append(showtimes, missingShowtimes...)
	} else {
		showtimes, _, err = s.showRepo.List(ctx, req.Page, req.PageSize, filters)
		if err != nil {
			logger.Error("failed to list showtimes", applog.Error(err))
			return nil, err
		}
	}

	logger.Info("list showtimes successfully", applog.Int("total", int(len(showtimes))))

	if err := s.showCache.SetShowtimeList(ctx, showtimes, filters, 0); err != nil {
		logger.Error("failed to set showtime list to cache", applog.Error(err))
	}
	return fn(showtimes), nil
}
