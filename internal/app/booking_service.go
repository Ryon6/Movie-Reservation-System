package app

import (
	"context"
	"errors"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/domain/booking"
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared/lock"
	"mrs/internal/domain/shared/vo"
	"mrs/internal/domain/showtime"
	applog "mrs/pkg/log"
)

type BookingService interface {
	CreateBooking(ctx context.Context, req *request.CreateBookingRequest) (*response.BookingResponse, error)
}

type bookingService struct {
	bookingRepo   booking.BookingRepository
	showtimeRepo  showtime.ShowtimeRepository
	seatCache     cinema.SeatCache
	showtimeCache showtime.ShowtimeCache
	lockProvider  lock.LockProvider
	logger        applog.Logger
}

func NewBookingService(
	bookingRepo booking.BookingRepository,
	showtimeRepo showtime.ShowtimeRepository,
	seatCache cinema.SeatCache,
	showtimeCache showtime.ShowtimeCache,
	lockProvider lock.LockProvider,
	logger applog.Logger) BookingService {

	return &bookingService{
		bookingRepo:   bookingRepo,
		showtimeRepo:  showtimeRepo,
		seatCache:     seatCache,
		showtimeCache: showtimeCache,
		lockProvider:  lockProvider,
		logger:        logger.With(applog.String("Service", "BookingService")),
	}
}

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
	if err := s.seatCache.LockSeats(ctx, vo.ShowtimeID(req.ShowtimeID), seatIDs); err != nil {
		if errors.Is(err, cinema.ErrSeatAlreadyLocked) {
			logger.Warn("seat already locked", applog.Error(err))
			return nil, err
		}
		logger.Error("failed to lock seats", applog.Error(err))
		return nil, err
	}

	// 创建预订，需要createBookedSeats和createBooking两个操作
	// TODO: 使用事务，确保两个操作要么都成功，要么都失败
	bookedSeats := make([]booking.BookedSeat, len(seatIDs))
	for i, seatID := range seatIDs {
		bookedSeats[i] = booking.BookedSeat{
			SeatID: seatID,
			Price:  st.Price,
		}
	}

	totalPrice := float64(len(seatIDs)) * st.Price
	booking := booking.NewBooking(vo.UserID(req.UserID), vo.ShowtimeID(req.ShowtimeID), bookedSeats, totalPrice)
	if _, err := s.bookingRepo.CreateBooking(ctx, booking); err != nil {
		logger.Error("failed to create booking", applog.Error(err))
		return nil, err
	}

	// TODO: 确认订单经过支付后，更新订单状态

	logger.Info("create booking successfully", applog.Float64("total_price", totalPrice))
	return response.ToBookingResponse(booking), nil
}
