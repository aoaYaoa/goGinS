package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aoaYaoa/go-gin-starter/internal/middleware"
	"github.com/aoaYaoa/go-gin-starter/internal/models"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tenantctx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type stubAuditWriter struct {
	ch chan *models.AuditLog
}

func (s *stubAuditWriter) Create(_ context.Context, entry *models.AuditLog) error {
	s.ch <- entry
	return nil
}

func TestAudit_WritesEntryWithTenantAndUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tenantID := uuid.New()
	userID := uuid.New()
	writer := &stubAuditWriter{ch: make(chan *models.AuditLog, 1)}
	middleware.SetAuditWriter(writer)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("claims", &jwt.Claims{
			UserID:   userID,
			TenantID: tenantID,
			Role:     "admin",
		})
		c.Request = c.Request.WithContext(tenantctx.WithTenantID(c.Request.Context(), tenantID))
		c.Next()
	})
	r.POST("/tasks/:id", middleware.Audit("task.create", "task"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks/"+uuid.New().String(), nil)
	req.Header.Set("User-Agent", "audit-test")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	select {
	case entry := <-writer.ch:
		require.Equal(t, tenantID, entry.TenantID)
		require.NotNil(t, entry.UserID)
		require.Equal(t, userID, *entry.UserID)
		require.Equal(t, "task.create", entry.Action)
		require.Equal(t, "task", entry.ResourceType)
		require.Equal(t, "audit-test", entry.UserAgent)
	case <-time.After(2 * time.Second):
		t.Fatal("expected audit entry to be written asynchronously")
	}
}
