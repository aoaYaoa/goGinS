package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	TaskStatusPending   = 0
	TaskStatusCompleted = 1
)

type Task struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID `json:"tenant_id" gorm:"type:uuid;index"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Title       string    `json:"title" gorm:"type:text;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Status      int       `json:"status" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"type:timestamptz;default:now()"`
}

func (Task) TableName() string {
	return "tasks"
}
