package apperr

import "net/http"

// AppError 统一应用错误
type AppError struct {
	Code    int    // HTTP 状态码
	Message string // 对外展示的错误信息
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// 常用预定义错误
var (
	ErrUnauthorized     = New(http.StatusUnauthorized, "未授权")
	ErrForbidden        = New(http.StatusForbidden, "权限不足")
	ErrNotFound         = New(http.StatusNotFound, "资源不存在")
	ErrBadRequest       = New(http.StatusBadRequest, "请求参数错误")
	ErrInternalServer   = New(http.StatusInternalServerError, "服务器内部错误")
	ErrInvalidCaptcha   = New(http.StatusBadRequest, "验证码错误")
	ErrInvalidToken     = New(http.StatusUnauthorized, "无效的认证令牌")
	ErrTokenExpired     = New(http.StatusUnauthorized, "认证令牌已过期")
	ErrUserNotFound     = New(http.StatusNotFound, "用户不存在")
	ErrUserExists       = New(http.StatusConflict, "用户名或邮箱已存在")
	ErrInvalidPassword  = New(http.StatusUnauthorized, "用户名或密码错误")
	ErrRefreshToken     = New(http.StatusUnauthorized, "无效或已过期的刷新令牌")
	ErrAccountDisabled  = New(http.StatusForbidden, "账号已被禁用")
)
