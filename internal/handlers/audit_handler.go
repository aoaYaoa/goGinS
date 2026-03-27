package handlers

import (
	"net/http"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/services"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	auditSvc services.AuditService
}

func NewAuditHandler(auditSvc services.AuditService) *AuditHandler {
	return &AuditHandler{auditSvc: auditSvc}
}

// List godoc
// @Summary 管理员查询审计日志
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} dto.AuditLogListResponse
// @Router /admin/audit-logs [get]
func (h *AuditHandler) List(c *gin.Context) {
	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.FailWithCode(c, http.StatusBadRequest, err.Error())
		return
	}
	query.Normalize()

	resp, err := h.auditSvc.List(c.Request.Context(), query.Page, query.Size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}
