package app

import (
	"context"
	"fmt"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/domain/booking"
	applog "mrs/pkg/log"
	"time"
)

type ReportService interface {
	GenerateSalesReport(ctx context.Context, req *request.GenerateSalesReportRequest) (*response.GenerateSalesReportResponse, error)
}

type reportService struct {
	logger      applog.Logger
	bookingRepo booking.BookingRepository
}

func NewReportService(
	logger applog.Logger,
	bookingRepo booking.BookingRepository,
) ReportService {
	return &reportService{
		logger:      logger.With(applog.String("Service", "ReportService")),
		bookingRepo: bookingRepo,
	}
}

func (s *reportService) GenerateSalesReport(ctx context.Context, req *request.GenerateSalesReportRequest) (*response.GenerateSalesReportResponse, error) {
	logger := s.logger.With(applog.String("Method", "GenerateSalesReport"))

	// 1. 准备查询选项
	options := &booking.SalesQueryOptions{
		MovieID:   req.MovieID,
		CinemaID:  req.CinemaID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	// 2. 获取销售统计数据
	stats, err := s.bookingRepo.GetSalesStatistics(ctx, options)
	if err != nil {
		logger.Error("failed to get sales statistics", applog.Error(err))
		return nil, fmt.Errorf("failed to get sales statistics: %w", err)
	}

	// 3. 构建响应
	now := time.Now()
	resp := &response.GenerateSalesReportResponse{
		ReportDate:    now.Format("2006-01-02 15:04:05"),
		StartDate:     req.StartDate.Format("2006-01-02"),
		EndDate:       req.EndDate.Format("2006-01-02"),
		TotalRevenue:  stats.TotalRevenue,
		TotalBookings: stats.TotalBookings,
	}

	logger.Info("generate sales report successfully")
	return resp, nil
}
