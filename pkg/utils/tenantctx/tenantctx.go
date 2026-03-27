package tenantctx

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const tenantIDKey contextKey = "tenant_id"

func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	if tenantID == uuid.Nil {
		return ctx
	}
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

func FromContext(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}
	value := ctx.Value(tenantIDKey)
	tenantID, ok := value.(uuid.UUID)
	if !ok || tenantID == uuid.Nil {
		return uuid.Nil, false
	}
	return tenantID, true
}
