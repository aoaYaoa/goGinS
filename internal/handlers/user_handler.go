package handlers

import (
	"net/http"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/services"
	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/captcha"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mojocn/base64Captcha"
)

type UserHandler struct {
	userSvc services.UserService
}

func NewUserHandler(userSvc services.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// GetCaptcha godoc
// @Summary 获取验证码
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/captcha [get]
func (h *UserHandler) GetCaptcha(c *gin.Context) {
	driver := base64Captcha.NewDriverDigit(80, 240, 6, 0.7, 80)
	cap := base64Captcha.NewCaptcha(driver, captcha.GetStore())
	id, b64s, _, err := cap.Generate()
	if err != nil {
		response.Fail(c, apperr.ErrInternalServer)
		return
	}
	response.OK(c, gin.H{
		"captcha_id":    id,
		"captcha_image": b64s,
	})
}

// Register godoc
// @Summary 注册
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.RegisterRequest true "注册信息"
// @Success 200 {object} dto.AuthResponse
// @Router /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}

	// 校验验证码
	if !captcha.GetStore().Verify(req.CaptchaID, req.CaptchaCode, true) {
		response.Fail(c, apperr.ErrInvalidCaptcha)
		return
	}

	resp, err := h.userSvc.Register(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

// Login godoc
// @Summary 登录
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "登录信息"
// @Success 200 {object} dto.AuthResponse
// @Router /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}

	// 校验验证码
	if !captcha.GetStore().Verify(req.CaptchaID, req.CaptchaCode, true) {
		response.Fail(c, apperr.ErrInvalidCaptcha)
		return
	}

	resp, err := h.userSvc.Login(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

// RefreshToken godoc
// @Summary 刷新 Token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.RefreshRequest true "Refresh Token"
// @Success 200 {object} dto.AuthResponse
// @Router /auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.userSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

// Logout godoc
// @Summary 登出
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.LogoutRequest true "Refresh Token"
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.userSvc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "已登出"})
}

// RequestPasswordReset godoc
// @Summary 发起密码重置
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.PasswordResetRequest true "邮箱信息"
// @Success 200 {object} map[string]string
// @Router /auth/forgot-password [post]
func (h *UserHandler) RequestPasswordReset(c *gin.Context) {
	var req dto.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.userSvc.RequestPasswordReset(c.Request.Context(), req.Email); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "如果邮箱存在，将收到密码重置邮件"})
}

// ResetPassword godoc
// @Summary 重置密码
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} map[string]string
// @Router /auth/reset-password [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.userSvc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "密码已重置"})
}

// GetProfile godoc
// @Summary 获取当前用户资料
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Router /user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := mustUserID(c)
	resp, err := h.userSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

// UpdateProfile godoc
// @Summary 更新当前用户资料
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body dto.UpdateProfileRequest true "更新内容"
// @Success 200 {object} dto.UserResponse
// @Router /user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := mustUserID(c)
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := h.userSvc.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

// mustUserID 从 JWT 中间件注入的 context 读取 userID
func mustUserID(c *gin.Context) uuid.UUID {
	claims, _ := c.Get("claims")
	return claims.(*jwt.Claims).UserID
}
