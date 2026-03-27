package response

import (
	"net/http"

	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool   `json:"success"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Code:    http.StatusOK,
		Data:    data,
	})
}

func OKWithMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	})
}

func Fail(c *gin.Context, err error) {
	if appErr, ok := err.(*apperr.AppError); ok {
		c.JSON(appErr.Code, Response{
			Success: false,
			Code:    appErr.Code,
			Error:   appErr.Message,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Code:    http.StatusInternalServerError,
		Error:   err.Error(),
	})
}

func FailWithCode(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Success: false,
		Code:    code,
		Error:   message,
	})
}

// ValidationError 处理 Gin 验证错误，返回友好的错误信息
func ValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Code:    http.StatusBadRequest,
		Error:   "请求参数验证失败: " + err.Error(),
	})
}
