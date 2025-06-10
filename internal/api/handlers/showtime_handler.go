package handlers

import (
	"errors"
	"mrs/internal/api/dto/request"
	"mrs/internal/app"
	"mrs/internal/domain/showtime"
	applog "mrs/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShowtimeHandler struct {
	showtimeService app.ShowtimeService
	logger          applog.Logger
}

func NewShowtimeHandler(showtimeService app.ShowtimeService, logger applog.Logger) *ShowtimeHandler {
	return &ShowtimeHandler{showtimeService: showtimeService, logger: logger.With(applog.String("Handler", "ShowtimeHandler"))}
}

// 创建放映场次 POST /api/v1/admin/showtimes
func (h *ShowtimeHandler) CreateShowtime(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "CreateShowtime"))
	var req request.CreateShowtimeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("failed to bind request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	showtimeResp, err := h.showtimeService.CreateShowtime(ctx, &req)
	if err != nil {
		logger.Error("failed to create showtime", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("showtime created successfully", applog.Uint("showtime_id", uint(showtimeResp.ID)))
	ctx.JSON(http.StatusOK, showtimeResp)
}

// 获取放映场次（详情） GET /api/v1/showtimes/:id
func (h *ShowtimeHandler) GetShowtime(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "GetShowtime"))
	id, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to get id from path", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req := request.GetShowtimeRequest{ID: id}

	showtimeResp, err := h.showtimeService.GetShowtime(ctx, &req)
	if err != nil {
		if errors.Is(err, showtime.ErrShowtimeNotFound) {
			logger.Warn("showtime not found")
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		logger.Error("failed to get showtime", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("showtime retrieved successfully", applog.Uint("showtime_id", uint(showtimeResp.ID)))
	ctx.JSON(http.StatusOK, showtimeResp)
}

// 列出放映场次（分页），支持过滤 GET /api/v1/showtimes
func (h *ShowtimeHandler) ListShowtimes(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "ListShowtimes"))
	var req request.ListShowtimesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logger.Error("failed to bind request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	showtimeResp, err := h.showtimeService.ListShowtimes(ctx, &req)
	if err != nil {
		logger.Error("failed to list showtimes", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("list showtimes successfully")
	ctx.JSON(http.StatusOK, showtimeResp)
}

// 更新放映场次 PUT /api/v1/admin/showtimes/:id
func (h *ShowtimeHandler) UpdateShowtime(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "UpdateShowtime"))
	var req request.UpdateShowtimeRequest
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

	showtimeResp, err := h.showtimeService.UpdateShowtime(ctx, &req)
	if err != nil {
		if errors.Is(err, showtime.ErrShowtimeNotFound) {
			logger.Warn("showtime not found")
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		logger.Error("failed to update showtime", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("showtime updated successfully", applog.Uint("showtime_id", uint(showtimeResp.ID)))
	ctx.JSON(http.StatusOK, showtimeResp)
}

// 删除放映场次 DELETE /api/v1/admin/showtimes/:id
func (h *ShowtimeHandler) DeleteShowtime(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "DeleteShowtime"))
	id, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to get id from path", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req := request.DeleteShowtimeRequest{ID: id}

	if err := h.showtimeService.DeleteShowtime(ctx, &req); err != nil {
		if errors.Is(err, showtime.ErrShowtimeNotFound) {
			logger.Warn("showtime not found")
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		logger.Error("failed to delete showtime", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("showtime deleted successfully", applog.Uint("showtime_id", uint(req.ID)))
	ctx.JSON(http.StatusNoContent, nil)
}
