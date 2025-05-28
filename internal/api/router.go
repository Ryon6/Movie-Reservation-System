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
	// ... 其他处理器 ...
) *gin.Engine {
	router := gin.Default() // 或者 gin.New() 并添加必要的中间件

	// 健康检查路由
	router.GET("/health", healthHandler.CheckHealth)

	// ... 其他 API 路由分组和定义 ...
	// apiV1 := router.Group("/api/v1")
	// {
	// // 在这里定义 v1 版本的 API
	// }

	return router
}
