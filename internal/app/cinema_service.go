package app

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/shared"
	applog "mrs/pkg/log"
)

type CinemaService interface {
	CreateCinemaHall(ctx context.Context, req *request.CreateCinemaHallRequest) (*response.CinemaHallResponse, error)
	GetCinemaHall(ctx context.Context, req *request.GetCinemaHallRequest) (*response.CinemaHallResponse, error)
	CreateHallSeats(ctx context.Context, req *request.CreateHallSeatsRequest) (*response.CinemaHallResponse, error)
}

type cinemaService struct {
	uow            shared.UnitOfWork
	seatRepo       cinema.SeatRepository
	cinemaHallRepo cinema.CinemaHallRepository
	logger         applog.Logger
}

func NewCinemaService(
	uow shared.UnitOfWork,
	seatRepo cinema.SeatRepository,
	cinemaHallRepo cinema.CinemaHallRepository,
	logger applog.Logger,
) *cinemaService {
	return &cinemaService{
		uow:            uow,
		seatRepo:       seatRepo,
		cinemaHallRepo: cinemaHallRepo,
		logger:         logger.With(applog.String("Service", "CinemaHallService")),
	}
}

// 创建影厅
func (s *cinemaService) CreateCinemaHall(ctx context.Context, req *request.CreateCinemaHallRequest) (*response.CinemaHallResponse, error) {
	logger := s.logger.With(applog.String("Method", "CreateCinemaHall"), applog.String("cinema_hall_name", req.Name))

	cinemaHall := req.ToDomain()
	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		var err error
		cinemaHall, err = provider.GetCinemaHallRepository().Create(ctx, cinemaHall)
		if err != nil {
			if errors.Is(err, cinema.ErrCinemaHallAlreadyExists) {
				logger.Warn("cinema hall already exists", applog.Error(err))
				return fmt.Errorf("%w: %w", cinema.ErrCinemaHallAlreadyExists, err)
			}
			logger.Error("failed to create cinema hall", applog.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to create cinema hall", applog.Error(err))
		return nil, err
	}

	logger.Info("create cinema hall successfully", applog.Uint("cinema_hall_id", uint(cinemaHall.ID)))
	return response.ToCinemaHallResponse(cinemaHall), nil
}

// 获取影厅
func (s *cinemaService) GetCinemaHall(ctx context.Context, req *request.GetCinemaHallRequest) (*response.CinemaHallResponse, error) {
	logger := s.logger.With(applog.String("Method", "GetCinemaHall"), applog.Uint("cinema_hall_id", req.ID))

	cinemaHall, err := s.cinemaHallRepo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, cinema.ErrCinemaHallNotFound) {
			logger.Warn("cinema hall not found", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", cinema.ErrCinemaHallNotFound, err)
		}
		logger.Error("failed to get cinema hall", applog.Error(err))
		return nil, err
	}

	logger.Info("get cinema hall successfully", applog.Uint("cinema_hall_id", uint(cinemaHall.ID)))
	return response.ToCinemaHallResponse(cinemaHall), nil
}

// 创建影厅座位
func (s *cinemaService) CreateHallSeats(ctx context.Context, req *request.CreateHallSeatsRequest) (*response.CinemaHallResponse, error) {
	logger := s.logger.With(applog.String("Method", "CreateHallSeats"), applog.Uint("cinema_hall_id", req.CinemaHallID))

	// 获取影厅，判断影厅是否存在
	cinemaHall, err := s.cinemaHallRepo.FindByID(ctx, req.CinemaHallID)
	if err != nil {
		if errors.Is(err, cinema.ErrCinemaHallNotFound) {
			logger.Warn("cinema hall not found", applog.Error(err))
			return nil, fmt.Errorf("%w: %w", cinema.ErrCinemaHallNotFound, err)
		}
		logger.Error("failed to get cinema hall", applog.Error(err))
		return nil, err
	}

	seats := req.ToDomain()
	err = s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		seats, err = provider.GetSeatRepository().CreateBatch(ctx, seats)
		if err != nil {
			logger.Error("failed to create seats", applog.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to create seats", applog.Error(err))
		return nil, err
	}

	cinemaHall.Seats = seats
	logger.Info("create seats successfully", applog.Int("seat count", len(seats)))
	return response.ToCinemaHallResponse(cinemaHall), nil
}
