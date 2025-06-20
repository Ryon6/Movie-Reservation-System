package test

import (
	"encoding/json"
	"fmt"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/test/e2e/testutils"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSalesReport(t *testing.T) {
	ts := testutils.NewTestServer(t)
	defer ts.Close()

	// logger := ts.Logger.With(applog.String("Test", "TestGenerateSalesReport"))

	ts.AdminToken = ts.Login(t, "admin", "admin123")
	ts.UserToken = ts.Login(t, "user", "user123")

	// 1. 准备测试数据
	// 1.1 创建电影
	movieReq := request.CreateMovieRequest{
		Title:           "Test Movie",
		GenreNames:      []string{"Action"},
		Description:     "Test Description",
		ReleaseDate:     time.Now(),
		DurationMinutes: 120,
		Rating:          8.5,
		PosterURL:       "http://example.com/poster.jpg",
		AgeRating:       "PG-13",
		Cast:            "Actor 1, Actor 2",
	}
	resp, body := ts.DoRequest(t, http.MethodPost, "/api/v1/admin/movies", movieReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var movieResp response.MovieResponse
	json.Unmarshal(body, &movieResp)

	// 1.2 创建影厅
	hallReq := request.CreateCinemaHallRequest{
		Name:        "Test Hall",
		ScreenType:  "2D",
		SoundSystem: "Dolby",
		Seats: []*request.SeatRequest{
			{RowIdentifier: "A", SeatNumber: "1", Type: "standard"},
			{RowIdentifier: "A", SeatNumber: "2", Type: "standard"},
		},
	}
	resp, body = ts.DoRequest(t, http.MethodPost, "/api/v1/admin/cinema-halls", hallReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var hallResp response.CinemaHallResponse
	json.Unmarshal(body, &hallResp)

	// 1.3 创建场次
	now := time.Now()
	showtimeReq := request.CreateShowtimeRequest{
		MovieID:      movieResp.ID,
		CinemaHallID: hallResp.ID,
		StartTime:    now.Add(time.Hour),
		EndTime:      now.Add(3 * time.Hour),
		Price:        50.0,
	}
	resp, body = ts.DoRequest(t, http.MethodPost, "/api/v1/admin/showtimes", showtimeReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var showtimeResp response.ShowtimeResponse
	json.Unmarshal(body, &showtimeResp)

	// 1.4 创建订单
	bookingReq := request.CreateBookingRequest{
		ShowtimeID: showtimeResp.ID,
		SeatIDs:    []uint{1, 2}, // 假设座位ID为1和2
	}
	resp, body = ts.DoRequest(t, http.MethodPost, "/api/v1/bookings", bookingReq, ts.UserToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var bookingResp response.BookingResponse
	json.Unmarshal(body, &bookingResp)

	// 1.5 确认订单
	confirmReq := request.ConfirmBookingRequest{
		ID: bookingResp.ID,
	}
	resp, _ = ts.DoRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/bookings/%d/confirm", bookingResp.ID), confirmReq, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 2. 测试生成销售报告
	// 2.1 测试无过滤条件的报告
	reportReq := request.GenerateSalesReportRequest{
		StartDate: now.Add(-24 * time.Hour),
		EndDate:   now.Add(24 * time.Hour),
	}
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/admin/reports/sales", reportReq, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var reportResp response.GenerateSalesReportResponse
	err := json.Unmarshal(body, &reportResp)
	assert.NoError(t, err)

	// 验证报告数据
	assert.Equal(t, 100.0, reportResp.TotalRevenue) // 2张票，每张50元
	assert.Equal(t, 1, reportResp.TotalBookings)

	// 2.2 测试按电影ID过滤的报告
	reportReq = request.GenerateSalesReportRequest{
		MovieID:   movieResp.ID,
		StartDate: now.Add(-24 * time.Hour),
		EndDate:   now.Add(24 * time.Hour),
	}
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/admin/reports/sales", reportReq, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.Unmarshal(body, &reportResp)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, reportResp.TotalRevenue)
	assert.Equal(t, 1, reportResp.TotalBookings)

	// 2.3 测试按影厅ID过滤的报告
	reportReq = request.GenerateSalesReportRequest{
		CinemaID:  hallResp.ID,
		StartDate: now.Add(-24 * time.Hour),
		EndDate:   now.Add(24 * time.Hour),
	}
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/admin/reports/sales", reportReq, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.Unmarshal(body, &reportResp)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, reportResp.TotalRevenue)
	assert.Equal(t, 1, reportResp.TotalBookings)

	// 2.4 测试日期范围外的报告
	reportReq = request.GenerateSalesReportRequest{
		StartDate: now.Add(48 * time.Hour),
		EndDate:   now.Add(72 * time.Hour),
	}
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/admin/reports/sales", reportReq, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.Unmarshal(body, &reportResp)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, reportResp.TotalRevenue)
	assert.Equal(t, 0, reportResp.TotalBookings)

	// 2.5 测试权限控制
	resp, _ = ts.DoRequest(t, http.MethodGet, "/api/v1/admin/reports/sales", reportReq, ts.UserToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
