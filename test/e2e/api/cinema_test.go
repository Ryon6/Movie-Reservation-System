package test

import (
	"fmt"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	applog "mrs/pkg/log"
	"mrs/test/e2e/testutils"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCinemaHallFlow(t *testing.T) {
	// 初始化测试服务器
	ts := testutils.NewTestServer(t)
	defer ts.Close()

	logger := ts.Logger.With(applog.String("Test", "TestCinemaHallFlow"))

	// 1. 管理员登录
	ts.AdminToken = ts.Login(t, "admin", "admin123")

	// 2. 创建影厅
	createReq := request.CreateCinemaHallRequest{
		Name:        "IMAX大厅",
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

	resp, body := ts.DoRequest(t, http.MethodPost, "/api/v1/admin/cinema-halls", createReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var hallResp response.CinemaHallResponse
	testutils.ParseResponse(t, body, &hallResp)
	assert.Equal(t, createReq.Name, hallResp.Name)
	assert.Equal(t, createReq.ScreenType, hallResp.ScreenType)
	assert.Equal(t, createReq.SoundSystem, hallResp.SoundSystem)
	assert.Len(t, hallResp.Seats, 2)

	hallID := hallResp.ID
	logger.Debug("create cinema hall test", applog.Uint("hall_id", hallID), applog.Any("hallResp", hallResp))

	// 3. 更新影厅信息
	updateReq := request.UpdateCinemaHallRequest{
		ID:          hallID,
		Name:        "升级IMAX厅",
		ScreenType:  "IMAX 3D",
		SoundSystem: "Dolby Atmos 7.1",
	}

	resp, body = ts.DoRequest(t, http.MethodPut, "/api/v1/admin/cinema-halls/"+fmt.Sprintf("%d", hallID), updateReq, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedHallResp response.CinemaHallResponse
	testutils.ParseResponse(t, body, &updatedHallResp)
	assert.Equal(t, updateReq.Name, updatedHallResp.Name)
	assert.Equal(t, updateReq.ScreenType, updatedHallResp.ScreenType)
	assert.Equal(t, updateReq.SoundSystem, updatedHallResp.SoundSystem)

	logger.Debug("update cinema hall test", applog.Any("updatedHallResp", updatedHallResp))

	// 4. 普通用户登录
	ts.UserToken = ts.Login(t, "user", "user123")

	// 5. 查询影厅列表
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/cinema-halls", nil, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp response.ListAllCinemaHallsResponse
	testutils.ParseResponse(t, body, &listResp)
	assert.NotEmpty(t, listResp.CinemaHalls)
	assert.Equal(t, updateReq.Name, listResp.CinemaHalls[0].Name)

	logger.Debug("list cinema halls test", applog.Any("listResp", listResp))

	// 6. 查询特定影厅详情
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/cinema-halls/"+fmt.Sprintf("%d", hallID), nil, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var detailResp response.CinemaHallResponse
	testutils.ParseResponse(t, body, &detailResp)
	assert.Equal(t, updateReq.Name, detailResp.Name)
	assert.Equal(t, updateReq.ScreenType, detailResp.ScreenType)
	assert.Equal(t, updateReq.SoundSystem, detailResp.SoundSystem)
	assert.Len(t, detailResp.Seats, 2)

	logger.Debug("get cinema hall detail test", applog.Any("detailResp", detailResp))

	// 7. 管理员删除影厅
	resp, _ = ts.DoRequest(t, http.MethodDelete, "/api/v1/admin/cinema-halls/"+fmt.Sprintf("%d", hallID), nil, ts.AdminToken)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// 8. 验证影厅已被删除
	resp, _ = ts.DoRequest(t, http.MethodGet, "/api/v1/cinema-halls/"+fmt.Sprintf("%d", hallID), nil, ts.UserToken)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
