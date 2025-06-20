package test

import (
	"fmt"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	applog "mrs/pkg/log"
	"mrs/test/e2e/testutils"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMovieManagementFlow(t *testing.T) {
	// 初始化测试服务器
	ts := testutils.NewTestServer(t)
	defer ts.Close()

	logger := ts.Logger.With(applog.String("Test", "TestMovieManagementFlow"))

	// 1. 管理员登录
	ts.AdminToken = ts.Login(t, "admin", "admin123")

	// 2. 创建电影类型
	createGenreReq := request.CreateGenreRequest{
		Name: "动作片",
	}
	resp, _ := ts.DoRequest(t, http.MethodPost, "/api/v1/admin/genres", createGenreReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	createGenreReq2 := request.CreateGenreRequest{
		Name: "科幻片",
	}
	resp, _ = ts.DoRequest(t, http.MethodPost, "/api/v1/admin/genres", createGenreReq2, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	logger.Debug("create genre test", applog.Any("createGenreReq", createGenreReq))

	// 3. 创建电影
	createMovieReq := request.CreateMovieRequest{
		Title:           "黑客帝国",
		Description:     "一部经典的科幻动作电影",
		GenreNames:      []string{"动作片", "科幻片"},
		DurationMinutes: 136,
		ReleaseDate:     time.Date(1999, 3, 31, 0, 0, 0, 0, time.UTC),
		Cast:            "基努·里维斯,劳伦斯·菲什伯恩",
		AgeRating:       "PG-13",
		Rating:          9.0,
		PosterURL:       "https://example.com/matrix.jpg",
	}

	resp, body := ts.DoRequest(t, http.MethodPost, "/api/v1/admin/movies", createMovieReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var movieResp response.MovieResponse
	testutils.ParseResponse(t, body, &movieResp)
	movieID := movieResp.ID

	logger.Debug("create movie test", applog.Any("movieResp", movieResp))

	// 4. 普通用户登录
	ts.UserToken = ts.Login(t, "user", "user123")

	// 5. 查询电影列表（按类型筛选）
	listMoviesReq := request.ListMovieRequest{
		GenreName: "科幻片",
		PaginationRequest: request.PaginationRequest{
			Page:     1,
			PageSize: 10,
		},
	}
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/movies", listMoviesReq, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp response.PaginatedMovieResponse
	testutils.ParseResponse(t, body, &listResp)
	assert.NotEmpty(t, listResp.Movies)
	assert.Equal(t, createMovieReq.Title, listResp.Movies[0].Title)

	logger.Debug("list movie test", applog.Any("listResp", listResp))

	// 6. 查询电影详情
	resp, body = ts.DoRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/movies/%d", movieID), nil, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var detailResp response.MovieResponse
	testutils.ParseResponse(t, body, &detailResp)
	assert.Equal(t, createMovieReq.Title, detailResp.Title)
	assert.Equal(t, createMovieReq.Description, detailResp.Description)
	assert.Equal(t, createMovieReq.DurationMinutes, detailResp.DurationMinutes)

	logger.Debug("get movie detail test", applog.Any("detailResp", detailResp))

	// 7. 管理员更新电影信息
	updateMovieReq := request.UpdateMovieRequest{
		ID:          movieID,
		Title:       "黑客帝国：重装上阵",
		Description: "经典科幻动作电影的续集",
		Rating:      8.8,
	}

	resp, body = ts.DoRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/admin/movies/%d", movieID), updateMovieReq, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedMovieResp response.MovieResponse
	testutils.ParseResponse(t, body, &updatedMovieResp)
	assert.Equal(t, updateMovieReq.Title, updatedMovieResp.Title)
	assert.Equal(t, updateMovieReq.Description, updatedMovieResp.Description)
	assert.InDelta(t, updateMovieReq.Rating, updatedMovieResp.Rating, 0.0001)

	logger.Debug("update movie test", applog.Any("updatedMovieResp", updatedMovieResp))

	// 8. 查询所有电影类型
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/genres", nil, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var genresResp response.ListAllGenresResponse
	testutils.ParseResponse(t, body, &genresResp)
	assert.Len(t, genresResp.Genres, 2)
	assert.Contains(t, []string{genresResp.Genres[0].Name, genresResp.Genres[1].Name}, "动作片")
	assert.Contains(t, []string{genresResp.Genres[0].Name, genresResp.Genres[1].Name}, "科幻片")

	logger.Debug("list genres test", applog.Any("genresResp", genresResp))

	// 9. 管理员更新电影类型
	updateGenreReq := request.UpdateGenreRequest{
		ID:   1, // 假设ID为1
		Name: "动作电影",
	}
	resp, _ = ts.DoRequest(t, http.MethodPut, "/api/v1/admin/genres/1", updateGenreReq, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	logger.Debug("update genre test", applog.Any("updateGenreReq", updateGenreReq))

	// 10. 管理员删除电影类型（应该失败，因为有电影关联）
	resp, _ = ts.DoRequest(t, http.MethodDelete, "/api/v1/admin/genres/1", nil, ts.AdminToken)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// 11. 管理员删除电影
	resp, _ = ts.DoRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/movies/%d", movieID), nil, ts.AdminToken)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// 12. 验证电影已被删除
	resp, _ = ts.DoRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/movies/%d", movieID), nil, ts.UserToken)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// 13. 现在可以删除电影类型了
	resp, _ = ts.DoRequest(t, http.MethodDelete, "/api/v1/admin/genres/1", nil, ts.AdminToken)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
