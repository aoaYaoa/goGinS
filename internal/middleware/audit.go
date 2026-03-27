package middleware

import (
	"context"

	"github.com/aoaYaoa/go-gin-starter/internal/models"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/logger"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tenantctx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuditWriter interface {
	Create(ctx context.Context, entry *models.AuditLog) error
}

var defaultAuditWriter AuditWriter

func SetAuditWriter(writer AuditWriter) {
	defaultAuditWriter = writer
}

func Audit(action, resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if defaultAuditWriter == nil {
			return
		}

		entry := &models.AuditLog{
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   c.Param("id"),
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
		}

		if tenantID, ok := tenantctx.FromContext(c.Request.Context()); ok {
			entry.TenantID = tenantID
		}

		if value, exists := c.Get("claims"); exists {
			if claims, ok := value.(*jwt.Claims); ok && claims.UserID != uuid.Nil {
				entry.UserID = &claims.UserID
			}
		}

		ctx := c.Request.Context()
		go func() {
			if err := defaultAuditWriter.Create(ctx, entry); err != nil {
				logger.Warnf("[audit] 写入失败: %v", err)
			}
		}()
	}
}
