package handlers

import (
	"errors"
	"mrs/internal/api/dto/request"
	"mrs/internal/app"
	"mrs/internal/domain/movie"
	applog "mrs/pkg/log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MovieHandler struct {
	movieService app.MovieService
	logger       applog.Logger
}

func NewMovieHandler(movieService app.MovieService, logger applog.Logger) *MovieHandler {
	return &MovieHandler{
		movieService: movieService,
		logger:       logger.With(applog.String("Handler", "MovieHandler")),
	}
}

// 创建电影 POST /api/v1/admin/movies
func (h *MovieHandler) CreateMovie(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "CreateMovie"))
	var req request.CreateMovieRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind create movie request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movieResp, err := h.movieService.CreateMovie(ctx, &req)
	if err != nil {
		// 电影可能已存在
		if errors.Is(err, movie.ErrMovieAlreadyExists) {
			logger.Warn("movie already exists", applog.Error(err))
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		logger.Error("failed to create movie", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("movie created successfully", applog.Uint("movie_id", uint(movieResp.ID)))
	ctx.JSON(http.StatusCreated, movieResp)
}

// 获取电影 GET /api/v1/movies/{movieId}
func (h *MovieHandler) GetMovie(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "GetMovie"))
	movieId := ctx.Param("id")
	movieID, err := strconv.ParseUint(movieId, 10, 32)
	if err != nil {
		logger.Warn("failed to parse movie id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movieResp, err := h.movieService.GetMovie(ctx, &request.GetMovieRequest{ID: uint(movieID)})
	if err != nil {
		logger.Error("failed to get movie", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("movie retrieved successfully", applog.Uint("movie_id", uint(movieResp.ID)))
	ctx.JSON(http.StatusOK, movieResp)
}

// 更新电影 PUT /api/v1/admin/movies/{movieId}
func (h *MovieHandler) UpdateMovie(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "UpdateMovie"))
	movieId := ctx.Param("id")
	movieID, err := strconv.ParseUint(movieId, 10, 32)
	if err != nil {
		logger.Warn("failed to parse movie id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req request.UpdateMovieRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		logger.Warn("failed to bind update movie request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = uint(movieID)

	if err := h.movieService.UpdateMovie(ctx, &req); err != nil {
		logger.Error("failed to update movie", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("movie updated successfully", applog.Uint("movie_id", uint(req.ID)))
	ctx.JSON(http.StatusOK, gin.H{"message": "movie updated successfully"})
}

// 删除电影 DELETE /api/v1/admin/movies/{movieId}
func (h *MovieHandler) DeleteMovie(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "DeleteMovie"))
	movieId := ctx.Param("id")
	movieID, err := strconv.ParseUint(movieId, 10, 32)
	if err != nil {
		logger.Warn("failed to parse movie id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.movieService.DeleteMovie(ctx, &request.DeleteMovieRequest{ID: uint(movieID)}); err != nil {
		logger.Error("failed to delete movie", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("movie deleted successfully", applog.Uint("movie_id", uint(movieID)))
	ctx.JSON(http.StatusNoContent, nil)
}

// 获取电影列表 GET /api/v1/movies
func (h *MovieHandler) ListMovies(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "ListMovies"))
	var req request.ListMovieRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind list movie request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movieResp, err := h.movieService.ListMovies(ctx, &req)
	if err != nil {
		logger.Error("failed to list movies", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("movies listed successfully", applog.Int("total", len(movieResp.Movies)))
	ctx.JSON(http.StatusOK, movieResp)
}

// 创建类型 POST /api/v1/admin/genres
func (h *MovieHandler) CreateGenre(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "CreateGenre"))
	var req request.CreateGenreRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind create genre request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	genreResp, err := h.movieService.CreateGenre(ctx, &req)
	if err != nil {
		logger.Error("failed to create genre", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("genre created successfully", applog.String("genre_name", genreResp.Name))
	ctx.JSON(http.StatusCreated, genreResp)
}

// 更新类型(根据ID更新Name) PUT /api/v1/admin/genres/{genreId}
func (h *MovieHandler) UpdateGenre(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "UpdateGenre"))
	genreId := ctx.Param("id")
	genreID, err := strconv.ParseUint(genreId, 10, 32)
	if err != nil {
		logger.Warn("failed to parse genre id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req request.UpdateGenreRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind update genre request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = uint(genreID)

	genreResp, err := h.movieService.UpdateGenre(ctx, &req)
	if err != nil {
		logger.Error("failed to update genre", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("genre updated successfully", applog.String("genre_name", genreResp.Name))
	ctx.JSON(http.StatusOK, genreResp)
}

// 删除类型 DELETE /api/v1/admin/genres/{genreId}
func (h *MovieHandler) DeleteGenre(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "DeleteGenre"))
	genreId := ctx.Param("id")
	genreID, err := strconv.ParseUint(genreId, 10, 32)
	if err != nil {
		logger.Warn("failed to parse genre id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.movieService.DeleteGenre(ctx, &request.DeleteGenreRequest{ID: uint(genreID)}); err != nil {
		logger.Error("failed to delete genre", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("genre deleted successfully", applog.Uint("genre_id", uint(genreID)))
	ctx.JSON(http.StatusNoContent, nil)
}

// 获取所有类型 GET /api/v1/genres
func (h *MovieHandler) ListAllGenres(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "ListGenres"))
	genreResp, err := h.movieService.ListAllGenres(ctx)
	if err != nil {
		logger.Error("failed to list genres", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("genres listed successfully", applog.Int("total", len(genreResp.Genres)))
	ctx.JSON(http.StatusOK, genreResp)
}
