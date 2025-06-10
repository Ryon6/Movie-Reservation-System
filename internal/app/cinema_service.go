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
	ListAllCinemaHalls(ctx context.Context) (*response.ListAllCinemaHallsResponse, error)
	UpdateCinemaHall(ctx context.Context, req *request.UpdateCinemaHallRequest) (*response.CinemaHallResponse, error)
	DeleteCinemaHall(ctx context.Context, req *request.DeleteCinemaHallRequest) error
}

type cinemaService struct {
	uow            shared.UnitOfWork
	cinemaHallRepo cinema.CinemaHallRepository
	seatRepo       cinema.SeatRepository
	logger         applog.Logger
}

func NewCinemaService(
	uow shared.UnitOfWork,
	cinemaHallRepo cinema.CinemaHallRepository,
	seatRepo cinema.SeatRepository,
	logger applog.Logger,
) CinemaService {
	return &cinemaService{
		uow:            uow,
		cinemaHallRepo: cinemaHallRepo,
		seatRepo:       seatRepo,
		logger:         logger.With(applog.String("Service", "CinemaHallService")),
	}
}

// 创建影厅
func (s *cinemaService) CreateCinemaHall(ctx context.Context, req *request.CreateCinemaHallRequest) (*response.CinemaHallResponse, error) {
	logger := s.logger.With(applog.String("Method", "CreateCinemaHall"), applog.String("cinema_hall_name", req.Name))

	cinemaHall := req.ToDomain()
	if len(cinemaHall.Seats) == 0 {
		cinemaHall.Seats = cinema.GenerateDefaultSeats()
	}

	// 创建影厅时，需要创建座位，所以需要使用事务
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

		seats, err := provider.GetSeatRepository().CreateBatch(ctx, cinemaHall.Seats)
		if err != nil {
			logger.Error("failed to create seats", applog.Error(err))
			return err
		}
		cinemaHall.Seats = seats
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

// 获取所有影厅
func (s *cinemaService) ListAllCinemaHalls(ctx context.Context) (*response.ListAllCinemaHallsResponse, error) {
	logger := s.logger.With(applog.String("Method", "ListAllCinemaHalls"))

	cinemaHalls, err := s.cinemaHallRepo.ListAll(ctx)
	if err != nil {
		logger.Error("failed to list all cinema halls", applog.Error(err))
		return nil, err
	}

	logger.Info("list all cinema halls successfully", applog.Int("cinema_hall_count", len(cinemaHalls)))
	return response.ToListAllCinemaHallsResponse(cinemaHalls), nil
}

// 更新影厅
func (s *cinemaService) UpdateCinemaHall(ctx context.Context, req *request.UpdateCinemaHallRequest) (*response.CinemaHallResponse, error) {
	logger := s.logger.With(applog.String("Method", "UpdateCinemaHall"), applog.Uint("cinema_hall_id", req.ID))

	cinemaHall := req.ToDomain()
	// 影厅更新只涉及单条记录，不需要事务
	if err := s.cinemaHallRepo.Update(ctx, cinemaHall); err != nil {
		if errors.Is(err, cinema.ErrCinemaHallNotFound) {
			logger.Warn("cinema hall not found")
			return nil, err
		}
		logger.Error("failed to update cinema hall", applog.Error(err))
		return nil, err
	}

	logger.Info("update cinema hall successfully", applog.Uint("cinema_hall_id", req.ID))
	return response.ToCinemaHallResponse(cinemaHall), nil
}

// 删除影厅
func (s *cinemaService) DeleteCinemaHall(ctx context.Context, req *request.DeleteCinemaHallRequest) error {
	logger := s.logger.With(applog.String("Method", "DeleteCinemaHall"), applog.Uint("cinema_hall_id", req.ID))

	// 由于座位的外键约束为CASCADE，影厅删除时，座位也会被删除，所以需要使用事务
	err := s.uow.Execute(ctx, func(ctx context.Context, provider shared.RepositoryProvider) error {
		cinemaHallRepo := provider.GetCinemaHallRepository()
		// 无需先检查影厅是否存在，因为仓库底层实现会根据RowAffected判断记录是否存在
		if err := cinemaHallRepo.Delete(ctx, req.ID); err != nil {
			if errors.Is(err, cinema.ErrCinemaHallNotFound) {
				logger.Warn("cinema hall not found", applog.Error(err))
				return fmt.Errorf("%w: %w", cinema.ErrCinemaHallNotFound, err)
			}
			logger.Error("failed to delete cinema hall", applog.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("failed to delete cinema hall", applog.Error(err))
		return err
	}

	logger.Info("delete cinema hall successfully", applog.Uint("cinema_hall_id", req.ID))
	return nil
}
