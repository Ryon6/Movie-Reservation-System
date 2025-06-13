package handlers

import (
	"mrs/internal/api/dto/request"
	"mrs/internal/api/middleware"
	"mrs/internal/app"
	applog "mrs/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookingService app.BookingService
	logger         applog.Logger
}

func NewBookingHandler(bookingService app.BookingService, logger applog.Logger) *BookingHandler {
	return &BookingHandler{bookingService: bookingService, logger: logger.With(applog.String("Handler", "BookingHandler"))}
}

// 创建订单 POST /api/v1/bookings
func (h *BookingHandler) CreateBooking(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "CreateBooking"))

	var req request.CreateBookingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("failed to bind request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookingResp, err := h.bookingService.CreateBooking(ctx, &req)
	if err != nil {
		logger.Error("failed to create booking", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("create booking successfully", applog.Uint("booking_id", uint(bookingResp.ID)))
	ctx.JSON(http.StatusOK, bookingResp)
}

// 查询订单列表 GET /api/v1/bookings
func (h *BookingHandler) ListBookings(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "ListBookings"))

	var req request.ListBookingsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logger.Error("failed to bind request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := ctx.Get(middleware.UserIDKey)
	if !exists {
		logger.Error("user id not found")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
		return
	}
	req.UserID = userID.(uint)

	bookingsResp, err := h.bookingService.ListBookings(ctx, &req)
	if err != nil {
		logger.Error("failed to list bookings", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("list bookings successfully", applog.Uint("user_id", uint(req.UserID)))
	ctx.JSON(http.StatusOK, bookingsResp)
}

// 获取订单 GET /api/v1/bookings/:id
func (h *BookingHandler) GetBooking(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "GetBooking"))

	var req request.GetBookingRequest
	bookingID, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to get booking id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = bookingID

	bookingResp, err := h.bookingService.GetBooking(ctx, &req)
	if err != nil {
		logger.Error("failed to get booking", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("get booking successfully", applog.Uint("booking_id", uint(bookingResp.ID)))
	ctx.JSON(http.StatusOK, bookingResp)
}

// 取消订单 POST /api/v1/bookings/:id/cancel
func (h *BookingHandler) CancelBooking(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "CancelBooking"))

	bookingID, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to get booking id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req := request.CancelBookingRequest{ID: bookingID}

	bookingResp, err := h.bookingService.CancelBooking(ctx, &req)
	if err != nil {
		logger.Error("failed to cancel booking", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("cancel booking successfully", applog.Uint("booking_id", uint(bookingResp.ID)))
	ctx.JSON(http.StatusOK, bookingResp)
}

// 确认订单 POST /api/v1/bookings/:id/confirm
func (h *BookingHandler) ConfirmBooking(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "ConfirmBooking"))

	bookingID, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to get booking id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req := request.ConfirmBookingRequest{ID: bookingID}

	bookingResp, err := h.bookingService.ConfirmBooking(ctx, &req)
	if err != nil {
		logger.Error("failed to confirm booking", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("confirm booking successfully", applog.Uint("booking_id", uint(bookingResp.ID)))
	ctx.JSON(http.StatusOK, bookingResp)
}
