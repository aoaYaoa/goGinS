package repositories

import (
	"context"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tenantctx"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func applyTenantScope(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tenantID, ok := tenantctx.FromContext(ctx); ok {
		return db.Where("tenant_id = ?", tenantID)
	}
	return db
}

func tenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	return tenantctx.FromContext(ctx)
}
