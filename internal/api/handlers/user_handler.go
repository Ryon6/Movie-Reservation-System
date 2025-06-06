package handlers

import (
	"errors"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/api/middleware"
	"mrs/internal/app"
	"mrs/internal/domain/user"
	applog "mrs/pkg/log"
	"net/http"

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
func (h *UserHandler) RegisterUser(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "RegisterUser"))
	var req request.RegisterUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind register request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// 在service层判断DefaultRole是否为空
	usr, err := h.userService.RegisterUser(ctx, req.Username, req.Email, req.Password, req.DefaultRole)
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

	userResp := response.UserResponse{
		ID:       uint(usr.ID),
		Username: usr.Username,
		Email:    usr.Email,
		RoleName: usr.Role.Name,
	}
	logger.Info("user registered successfully", applog.Uint("user_id", uint(usr.ID)))
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

	usr, err := h.userService.GetUserByID(ctx, id)
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

	resp := response.UserResponse{
		ID:       uint(usr.ID),
		Username: usr.Username,
		Email:    usr.Email,
		RoleName: usr.Role.Name,
	}
	logger.Info("user profile retrieved successfully", applog.Uint("user_id", uint(usr.ID)))
	ctx.JSON(http.StatusOK, resp)
}
