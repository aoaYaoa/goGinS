package middleware

import (
	"net/http"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/logger"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

// Recovery 捕获 panic，返回 500
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("panic recovered: %v", err)
				response.FailWithCode(c, http.StatusInternalServerError, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
