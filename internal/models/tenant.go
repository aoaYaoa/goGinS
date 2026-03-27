package models

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"type:text;not null"`
	Slug      string    `json:"slug" gorm:"type:text;uniqueIndex;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamptz;default:now()"`
}

func (Tenant) TableName() string {
	return "tenants"
}
