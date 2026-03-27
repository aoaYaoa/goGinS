package repositories

import (
	"context"

	"github.com/aoaYaoa/go-gin-starter/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditRepository interface {
	Create(ctx context.Context, entry *models.AuditLog) error
	List(ctx context.Context, offset, limit int) ([]*models.AuditLog, int64, error)
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, entry *models.AuditLog) error {
	if tenantID, ok := tenantIDFromContext(ctx); ok && entry.TenantID == uuid.Nil {
		entry.TenantID = tenantID
	}
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *auditRepository) List(ctx context.Context, offset, limit int) ([]*models.AuditLog, int64, error) {
	var entries []*models.AuditLog
	var total int64

	db := applyTenantScope(ctx, r.db.WithContext(ctx)).Model(&models.AuditLog{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&entries).Error; err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}
