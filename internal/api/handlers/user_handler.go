// TODO: username already exists 等错误直接定义好用于调用
// TODO: 自顶向下修改错误的使用方法

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
		logger:      logger,
	}
}

// 处理注册请求
func (h *UserHandler) RegisterUser(ctx *gin.Context) {
	logger := h.logger.With(applog.String("handler", "UserHandler.RegisterUser"))
	var req request.RegisterUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind register request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// 在service层判断DefaultRole是否为空
	usrPrf, err := h.userService.RegisterUser(ctx, req.Username, req.Email, req.Password, req.DefaultRole)
	if err != nil {
		if errors.Is(err, user.ErrEmailExists) || errors.Is(err, user.ErrUsernameExists) {
			logger.Warn("User registration conflict", applog.Error(err))
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		logger.Error("user registration service failed", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userResp := response.UserResponse{
		ID:       usrPrf.ID,
		Username: usrPrf.Username,
		Email:    usrPrf.Email,
		RoleName: usrPrf.RoleName,
		CreateAt: usrPrf.CreateAt,
		UpdateAt: usrPrf.UpdateAt,
	}
	logger.Info("User profile retrieved successfully", applog.Uint("user_id", usrPrf.ID))
	ctx.JSON(http.StatusOK, userResp)
}

func (h *UserHandler) GetUserProfile(ctx *gin.Context) {
	logger := h.logger.With(applog.String("handler", "UserHandler.GetUserProfile"))
	userID, exists := ctx.Get(middleware.UserIDKey)
	if !exists {
		logger.Error("userID not found in context, auth middleware might not have run or failed")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User ID not found in context"})
		return
	}

	id, ok := userID.(uint)
	if !ok {
		logger.Error("UserID in context is not of type uint", applog.Any("userID_type", userID))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: User ID format issue"})
		return
	}

	userProfile, err := h.userService.GetUserByID(ctx, id)
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
		ID:       userProfile.ID,
		Username: userProfile.Username,
		Email:    userProfile.Email,
		RoleName: userProfile.RoleName,
		CreateAt: userProfile.CreateAt,
		UpdateAt: userProfile.UpdateAt,
	}
	logger.Info("user profile retrieved successfully", applog.Uint("user_id", userProfile.ID))
	ctx.JSON(http.StatusOK, resp)
}
