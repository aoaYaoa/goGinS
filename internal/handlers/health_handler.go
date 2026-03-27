package handlers

import (
	"net/http"

	"github.com/aoaYaoa/go-gin-starter/internal/database"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *database.DB
}

func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Liveness godoc
// @Summary 存活检查
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readiness godoc
// @Summary 就绪检查（含 DB ping）
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /readyz [get]
func (h *HealthHandler) Readiness(c *gin.Context) {
	if err := h.db.Ping(c.Request.Context()); err != nil {
		response.FailWithCode(c, http.StatusServiceUnavailable, "数据库不可用")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
