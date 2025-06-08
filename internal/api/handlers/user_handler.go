package handlers

import (
	"errors"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/middleware"
	"mrs/internal/app"
	"mrs/internal/domain/user"
	applog "mrs/pkg/log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService app.UserService
	logger      applog.Logger
}

func NewUserHandler(userService app.UserService, logger applog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger.With(applog.String("Handler", "UserHandler")),
	}
}

// 处理注册请求
func (h *UserHandler) Register(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "Register"))
	var req request.RegisterUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind register request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// 在service层判断DefaultRole是否为空
	userResp, err := h.userService.Register(ctx, &req)
	if err != nil {
		// 用户可能已存在
		if errors.Is(err, user.ErrUserAlreadyExists) {
			logger.Warn("User registration conflict", applog.Error(err))
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		logger.Error("user registration service failed", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("user registered successfully")
	ctx.JSON(http.StatusOK, userResp)
}

func (h *UserHandler) GetUserProfile(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "GetUserProfile"))
	userID, exists := ctx.Get(middleware.UserIDKey)
	if !exists {
		logger.Error("userID not found in context, auth middleware might not have run or failed")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User ID not found in context"})
		return
	}

	id, ok := userID.(uint)
	if !ok {
		logger.Error("user_id in context is not of type uint", applog.Any("user_id_type", userID))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: User ID format issue"})
		return
	}

	userResp, err := h.userService.GetUserProfile(ctx, &request.GetUserRequest{ID: id})
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn("user profile not found", applog.Uint("user_id", id))
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User profile not found"})
			return
		}
		logger.Error("Failed to get user profile", applog.Uint("user_id", id), applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: User ID format issue"})
		return
	}

	logger.Info("user profile retrieved successfully", applog.Uint("user_id", id))
	ctx.JSON(http.StatusOK, userResp)
}

// 用户更新自身信息
func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "UpdateUser"))
	var req request.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind update request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	userID, exists := ctx.Get(middleware.UserIDKey)
	if !exists {
		logger.Error("userID not found in context, auth middleware might not have run or failed")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User ID not found in context"})
		return
	}

	id, ok := userID.(uint)
	if !ok {
		logger.Error("user_id in context is not of type uint", applog.Any("user_id_type", userID))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: User ID format issue"})
		return
	}
	req.ID = id

	userResp, err := h.userService.UpdateUserProfile(ctx, &req)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn("user not found", applog.Uint("user_id", id))
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		logger.Error("failed to update user", applog.Uint("user_id", id), applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	logger.Info("user profile updated successfully", applog.Uint("user_id", id))
	ctx.JSON(http.StatusOK, userResp)
}

// 管理员更新用户信息
func (h *UserHandler) AdminUpdateUser(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "AdminUpdateUser"))
	var req request.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind update request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	userID, exists := ctx.Params.Get("id")
	if !exists {
		logger.Error("user_id not found in params")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in params"})
		return
	}

	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		logger.Error("failed to parse user_id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	req.ID = uint(id)

	userResp, err := h.userService.UpdateUserProfile(ctx, &req)
	if err != nil {
		logger.Error("failed to update user", applog.Uint("user_id", uint(id)), applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("user profile updated successfully", applog.Uint("user_id", uint(id)))
	ctx.JSON(http.StatusOK, userResp)
}
