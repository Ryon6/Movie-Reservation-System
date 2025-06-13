package app

import (
	"context"
	"errors"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/domain/booking"
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/shared/lock"
	"mrs/internal/domain/shared/vo"
	"mrs/internal/domain/showtime"
	applog "mrs/pkg/log"
)

type BookingService interface {
	CreateBooking(ctx context.Context, req *request.CreateBookingRequest) (*response.BookingResponse, error)
}

type bookingService struct {
	uow           shared.UnitOfWork
	bookingRepo   booking.BookingRepository
	showtimeRepo  showtime.ShowtimeRepository
	seatCache     cinema.SeatCache
	showtimeCache showtime.ShowtimeCache
	lockProvider  lock.LockProvider
	logger        applog.Logger
}

func NewBookingService(
	uow shared.UnitOfWork,
	bookingRepo booking.BookingRepository,
	showtimeRepo showtime.ShowtimeRepository,
	seatCache cinema.SeatCache,
	showtimeCache showtime.ShowtimeCache,
	lockProvider lock.LockProvider,
	logger applog.Logger) BookingService {

	return &bookingService{
		uow:           uow,
		bookingRepo:   bookingRepo,
		showtimeRepo:  showtimeRepo,
		seatCache:     seatCache,
		showtimeCache: showtimeCache,
		lockProvider:  lockProvider,
		logger:        logger.With(applog.String("Service", "BookingService")),
	}
}

// CreateBooking 创建订单
func (s *bookingService) CreateBooking(ctx context.Context, req *request.CreateBookingRequest) (*response.BookingResponse, error) {
	logger := s.logger.With(applog.String("Method", "CreateBooking"))
	lockKey := cinema.GetShowtimeSeatsLockKey(vo.ShowtimeID(req.ShowtimeID))

	// 获取场次信息
	st, err := s.showtimeCache.GetShowtime(ctx, vo.ShowtimeID(req.ShowtimeID))
	if err != nil {
		logger.Error("failed to get showtime from cache", applog.Error(err))
	} else if st == nil {
		st, err = s.showtimeRepo.FindByID(ctx, vo.ShowtimeID(req.ShowtimeID))
		if err != nil {
			logger.Error("failed to get showtime from repo", applog.Error(err))
			return nil, err
		}
		if err := s.showtimeCache.SetShowtime(ctx, st, showtime.DefaultExpiration); err != nil {
			logger.Error("failed to set showtime to cache", applog.Error(err))
		}
	}

	// 获取分布式锁（场次锁）
	lock, err := s.lockProvider.Acquire(ctx, lockKey, lock.DefaultLockTTL)
	if err != nil {
		logger.Error("failed to acquire lock", applog.Error(err))
		return nil, err
	}
	defer lock.Release(ctx)

	// 获取座位ID列表
	seatIDs := make([]vo.SeatID, len(req.SeatIDs))
	for i, seatID := range req.SeatIDs {
		seatIDs[i] = vo.SeatID(seatID)
	}

	// 在缓存中锁定座位（防止超额预订）
	if err = s.seatCache.LockSeats(ctx, vo.ShowtimeID(req.ShowtimeID), seatIDs); err != nil {
		if errors.Is(err, booking.ErrBookedSeatAlreadyLocked) {
			logger.Warn("booked seats already locked", applog.Error(err))
			return nil, err
		}
		logger.Error("failed to lock seats", applog.Error(err))
		return nil, err
	}

	// 若创建订单失败，则释放座位锁
	defer func() {
		if err != nil {
			s.seatCache.ReleaseSeats(ctx, vo.ShowtimeID(req.ShowtimeID), seatIDs)
		}
	}()

	// 创建已预订的座位
	bookedSeats := make([]*booking.BookedSeat, len(seatIDs))
	for i, seatID := range seatIDs {
		bookedSeats[i] = booking.NewBookedSeat(vo.ShowtimeID(req.ShowtimeID), seatID, st.Price)
	}

	// 创建订单
	totalPrice := float64(len(seatIDs)) * st.Price
	booking := booking.NewBooking(vo.UserID(req.UserID), vo.ShowtimeID(req.ShowtimeID), bookedSeats, totalPrice)

	// 使用事务，确保两个操作要么都成功，要么都失败(先创建订单，再将订单ID写入bookedSeats)
	err = s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		bookingRepo := provider.GetBookingRepository()
		bookedSeatRepo := provider.GetBookedSeatRepository()
		booking, err = bookingRepo.CreateBooking(ctx, booking)
		if err != nil {
			logger.Error("failed to create booking", applog.Error(err))
			return err
		}

		// 将订单ID写入已预订的座位
		for i := range bookedSeats {
			bookedSeats[i].BookingID = booking.ID
		}

		// 创建已预订的座位
		bookedSeats, err = bookedSeatRepo.CreateBookedSeats(ctx, bookedSeats)
		if err != nil {
			logger.Error("failed to create booked seats", applog.Error(err))
			return err
		}
		return nil
	})

	if err != nil {
		logger.Error("transaction error", applog.Error(err))
		return nil, err
	}

	logger.Info("create booking successfully", applog.Float64("total_price", totalPrice))
	return response.ToBookingResponse(booking), nil
}
