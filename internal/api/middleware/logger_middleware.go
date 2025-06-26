package middleware

import (
	applog "mrs/pkg/log"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware(logger applog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		latency := time.Since(start)
		logger.Info(path,
			applog.Int("status", c.Writer.Status()),
			applog.String("method", c.Request.Method),
			applog.String("path", path),
			applog.String("query", query),
			applog.String("ip", c.ClientIP()),
			applog.String("user-agent", c.Request.UserAgent()),
			applog.Duration("latency", latency),
		)
	}
}
