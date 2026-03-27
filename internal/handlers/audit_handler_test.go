package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockAuditService struct {
	listFn func(ctx context.Context, page, size int) (*dto.AuditLogListResponse, error)
}

func (m *mockAuditService) List(ctx context.Context, page, size int) (*dto.AuditLogListResponse, error) {
	return m.listFn(ctx, page, size)
}

func TestAuditHandlerList_Success(t *testing.T) {
	svc := &mockAuditService{
		listFn: func(_ context.Context, page, size int) (*dto.AuditLogListResponse, error) {
			assert.Equal(t, 1, page)
			assert.Equal(t, 20, size)
			return &dto.AuditLogListResponse{
				Total: 1,
				Page:  page,
				Size:  size,
				Items: []dto.AuditLogResponse{{
					ID:           1,
					TenantID:     uuid.New(),
					UserID:       uuid.New(),
					Action:       "task.create",
					ResourceType: "task",
					ResourceID:   "resource-1",
					IP:           "127.0.0.1",
					UserAgent:    "test",
					CreatedAt:    time.Now(),
				}},
			}, nil
		},
	}

	h := handlers.NewAuditHandler(svc)
	r := gin.New()
	r.GET("/audit-logs", h.List)

	w := getReq(r, "/audit-logs")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "task.create")
}
