package handlers

import (
	"net/http"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/services"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	taskSvc services.TaskService
}

func NewTaskHandler(taskSvc services.TaskService) *TaskHandler {
	return &TaskHandler{taskSvc: taskSvc}
}

// Create godoc
// @Summary 创建任务
// @Tags tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body dto.CreateTaskRequest true "任务信息"
// @Success 200 {object} dto.TaskResponse
// @Router /tasks [post]
func (h *TaskHandler) Create(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := h.taskSvc.Create(c.Request.Context(), mustUserID(c), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

// GetByID godoc
// @Summary 获取任务详情
// @Tags tasks
// @Security BearerAuth
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} dto.TaskResponse
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetByID(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.FailWithCode(c, http.StatusBadRequest, "无效的任务 ID")
		return
	}
	resp, err := h.taskSvc.GetByID(c.Request.Context(), mustUserID(c), taskID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

// List godoc
// @Summary 获取当前用户任务列表
// @Tags tasks
// @Security BearerAuth
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} dto.TaskListResponse
// @Router /tasks [get]
func (h *TaskHandler) List(c *gin.Context) {
	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}
	query.Normalize()
	resp, err := h.taskSvc.List(c.Request.Context(), mustUserID(c), query.Page, query.Size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

// Update godoc
// @Summary 更新任务
// @Tags tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "任务 ID"
// @Param body body dto.UpdateTaskRequest true "更新内容"
// @Success 200 {object} dto.TaskResponse
// @Router /tasks/{id} [put]
func (h *TaskHandler) Update(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.FailWithCode(c, http.StatusBadRequest, "无效的任务 ID")
		return
	}
	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := h.taskSvc.Update(c.Request.Context(), mustUserID(c), taskID, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

// Delete godoc
// @Summary 删除任务
// @Tags tasks
// @Security BearerAuth
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} map[string]string
// @Router /tasks/{id} [delete]
func (h *TaskHandler) Delete(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.FailWithCode(c, http.StatusBadRequest, "无效的任务 ID")
		return
	}
	if err := h.taskSvc.Delete(c.Request.Context(), mustUserID(c), taskID); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "已删除"})
}
