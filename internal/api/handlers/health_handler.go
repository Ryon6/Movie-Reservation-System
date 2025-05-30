package handlers

import (
	"mrs/internal/api/dto/response"
	applog "mrs/pkg/log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db     *gorm.DB
	rdb    *redis.Client
	logger applog.Logger
}

func NewHealthHandler(db *gorm.DB, rdb *redis.Client, logger applog.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		rdb:    rdb,
		logger: logger,
	}
}

// CheckHealth 处理 /health API 请求。
func (h *HealthHandler) CheckHealth(c *gin.Context) {
	h.logger.Info("Health check requested")

	// 基本的健康响应
	resp := response.HealthResponse{
		OverallStatus: "UP",
		// Version: h.AppVersion, // 如果注入了版本号
		TimeStamp: time.Now().Format(time.RFC3339),
	}
	// 检查依赖状态
	var dependencies []response.ComponentStatus

	// 检查数据库
	if h.db != nil {
		dbStatus := response.ComponentStatus{Name: "database", Status: "UP"}
		sqlDB, err := h.db.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			dbStatus.Status = "DOWN"
			if err != nil {
				dbStatus.Message = err.Error()
			} else {
				dbStatus.Message = sqlDB.PingContext(c.Request.Context()).Error()
			}
			resp.OverallStatus = "DEGRADED" // 或 DOWN，根据策略
		}
		dependencies = append(dependencies, dbStatus)
	}

	// 检查 Redis
	if h.rdb != nil {
		redisStatus := response.ComponentStatus{Name: "cache", Status: "UP"}
		if err := h.rdb.Ping(c.Request.Context()).Err(); err != nil {
			redisStatus.Status = "DOWN"
			redisStatus.Message = err.Error()
			resp.OverallStatus = "DEGRADED" // 或 DOWN
		}
		dependencies = append(dependencies, redisStatus)
	}

	if len(dependencies) > 0 {
		resp.DependenciesStatus = dependencies
	}

	// 根据总体状态决定 HTTP 状态码
	httpStatus := http.StatusOK
	// if resp.OverallStatus == "DOWN" { // 极端情况下
	// httpStatus = http.StatusServiceUnavailable
	// }

	c.JSON(httpStatus, resp)
}
