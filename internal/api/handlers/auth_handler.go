package handlers

import (
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/app"
	applog "mrs/pkg/log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService app.AuthService
	logger      applog.Logger
}

func NewAuthHandler(authService app.AuthService, logger applog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Login 处理用户登录请求。
func (h *AuthHandler) Login(ctx *gin.Context) {
	logger := h.logger.With(applog.String("handler", "AuthHandler.Login"))
	// 绑定登录请求
	var req request.LoginRequest
	if err := ctx.BindJSON(&req); err != nil {
		logger.Warn("Failed to bind login request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// 验证登录信息
	loginResult, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		// 可添加自定义错误
		// if errors.Is(err, 自定义错误)
		logger.Error("Login service failed", applog.Error(err), applog.String("username", req.Username))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed due to an internal error"})
		return
	}

	// 发送响应报文
	loginResp := response.LoginResponse{
		Token:     loginResult.Token,
		ExpiresAt: time.Now().Add(time.Hour),
		User:      response.ToUserResponse(loginResult),
	}
	logger.Info("User logged in successfully", applog.String("username", req.Username))
	ctx.JSON(http.StatusOK, loginResp)
}
