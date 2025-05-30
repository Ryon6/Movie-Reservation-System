package middleware

import (
	"mrs/internal/utils"
	applog "mrs/pkg/log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey = "Authorization"
	BearerSchema           = "Bearer "
	UserIDKey              = "userID"
	UserRoleNameKey        = "userRoleName"
	// UsernameKey            = "username" // 如果也需要用户名
)

func AuthMiddleware(jwtManager utils.JWTManager, logger applog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger = logger.With(applog.String("middleware", "AuthMiddleware"))

		// 获取 Authorization 头部
		authHeader := ctx.GetHeader(AuthorizationHeaderKey)
		if authHeader == "" {
			logger.Warn("authorization header is missing")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// 校验 Authorization 格式
		if !strings.HasPrefix(authHeader, BearerSchema) {
			logger.Warn("authorization header format is invalid", applog.String("header", authHeader))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}

		// 提取Token
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, BearerSchema))
		if tokenString == "" {
			logger.Warn("token is empty after trimming Bearer prefix", applog.String("header", authHeader))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is missing"})
			return
		}

		// 校验Token
		claims, err := jwtManager.VerifyToken(tokenString)
		if err != nil {
			logger.Warn("failed to verify token", applog.Error(err))
			// 可以根据 err 类型返回更具体的错误，例如 token 过期
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		// 将用户信息存入 Gin 上下文
		ctx.Set(UserIDKey, claims.UserID)
		ctx.Set(UserRoleNameKey, claims.RoleName)

		logger.Info("user authenticated successfully", applog.Uint("userID", claims.UserID), applog.String("role", claims.RoleName))
		ctx.Next()
	}
}
