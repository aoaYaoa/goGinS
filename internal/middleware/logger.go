package middleware

import (
	"time"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/logger"
	"github.com/gin-gonic/gin"
)

// Logger 记录每个 HTTP 请求的基本信息
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		rid, _ := c.Get("request_id")

		if query != "" {
			path = path + "?" + query
		}

		logger.Infof("%s | %d | %v | %s %s | rid=%v",
			c.ClientIP(), status, latency, method, path, rid)
	}
}
