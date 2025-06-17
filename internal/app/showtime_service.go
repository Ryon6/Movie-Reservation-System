package app

import (
	"context"
	"errors"
	"fmt"
	"math"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/shared/lock"
	"mrs/internal/domain/shared/vo"
	"mrs/internal/domain/showtime"
	applog "mrs/pkg/log"
	"time"
)

type ShowtimeService interface {
	CreateShowtime(ctx context.Context, req *request.CreateShowtimeRequest) (*response.ShowtimeResponse, error)
	GetShowtime(ctx context.Context, req *request.GetShowtimeRequest) (*response.ShowtimeResponse, error)
	UpdateShowtime(ctx context.Context, req *request.UpdateShowtimeRequest) (*response.ShowtimeResponse, error)
	DeleteShowtime(ctx context.Context, req *request.DeleteShowtimeRequest) error
	ListShowtimes(ctx context.Context, req *request.ListShowtimesRequest) (*response.PaginatedShowtimeResponse, error)
	GetSeatMap(ctx context.Context, req *request.GetSeatMapRequest) (*response.SeatMapResponse, error)
}

type showtimeService struct {
	uow          shared.UnitOfWork
	showRepo     showtime.ShowtimeRepository
	seatRepo     cinema.SeatRepository
	showCache    showtime.ShowtimeCache
	seatCache    cinema.SeatCache
	lockProvider lock.LockProvider
	logger       applog.Logger
}

func NewShowtimeService(
	uow shared.UnitOfWork,
	showRepo showtime.ShowtimeRepository,
	seatRepo cinema.SeatRepository,
	showCache showtime.ShowtimeCache,
	seatCache cinema.SeatCache,
	lockProvider lock.LockProvider,
	logger applog.Logger,
) ShowtimeService {
	return &showtimeService{
		uow:          uow,
		showRepo:     showRepo,
		seatRepo:     seatRepo,
		showCache:    showCache,
		seatCache:    seatCache,
		lockProvider: lockProvider,
		logger:       logger.With(applog.String("Service", "ShowtimeService")),
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
		overlap, err := showtimeRepo.CheckOverlap(ctx, st.CinemaHallID, st.StartTime, st.EndTime)
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

	// 需要将完整信息设置到缓存中
	st, err = s.showRepo.FindByID(ctx, vo.ShowtimeID(st.ID))
	if err != nil {
		logger.Error("failed to find showtime", applog.Error(err))
		return nil, err
	}
	s.showCache.SetShowtime(ctx, st, 0)

	logger.Info("create showtime successfully", applog.Uint("showtime_id", uint(st.ID)))
	return response.ToShowtimeResponse(st), nil
}

func (s *showtimeService) GetShowtime(ctx context.Context, req *request.GetShowtimeRequest) (*response.ShowtimeResponse, error) {
	logger := s.logger.With(applog.String("Method", "GetShowtime"), applog.Uint("showtime_id", req.ID))
	showtime, err := s.showCache.GetShowtime(ctx, vo.ShowtimeID(req.ID))
	if err != nil {
		logger.Warn("failed to get showtime", applog.Error(err))
	} else {
		logger.Info("get showtime from cache successfully", applog.Uint("showtime_id", uint(showtime.ID)))
		return response.ToShowtimeResponse(showtime), nil
	}

	showtime, err = s.showRepo.FindByID(ctx, vo.ShowtimeID(req.ID))
	if err != nil {
		logger.Error("failed to get showtime", applog.Error(err))
		return nil, err
	}

	s.showCache.SetShowtime(ctx, showtime, 0)
	logger.Info("get showtime successfully", applog.Uint("showtime_id", uint(showtime.ID)))
	return response.ToShowtimeResponse(showtime), nil
}

// 更新场次
func (s *showtimeService) UpdateShowtime(ctx context.Context, req *request.UpdateShowtimeRequest) (*response.ShowtimeResponse, error) {
	logger := s.logger.With(applog.String("Method", "UpdateShowtime"), applog.Uint("showtime_id", req.ID))
	st := req.ToDomain()

	// 更新场次时，需要检查是否重叠，如果重叠，则返回错误。否则更新场次。
	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		showtimeRepo := provider.GetShowtimeRepository()
		// 检查是否重叠
		overlap, err := showtimeRepo.CheckOverlap(ctx, st.CinemaHallID, st.StartTime, st.EndTime, st.ID)
		if err != nil {
			logger.Error("failed to check overlap", applog.Error(err))
			return err
		}
		if overlap {
			logger.Error("showtime overlaps with existing showtime", applog.Uint("cinema_hall_id", uint(st.CinemaHallID)))
			return fmt.Errorf("ServiceError: %w", showtime.ErrShowtimeOverlap)
		}
		// 更新场次
		err = showtimeRepo.Update(ctx, st)
		if err != nil {
			// 如果场次不存在，则返回错误(CheckOverlap并不能保证场次存在)
			if errors.Is(err, showtime.ErrShowtimeNotFound) {
				logger.Warn("showtime not found")
				return err
			}
			logger.Error("failed to update showtime", applog.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to update showtime", applog.Error(err))
		return nil, err
	}

	// 缓存中需要包含完整内容
	st, err = s.showRepo.FindByID(ctx, vo.ShowtimeID(req.ID))
	if err != nil {
		logger.Error("failed to find showtime", applog.Error(err))
		return nil, err
	}
	s.showCache.SetShowtime(ctx, st, 0)

	logger.Info("update showtime successfully", applog.Uint("showtime_id", uint(st.ID)))
	return response.ToShowtimeResponse(st), nil
}

// 删除场次
func (s *showtimeService) DeleteShowtime(ctx context.Context, req *request.DeleteShowtimeRequest) error {
	logger := s.logger.With(applog.String("Method", "DeleteShowtime"), applog.Uint("showtime_id", req.ID))
	// 单条场次删除，不需要事务
	err := s.showRepo.Delete(ctx, vo.ShowtimeID(req.ID))
	if err != nil {
		// 如果场次不存在，则返回错误(仓库底层实现会根据RowAffected判断记录是否存在)
		if errors.Is(err, showtime.ErrShowtimeNotFound) {
			logger.Warn("showtime not found")
			return err
		}
		logger.Error("failed to delete showtime", applog.Error(err))
		return err
	}
	// 删除缓存
	s.showCache.DeleteShowtime(ctx, vo.ShowtimeID(req.ID))
	logger.Info("delete showtime successfully", applog.Uint("showtime_id", req.ID))
	return nil
}

func (s *showtimeService) ListShowtimes(ctx context.Context,
	req *request.ListShowtimesRequest) (*response.PaginatedShowtimeResponse, error) {
	logger := s.logger.With(applog.String("Method", "ListShowtimes"),
		applog.Uint("movie_id", req.MovieID),
		applog.Uint("cinema_hall_id", req.CinemaHallID),
		applog.Time("date", req.Date))

	options := req.ToDomain()

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
	cacheResult, err := s.showCache.GetShowtimeList(ctx, options)
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
		showtimes, _, err = s.showRepo.List(ctx, options)
		if err != nil {
			logger.Error("failed to list showtimes", applog.Error(err))
			return nil, err
		}
	}

	logger.Info("list showtimes successfully", applog.Int("total", int(len(showtimes))))

	if err := s.showCache.SetShowtimeList(ctx, showtimes, options, 0); err != nil {
		logger.Error("failed to set showtime list to cache", applog.Error(err))
	}
	return fn(showtimes), nil
}

// 获取座位表
func (s *showtimeService) GetSeatMap(ctx context.Context, req *request.GetSeatMapRequest) (*response.SeatMapResponse, error) {
	logger := s.logger.With(applog.String("Method", "GetSeatMap"), applog.Uint("showtime_id", req.ShowtimeID))
	seatInfos, err := s.seatCache.GetSeatMap(ctx, vo.ShowtimeID(req.ShowtimeID))
	if err == nil {
		return &response.SeatMapResponse{Seats: seatInfos}, nil
	}

	// 缓存未命中，且不是缓存缺失错误
	if !errors.Is(err, shared.ErrCacheMissing) {
		logger.Error("failed to get seat map from cache", applog.Error(err))
	}

	// 获取分布式锁，防止并发初始化座位表
	lockKey := cinema.GetShowtimeSeatsInitLockKey(vo.ShowtimeID(req.ShowtimeID))
	// 获取锁失败，则重试
	var initLock lock.Lock = nil
	for i := 0; i < lock.DefaultMaxRetries; i++ {
		initLock, err = s.lockProvider.Acquire(ctx, lockKey, lock.DefaultLockTTL)
		if errors.Is(err, lock.ErrLockNotAcquired) {
			logger.Warn("other process is initializing seat map, retrying...")
			time.Sleep(lock.DefaultBackoff)
			continue
		}
		if err != nil {
			logger.Error("failed to acquire lock", applog.Error(err))
			return nil, err
		}
		defer initLock.Release(ctx)
		break
	}
	// 若三次获取锁失败，则返回错误
	if initLock == nil {
		logger.Error("retry to acquire lock failed", applog.Int("retries", lock.DefaultMaxRetries))
		return nil, fmt.Errorf("retry to acquire lock failed")
	}

	// 获取场次信息（应用层服务，会先从缓存中获取，如果缓存未命中，则从数据库中获取）
	showtimeResp, err := s.GetShowtime(ctx, &request.GetShowtimeRequest{ID: req.ShowtimeID})
	if err != nil {
		if errors.Is(err, showtime.ErrShowtimeNotFound) {
			logger.Warn("showtime not found")
			return nil, err
		}
		logger.Error("failed to get showtime", applog.Error(err))
		return nil, err
	}

	// 获取座位表
	seats, err := s.seatRepo.FindByHallID(ctx, uint(showtimeResp.CinemaHall.ID))
	if err != nil {
		logger.Error("failed to find seats", applog.Error(err))
		return nil, err
	}

	// TODO: 获取已售座位 需要booking_repository支持

	// 座位表缓存过期时间设置为场次结束时间后10分钟
	expireTime := time.Until(showtimeResp.EndTime.Add(time.Minute * 10))
	if err := s.seatCache.InitSeatMap(ctx, vo.ShowtimeID(req.ShowtimeID), seats, nil, expireTime); err != nil {
		logger.Error("failed to init seat map", applog.Error(err))
		return nil, err
	}

	// 将座位表转换为响应格式
	seatInfos = make([]*cinema.SeatInfo, 0, len(seats))
	for _, seat := range seats {
		seatInfos = append(seatInfos, &cinema.SeatInfo{
			ID:            seat.ID,
			Type:          seat.Type,
			RowIdentifier: seat.RowIdentifier,
			SeatNumber:    seat.SeatNumber,
			Status:        cinema.SeatStatusAvailable, // 先默认可用，后续再根据已售座位更新状态
		})
	}

	return &response.SeatMapResponse{Seats: seatInfos}, nil
}
