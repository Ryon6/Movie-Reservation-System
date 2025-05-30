// TODO: username already exists 等错误直接定义好用于调用
// TODO: 自顶向下修改错误的使用方法

package handlers

import (
	"errors"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/app"
	"mrs/internal/domain/user"
	applog "mrs/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	logger      applog.Logger
	userService app.UserService
}

func NewUserHandler(logger applog.Logger, userService app.UserService) *UserHandler {
	return &UserHandler{
		logger:      logger,
		userService: userService,
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
