package handlers

import (
	"net/http"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/services"
	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	userSvc services.UserService
	taskSvc services.TaskService
}

func NewAdminHandler(userSvc services.UserService, taskSvc services.TaskService) *AdminHandler {
	return &AdminHandler{userSvc: userSvc, taskSvc: taskSvc}
}

// ListUsers godoc
// @Summary 管理员获取所有用户
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} dto.UserListResponse
// @Router /admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}
	query.Normalize()
	users, err := h.userSvc.ListUsers(c.Request.Context(), query.Page, query.Size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, users)
}

// GetUser godoc
// @Summary 管理员获取指定用户
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "用户 ID"
// @Success 200 {object} dto.UserResponse
// @Router /admin/users/{id} [get]
func (h *AdminHandler) GetUser(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.FailWithCode(c, http.StatusBadRequest, "无效的用户 ID")
		return
	}
	user, err := h.userSvc.GetProfile(c.Request.Context(), uid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, user)
}

// BanUser godoc
// @Summary 管理员封禁用户
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "用户 ID"
// @Success 200 {object} map[string]string
// @Router /admin/users/{id}/ban [put]
func (h *AdminHandler) BanUser(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.FailWithCode(c, http.StatusBadRequest, "无效的用户 ID")
		return
	}
	if err := h.userSvc.SetUserStatus(c.Request.Context(), uid, "banned"); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "用户已封禁"})
}

// UnbanUser godoc
// @Summary 管理员解封用户
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "用户 ID"
// @Success 200 {object} map[string]string
// @Router /admin/users/{id}/unban [put]
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.FailWithCode(c, http.StatusBadRequest, "无效的用户 ID")
		return
	}
	if err := h.userSvc.SetUserStatus(c.Request.Context(), uid, "active"); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "用户已解封"})
}

// ListAllTasks godoc
// @Summary 管理员获取所有任务
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} dto.TaskListResponse
// @Router /admin/tasks [get]
func (h *AdminHandler) ListAllTasks(c *gin.Context) {
	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}
	query.Normalize()
	tasks, err := h.taskSvc.ListAll(c.Request.Context(), query.Page, query.Size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, tasks)
}

// DeleteTask godoc
// @Summary 管理员强制删除任意任务
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} map[string]string
// @Router /admin/tasks/{id} [delete]
func (h *AdminHandler) DeleteTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.FailWithCode(c, http.StatusBadRequest, "无效的任务 ID")
		return
	}
	if err := h.taskSvc.AdminDelete(c.Request.Context(), taskID); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "已删除"})
}

// PromoteUser godoc
// @Summary 管理员提升用户为 admin
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "用户 ID"
// @Success 200 {object} map[string]string
// @Router /admin/users/{id}/promote [put]
func (h *AdminHandler) PromoteUser(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.FailWithCode(c, http.StatusBadRequest, "无效的用户 ID")
		return
	}
	if err := h.userSvc.SetUserRole(c.Request.Context(), uid, "admin"); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "已提升为管理员"})
}

// DemoteUser godoc
// @Summary 管理员降级用户
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "用户 ID"
// @Success 200 {object} map[string]string
// @Router /admin/users/{id}/demote [put]
func (h *AdminHandler) DemoteUser(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.FailWithCode(c, http.StatusBadRequest, "无效的用户 ID")
		return
	}
	if err := h.userSvc.SetUserRole(c.Request.Context(), uid, "user"); err != nil {
		response.Fail(c, apperr.ErrInternalServer)
		return
	}
	response.OK(c, gin.H{"message": "已降级为普通用户"})
}
