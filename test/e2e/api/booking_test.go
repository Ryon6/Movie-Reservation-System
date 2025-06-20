package test

import (
	"fmt"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/domain/booking"
	"mrs/test/e2e/testutils"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMovieBookingFlow(t *testing.T) {
	// 初始化测试服务器
	ts := testutils.NewTestServer(t)
	defer ts.Close()

	// 1. 管理员登录
	ts.AdminToken = ts.Login(t, "admin", "admin123")

	// 2. 创建影厅
	createHallReq := request.CreateCinemaHallRequest{
		Name:        "标准影厅1",
		ScreenType:  "2D",
		SoundSystem: "Dolby 5.1",
		Seats: []*request.SeatRequest{
			{
				RowIdentifier: "A",
				SeatNumber:    "1",
				Type:          "STANDARD",
			},
			{
				RowIdentifier: "A",
				SeatNumber:    "2",
				Type:          "STANDARD",
			},
		},
	}

	resp, body := ts.DoRequest(t, http.MethodPost, "/api/v1/admin/cinema-halls", createHallReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var hallResp response.CinemaHallResponse
	testutils.ParseResponse(t, body, &hallResp)
	hallID := hallResp.ID

	// 3. 创建电影
	createMovieReq := request.CreateMovieRequest{
		Title:           "测试电影",
		Description:     "这是一部测试电影",
		GenreNames:      []string{"动作", "冒险"},
		DurationMinutes: 120,
		ReleaseDate:     time.Now(),
		Cast:            "演员1,演员2",
		AgeRating:       "PG-13",
		Rating:          8.5,
	}

	resp, body = ts.DoRequest(t, http.MethodPost, "/api/v1/admin/movies", createMovieReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var movieResp response.MovieResponse
	testutils.ParseResponse(t, body, &movieResp)
	movieID := movieResp.ID

	// 4. 创建放映场次
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(time.Duration(createMovieReq.DurationMinutes) * time.Minute)
	createShowtimeReq := request.CreateShowtimeRequest{
		MovieID:      movieID,
		CinemaHallID: hallID,
		StartTime:    startTime,
		EndTime:      endTime,
		Price:        50.0,
	}

	resp, body = ts.DoRequest(t, http.MethodPost, "/api/v1/admin/showtimes", createShowtimeReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var showtimeResp response.ShowtimeResponse
	testutils.ParseResponse(t, body, &showtimeResp)
	showtimeID := showtimeResp.ID

	// 5. 用户登录
	ts.UserToken = ts.Login(t, "user", "user123")

	// 6. 查询电影列表
	listMoviesReq := request.ListMovieRequest{
		PaginationRequest: request.PaginationRequest{
			Page:     1,
			PageSize: 10,
		},
	}
	resp, _ = ts.DoRequest(t, http.MethodGet, "/api/v1/movies", listMoviesReq, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 7. 查询特定电影的放映场次
	resp, _ = ts.DoRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/showtimes/%d", movieID), nil, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 8. 创建订单
	createBookingReq := request.CreateBookingRequest{
		ShowtimeID: showtimeID,
		SeatIDs:    []uint{hallResp.Seats[0].ID}, // 预订第一个座位
	}

	resp, body = ts.DoRequest(t, http.MethodPost, "/api/v1/bookings", createBookingReq, ts.UserToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var bookingResp response.BookingResponse
	testutils.ParseResponse(t, body, &bookingResp)
	bookingID := bookingResp.ID

	// 9. 确认订单
	resp, _ = ts.DoRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/bookings/%d/confirm", bookingID), nil, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 10. 查询订单列表
	listBookingsReq := request.ListBookingsRequest{
		PaginationRequest: request.PaginationRequest{
			Page:     1,
			PageSize: 10,
		},
	}
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/bookings", listBookingsReq, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var listResp response.ListBookingsResponse
	testutils.ParseResponse(t, body, &listResp)
	assert.NotEmpty(t, listResp.Bookings)

	// 11. 查询特定订单详情
	resp, body = ts.DoRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/bookings/%d", bookingID), nil, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var detailResp response.BookingResponse
	testutils.ParseResponse(t, body, &detailResp)
	assert.Equal(t, string(booking.BookingStatusConfirmed), detailResp.Status)

	// 12. 尝试取消已确认的订单（应该失败）
	resp, _ = ts.DoRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/bookings/%d/cancel", bookingID), nil, ts.UserToken)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
