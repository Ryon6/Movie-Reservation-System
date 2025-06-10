package handlers

import (
	"errors"
	"mrs/internal/api/dto/request"
	"mrs/internal/app"
	"mrs/internal/domain/cinema"
	applog "mrs/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CinemaHandler struct {
	cinemaService app.CinemaService
	logger        applog.Logger
}

func NewCinemaHandler(cinemaService app.CinemaService, logger applog.Logger) *CinemaHandler {
	return &CinemaHandler{cinemaService: cinemaService, logger: logger.With(applog.String("Handler", "CinemaHandler"))}
}

// 创建影厅 POST /api/v1/admin/cinema-halls
func (h *CinemaHandler) CreateCinemaHall(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "CreateCinemaHall"))
	var req request.CreateCinemaHallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("failed to bind request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cinemaHallResp, err := h.cinemaService.CreateCinemaHall(ctx, &req)
	if err != nil {
		logger.Error("failed to create cinema hall", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("cinema hall created successfully", applog.Uint("cinema_hall_id", uint(cinemaHallResp.ID)))
	ctx.JSON(http.StatusOK, cinemaHallResp)
}

// 获取影厅 GET /api/v1/cinema-halls/:id
func (h *CinemaHandler) GetCinemaHall(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "GetCinemaHall"))
	var req request.GetCinemaHallRequest
	id, err := getIDFromPath(ctx)
	if err != nil {
		if errors.Is(err, cinema.ErrCinemaHallNotFound) {
			logger.Warn("cinema hall not found")
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		logger.Error("failed to get id from path", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = id

	cinemaHallResp, err := h.cinemaService.GetCinemaHall(ctx, &req)
	if err != nil {
		logger.Error("failed to get cinema hall", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("cinema hall retrieved successfully", applog.Uint("cinema_hall_id", uint(cinemaHallResp.ID)))
	ctx.JSON(http.StatusOK, cinemaHallResp)
}

// 获取所有影厅 GET /api/v1/cinema-halls
func (h *CinemaHandler) ListAllCinemaHalls(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "ListAllCinemaHalls"))

	cinemaHallResp, err := h.cinemaService.ListAllCinemaHalls(ctx)
	if err != nil {
		logger.Error("failed to list all cinema halls", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("list all cinema halls successfully")
	ctx.JSON(http.StatusOK, cinemaHallResp)
}

// 更新影厅 PUT /api/v1/admin/cinema-halls/:id
func (h *CinemaHandler) UpdateCinemaHall(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "UpdateCinemaHall"))
	var req request.UpdateCinemaHallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("failed to bind request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to get id from path", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = id

	cinemaHallResp, err := h.cinemaService.UpdateCinemaHall(ctx, &req)
	if err != nil {
		if errors.Is(err, cinema.ErrCinemaHallNotFound) {
			logger.Warn("cinema hall not found")
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		logger.Error("failed to update cinema hall", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("cinema hall updated successfully", applog.Uint("cinema_hall_id", uint(cinemaHallResp.ID)))
	ctx.JSON(http.StatusOK, cinemaHallResp)
}

// 删除影厅 DELETE /api/v1/admin/cinema-halls/:id
func (h *CinemaHandler) DeleteCinemaHall(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "DeleteCinemaHall"))
	var req request.DeleteCinemaHallRequest
	id, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to get id from path", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = id

	if err := h.cinemaService.DeleteCinemaHall(ctx, &req); err != nil {
		if errors.Is(err, cinema.ErrCinemaHallNotFound) {
			logger.Warn("cinema hall not found")
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		logger.Error("failed to delete cinema hall", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("cinema hall deleted successfully", applog.Uint("cinema_hall_id", uint(req.ID)))
	ctx.JSON(http.StatusOK, gin.H{"message": "cinema hall deleted successfully"})
}
