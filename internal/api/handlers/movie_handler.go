package handlers

import (
	"errors"
	"mrs/internal/api/dto/request"
	"mrs/internal/app"
	"mrs/internal/domain/movie"
	applog "mrs/pkg/log"
	"net/http"

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
	ctx.JSON(http.StatusOK, movieResp)
}

func (h *MovieHandler) GetMovie(ctx *gin.Context) {}

func (h *MovieHandler) UpdateMovie(ctx *gin.Context) {}

func (h *MovieHandler) DeleteMovie(ctx *gin.Context) {}

func (h *MovieHandler) ListMovies(ctx *gin.Context) {}
