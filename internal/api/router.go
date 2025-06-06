// internal/api/router.go
package api

import (
	"mrs/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置并返回 Gin 引擎。
// healthHandler 应作为参数传入，或者在此函数内部创建。
func SetupRouter(
	healthHandler *handlers.HealthHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	movieHandler *handlers.MovieHandler,
	authMiddleware gin.HandlerFunc,
	// ... 其他处理器 ...
) *gin.Engine {
	router := gin.Default() // 或者 gin.New() 并添加必要的中间件

	// 健康检查路由
	router.GET("/health", healthHandler.CheckHealth)

	// --- 公共路由 ---
	publicRoutes := router.Group("/v1/auth")
	{
		publicRoutes.POST("/login", authHandler.Login)
	}
	// 用户注册也是公开的
	router.POST("/v1/users/register", userHandler.RegisterUser)

	// 认证中间件，所有需要认证的接口都需要通过此中间件
	protectedAPIRoutes := router.Group("/api/v1")
	protectedAPIRoutes.Use(authMiddleware)
	{
		usersRoutes := protectedAPIRoutes.Group("/users")
		{
			usersRoutes.GET("me", userHandler.GetUserProfile) // 获取当前登录用户信息
		}
		movieRoutes := protectedAPIRoutes.Group("/movies")
		{
			movieRoutes.POST("", movieHandler.CreateMovie)
			movieRoutes.GET("/:id", movieHandler.GetMovie)
			movieRoutes.PUT("/:id", movieHandler.UpdateMovie)
			movieRoutes.DELETE("/:id", movieHandler.DeleteMovie)
			movieRoutes.GET("", movieHandler.ListMovies)
		}
	}

	// ... 其他 API 路由分组和定义 ...
	// apiV1 := router.Group("/api/v1")
	// {
	// // 在这里定义 v1 版本的 API
	// }

	return router
}
