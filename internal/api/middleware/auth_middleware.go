package middleware

import (
	"mrs/internal/domain/user"
	"mrs/internal/utils"
	applog "mrs/pkg/log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Auth gin.HandlerFunc  // 认证中间件
type Admin gin.HandlerFunc // 管理员识别中间件

const (
	AuthorizationHeaderKey = "Authorization"
	BearerSchema           = "Bearer "
	UserIDKey              = "userID"
	UserRoleNameKey        = "userRoleName"
	// UsernameKey            = "username" // 如果也需要用户名
)

// AuthMiddleware 检查用户是否已认证
func AuthMiddleware(jwtManager utils.JWTManager, logger applog.Logger) Auth {
	return func(ctx *gin.Context) {
		inlogger := logger.With(applog.String("middleware", "AuthMiddleware"))

		// 获取 Authorization 头部
		authHeader := ctx.GetHeader(AuthorizationHeaderKey)
		if authHeader == "" {
			inlogger.Warn("authorization header is missing")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// 校验 Authorization 格式
		if !strings.HasPrefix(authHeader, BearerSchema) {
			inlogger.Warn("authorization header format is invalid", applog.String("header", authHeader))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}

		// 提取Token
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, BearerSchema))
		if tokenString == "" {
			inlogger.Warn("token is empty after trimming Bearer prefix", applog.String("header", authHeader))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is missing"})
			return
		}

		// 校验Token
		claims, err := jwtManager.VerifyToken(tokenString)
		if err != nil {
			inlogger.Warn("failed to verify token", applog.Error(err))
			// 可以根据 err 类型返回更具体的错误，例如 token 过期
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		// 将用户信息存入 Gin 上下文
		ctx.Set(UserIDKey, claims.UserID)
		ctx.Set(UserRoleNameKey, claims.RoleName)

		inlogger.Info("user authenticated successfully", applog.Uint("userID", claims.UserID), applog.String("role", claims.RoleName))
		ctx.Next()
	}
}

// AdminMiddleware 检查用户是否为管理员
func AdminMiddleware(jwtManager utils.JWTManager, logger applog.Logger) Admin {
	return func(ctx *gin.Context) {
		inlogger := logger.With(applog.String("middleware", "AdminMiddleware"))

		userRoleName := ctx.GetString(UserRoleNameKey)
		if userRoleName != user.AdminRoleName {
			inlogger.Warn("user is not admin", applog.String("role", userRoleName))
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User is not admin"})
			return
		}

		inlogger.Info("user is admin", applog.String("role", userRoleName))
		ctx.Next()
	}
}
