package middleware

import (
	"time"

	"linke/internal/logger"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		if statusCode >= 500 {
			logger.Error("HTTP request completed",
				logger.String("method", method),
				logger.String("path", path),
				logger.String("client_ip", clientIP),
				logger.Int("status_code", statusCode),
				logger.Duration("latency", latency),
				logger.String("user_agent", userAgent),
			)
		} else if statusCode >= 400 {
			logger.Warn("HTTP request completed",
				logger.String("method", method),
				logger.String("path", path),
				logger.String("client_ip", clientIP),
				logger.Int("status_code", statusCode),
				logger.Duration("latency", latency),
				logger.String("user_agent", userAgent),
			)
		} else {
			logger.Info("HTTP request completed",
				logger.String("method", method),
				logger.String("path", path),
				logger.String("client_ip", clientIP),
				logger.Int("status_code", statusCode),
				logger.Duration("latency", latency),
				logger.String("user_agent", userAgent),
			)
		}
	}
}