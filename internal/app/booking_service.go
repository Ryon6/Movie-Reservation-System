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
	"time"
)

type BookingService interface {
	CreateBooking(ctx context.Context, req *request.CreateBookingRequest) (*response.BookingResponse, error)
	ListBookings(ctx context.Context, req *request.ListBookingsRequest) (*response.ListBookingsResponse, error)
	GetBooking(ctx context.Context, req *request.GetBookingRequest) (*response.BookingResponse, error)
	CancelBooking(ctx context.Context, req *request.CancelBookingRequest) (*response.BookingResponse, error)
	ConfirmBooking(ctx context.Context, req *request.ConfirmBookingRequest) (*response.BookingResponse, error)
}

type bookingService struct {
	uow             shared.UnitOfWork
	bookingRepo     booking.BookingRepository
	showtimeRepo    showtime.ShowtimeRepository
	seatCache       cinema.SeatCache
	showtimeCache   showtime.ShowtimeCache
	showtimeService ShowtimeService
	lockProvider    lock.LockProvider
	logger          applog.Logger
}

func NewBookingService(
	uow shared.UnitOfWork,
	bookingRepo booking.BookingRepository,
	showtimeRepo showtime.ShowtimeRepository,
	seatCache cinema.SeatCache,
	showtimeCache showtime.ShowtimeCache,
	showtimeService ShowtimeService,
	lockProvider lock.LockProvider,
	logger applog.Logger) BookingService {

	return &bookingService{
		uow:             uow,
		bookingRepo:     bookingRepo,
		showtimeRepo:    showtimeRepo,
		seatCache:       seatCache,
		showtimeCache:   showtimeCache,
		showtimeService: showtimeService,
		lockProvider:    lockProvider,
		logger:          logger.With(applog.String("Service", "BookingService")),
	}
}

// CreateBooking 创建订单
func (s *bookingService) CreateBooking(ctx context.Context, req *request.CreateBookingRequest) (*response.BookingResponse, error) {
	logger := s.logger.With(applog.String("Method", "CreateBooking"))
	lockKey := cinema.GetShowtimeSeatsLockKey(vo.ShowtimeID(req.ShowtimeID))

	// 获取场次信息
	st, err := s.showtimeService.GetShowtime(ctx, &request.GetShowtimeRequest{
		ID: req.ShowtimeID,
	})
	if err != nil {
		logger.Error("failed to get showtime by service", applog.Error(err))
		return nil, err
	}

	// 检查场次是否已结束
	if st.EndTime.Before(time.Now()) {
		logger.Warn("showtime has ended", applog.String("end_time", st.EndTime.Format(time.DateTime)))
		return nil, showtime.ErrShowtimeEnded
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
	err = s.lockSeatsWithRetry(ctx, vo.ShowtimeID(req.ShowtimeID), seatIDs)
	if err != nil {
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
		booking, err = bookingRepo.Create(ctx, booking)
		if err != nil {
			logger.Error("failed to create booking", applog.Error(err))
			return err
		}

		// 将订单ID写入已预订的座位
		for i := range bookedSeats {
			bookedSeats[i].BookingID = booking.ID
		}

		// 创建已预订的座位
		bookedSeats, err = bookedSeatRepo.CreateBatch(ctx, bookedSeats)
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

// lockSeatsWithRetry 尝试锁定座位，如果缓存未初始化则初始化后重试
func (s *bookingService) lockSeatsWithRetry(ctx context.Context, showtimeID vo.ShowtimeID, seatIDs []vo.SeatID) error {
	logger := s.logger.With(applog.String("Method", "lockSeatsWithRetry"))

	err := s.seatCache.LockSeats(ctx, showtimeID, seatIDs)
	if err == nil {
		return nil
	}

	// 如果不是缓存缺失错误，直接返回错误
	if !errors.Is(err, shared.ErrCacheMissing) {
		return err
	}

	// 初始化座位表
	if err := s.showtimeService.InitSeatMap(ctx, showtimeID); err != nil {
		// 如果是锁已被其他进程占用，则进入退避重试逻辑
		if errors.Is(err, lock.ErrLockAlreadyAcquired) {
			logger.Warn("another process is initializing the seat map, will retry locking seats...",
				applog.Uint("showtimeID", uint(showtimeID)))

			for i := 0; i < lock.DefaultMaxRetries; i++ {
				time.Sleep(lock.DefaultBackoff)
				if err := s.seatCache.LockSeats(ctx, showtimeID, seatIDs); err == nil {
					logger.Info("successfully locked seats after waiting",
						applog.Uint("showtimeID", uint(showtimeID)))
					return nil
				}
			}

			// 如果重试多次后仍然失败
			logger.Error("failed to lock seats after retries",
				applog.Uint("showtimeID", uint(showtimeID)),
				applog.Int("retries", lock.DefaultMaxRetries))
			return lock.ErrRetryLockFailed
		}

		// 如果是其他初始化错误，则直接返回
		logger.Error("failed to init seat map", applog.Error(err))
		return err
	}

	// 初始化成功后，再次尝试锁定座位
	return s.seatCache.LockSeats(ctx, showtimeID, seatIDs)
}

// ListBookings 查询订单列表
func (s *bookingService) ListBookings(ctx context.Context, req *request.ListBookingsRequest) (*response.ListBookingsResponse, error) {
	logger := s.logger.With(applog.String("Method", "ListBookings"))

	domainOptions := req.ToDomain()
	bookings, totalCount, err := s.bookingRepo.List(ctx, domainOptions)
	if err != nil {
		logger.Error("failed to list bookings", applog.Error(err))
		return nil, err
	}
	response := response.ToListBookingsResponse(bookings, int(totalCount), &req.PaginationRequest)
	return response, nil
}

// GetBooking 查询订单
func (s *bookingService) GetBooking(ctx context.Context, req *request.GetBookingRequest) (*response.BookingResponse, error) {
	logger := s.logger.With(applog.String("Method", "GetBooking"))

	booking, err := s.bookingRepo.FindByID(ctx, vo.BookingID(req.ID))
	if err != nil {
		logger.Error("failed to get booking", applog.Error(err))
		return nil, err
	}
	return response.ToBookingResponse(booking), nil
}

// CancelBooking 取消订单
func (s *bookingService) CancelBooking(ctx context.Context, req *request.CancelBookingRequest) (*response.BookingResponse, error) {
	logger := s.logger.With(applog.String("Method", "CancelBooking"))

	lockKey := cinema.GetShowtimeSeatsLockKey(vo.ShowtimeID(req.ID))
	lock, err := s.lockProvider.Acquire(ctx, lockKey, lock.DefaultLockTTL)
	if err != nil {
		logger.Error("failed to acquire lock", applog.Error(err))
		return nil, err
	}
	defer lock.Release(ctx)

	var bk *booking.Booking
	// 使用事务，保证操作的原子性
	err = s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		bookingRepo := provider.GetBookingRepository()
		bookedSeatRepo := provider.GetBookedSeatRepository()
		bk, err = bookingRepo.FindByID(ctx, vo.BookingID(req.ID))
		if err != nil {
			logger.Error("failed to get booking", applog.Error(err))
			return err
		}

		// 订单状态必须是pending
		if bk.Status != booking.BookingStatusPending {
			logger.Warn("booking is not pending", applog.String("status", string(bk.Status)))
			return booking.ErrBookingNotPending
		}

		// 底层已预加载座位信息
		seatIDs := make([]vo.SeatID, len(bk.BookedSeats))
		for i, bookedSeat := range bk.BookedSeats {
			seatIDs[i] = bookedSeat.SeatID
		}

		bk.Cancel()
		if err = bookingRepo.Update(ctx, bk); err != nil {
			logger.Error("failed to update booking", applog.Error(err))
			return err
		}

		if err = bookedSeatRepo.DeleteByBookingID(ctx, vo.BookingID(req.ID)); err != nil {
			logger.Error("failed to delete booked seats", applog.Error(err))
			return err
		}

		// 释放座位锁
		if err = s.seatCache.ReleaseSeats(ctx, vo.ShowtimeID(req.ID), seatIDs); err != nil {
			logger.Error("failed to release seats", applog.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to cancel booking", applog.Error(err))
		return nil, err
	}

	logger.Info("cancel booking successfully", applog.String("status", string(bk.Status)))
	return response.ToBookingResponse(bk), nil
}

// ConfirmBooking 确认订单（简单实现，后续需要对接支付系统）
func (s *bookingService) ConfirmBooking(ctx context.Context, req *request.ConfirmBookingRequest) (*response.BookingResponse, error) {
	logger := s.logger.With(applog.String("Method", "ConfirmBooking"))

	var bk *booking.Booking
	var err error
	err = s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		bookingRepo := provider.GetBookingRepository()
		bk, err = bookingRepo.FindByID(ctx, vo.BookingID(req.ID))
		if err != nil {
			logger.Error("failed to get booking", applog.Error(err))
			return err
		}

		if bk.Status != booking.BookingStatusPending {
			logger.Warn("booking is not pending", applog.String("status", string(bk.Status)))
			return booking.ErrBookingNotPending
		}

		bk.Confirm()
		if err = bookingRepo.Update(ctx, bk); err != nil {
			logger.Error("failed to update booking", applog.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to confirm booking", applog.Error(err))
		return nil, err
	}

	logger.Info("confirm booking successfully", applog.String("status", string(bk.Status)))
	return response.ToBookingResponse(bk), nil
}
