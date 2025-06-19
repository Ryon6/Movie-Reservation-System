package test

import (
	"fmt"
	"mrs/cmd/test/testutils"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/domain/cinema"
	applog "mrs/pkg/log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShowtimeManagementFlow(t *testing.T) {
	// 初始化测试服务器
	ts := testutils.NewTestServer(t)
	defer ts.Close()

	logger := ts.Logger.With(applog.String("Test", "TestShowtimeManagementFlow"))

	// 1. 管理员登录
	ts.AdminToken = ts.Login(t, "admin", "admin123")

	// 2. 创建影厅
	createHallReq := request.CreateCinemaHallRequest{
		Name:        "IMAX影厅",
		ScreenType:  "IMAX",
		SoundSystem: "Dolby Atmos",
		Seats: []*request.SeatRequest{
			{
				RowIdentifier: "A",
				SeatNumber:    "1",
				Type:          "VIP",
			},
			{
				RowIdentifier: "A",
				SeatNumber:    "2",
				Type:          "VIP",
			},
		},
	}

	resp, body := ts.DoRequest(t, http.MethodPost, "/api/v1/admin/cinema-halls", createHallReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var hallResp response.CinemaHallResponse
	testutils.ParseResponse(t, body, &hallResp)
	hallID := hallResp.ID

	logger.Debug("create cinema hall test", applog.Any("createHallReq", createHallReq))

	// 3. 创建电影
	createMovieReq := request.CreateMovieRequest{
		Title:           "复仇者联盟",
		Description:     "漫威超级英雄电影",
		GenreNames:      []string{"动作", "科幻"},
		DurationMinutes: 150,
		ReleaseDate:     time.Now(),
		Cast:            "小罗伯特·唐尼,克里斯·埃文斯",
		AgeRating:       "PG-13",
		Rating:          9.0,
	}

	resp, body = ts.DoRequest(t, http.MethodPost, "/api/v1/admin/movies", createMovieReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var movieResp response.MovieResponse
	testutils.ParseResponse(t, body, &movieResp)
	movieID := movieResp.ID

	logger.Debug("create movie test", applog.Any("movieResp", movieResp))

	// 4. 创建放映场次
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(time.Duration(createMovieReq.DurationMinutes) * time.Minute)
	createShowtimeReq := request.CreateShowtimeRequest{
		MovieID:      movieID,
		CinemaHallID: hallID,
		StartTime:    startTime,
		EndTime:      endTime,
		Price:        80.0,
	}

	resp, body = ts.DoRequest(t, http.MethodPost, "/api/v1/admin/showtimes", createShowtimeReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var showtimeResp response.ShowtimeResponse
	testutils.ParseResponse(t, body, &showtimeResp)
	showtimeID := showtimeResp.ID

	logger.Debug("create showtime test", applog.Any("showtimeResp", showtimeResp))

	// 5. 普通用户登录
	ts.UserToken = ts.Login(t, "user", "user123")

	// 6. 查询放映场次列表（按电影筛选）
	listShowtimesReq := request.ListShowtimesRequest{
		MovieID: movieID,
		PaginationRequest: request.PaginationRequest{
			Page:     1,
			PageSize: 10,
		},
	}
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/showtimes", listShowtimesReq, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp response.PaginatedShowtimeResponse
	testutils.ParseResponse(t, body, &listResp)
	assert.NotEmpty(t, listResp.Showtimes)
	assert.Equal(t, movieID, listResp.Showtimes[0].MovieID)

	logger.Debug("list showtime test", applog.Any("listResp", listResp))

	// 7. 查询特定放映场次详情
	resp, body = ts.DoRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/showtimes/%d", showtimeID), nil, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var detailResp response.ShowtimeResponse
	testutils.ParseResponse(t, body, &detailResp)
	assert.Equal(t, createShowtimeReq.Price, detailResp.Price)
	assert.Equal(t, createShowtimeReq.StartTime.Unix(), detailResp.StartTime.Unix())
	assert.Equal(t, createShowtimeReq.EndTime.Unix(), detailResp.EndTime.Unix())

	logger.Debug("get showtime detail test", applog.Any("detailResp", detailResp))

	// 8. 管理员更新放映场次信息
	updateShowtimeReq := request.UpdateShowtimeRequest{
		ID:    showtimeID,
		Price: 90.0, // 提高票价
	}

	resp, body = ts.DoRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/admin/showtimes/%d", showtimeID), updateShowtimeReq, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedShowtimeResp response.ShowtimeResponse
	testutils.ParseResponse(t, body, &updatedShowtimeResp)
	assert.Equal(t, updateShowtimeReq.Price, updatedShowtimeResp.Price)

	logger.Debug("update showtime test", applog.Any("updatedShowtimeResp", updatedShowtimeResp))

	// 9. 查询放映场次的座位表
	resp, body = ts.DoRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/showtimes/%d/seatmap", showtimeID), nil, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	logger.Debug("get showtime seats test", applog.Any("body", body), applog.Any("resp", resp))

	var seatMapResp response.SeatMapResponse
	testutils.ParseResponse(t, body, &seatMapResp)

	assert.Len(t, seatMapResp.Seats, 2) // 应该有两个座位
	if len(seatMapResp.Seats) > 0 {
		assert.Equal(t, cinema.SeatStatusAvailable, seatMapResp.Seats[0].Status) // 座位应该是可用的
	}

	// 10. 管理员删除放映场次
	resp, _ = ts.DoRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/showtimes/%d", showtimeID), nil, ts.AdminToken)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// 11. 验证放映场次已被删除
	resp, _ = ts.DoRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/showtimes/%d", showtimeID), nil, ts.UserToken)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
