package models

import "time"

const (
	OutboxStatusPending = "pending"
	OutboxStatusSent    = "sent"
	OutboxStatusFailed  = "failed"
)

// OutboxMessage Transactional Outbox 表
type OutboxMessage struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	Topic       string    `gorm:"type:text;not null;index"`
	Payload     string    `gorm:"type:jsonb;not null"`
	Status      string    `gorm:"type:text;default:'pending';index"`
	CreatedAt   time.Time `gorm:"type:timestamptz;default:now()"`
	ProcessedAt *time.Time `gorm:"type:timestamptz"`
}

func (OutboxMessage) TableName() string {
	return "outbox_messages"
}
