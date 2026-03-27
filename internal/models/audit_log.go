package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID           uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID     uuid.UUID  `json:"tenant_id" gorm:"type:uuid;index"`
	UserID       *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	Action       string     `json:"action" gorm:"type:text;not null;index"`
	ResourceType string     `json:"resource_type" gorm:"type:text;not null;index"`
	ResourceID   string     `json:"resource_id" gorm:"type:text"`
	IP           string     `json:"ip" gorm:"type:text"`
	UserAgent    string     `json:"user_agent" gorm:"type:text"`
	CreatedAt    time.Time  `json:"created_at" gorm:"type:timestamptz;default:now();index"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
