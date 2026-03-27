package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;index"`
	Username    string     `json:"username" gorm:"type:text;uniqueIndex;not null"`
	Email       string     `json:"email" gorm:"type:text;uniqueIndex"`
	Password    string     `json:"-" gorm:"type:text;not null"`
	Role        string     `json:"role" gorm:"type:text;default:'user'"`
	FullName    *string    `json:"full_name,omitempty" gorm:"type:text"`
	AvatarURL   *string    `json:"avatar_url,omitempty" gorm:"type:text"`
	Phone       *string    `json:"phone,omitempty" gorm:"type:text"`
	Status      string     `json:"status" gorm:"type:text;default:'active'"`
	IsVerified  bool       `json:"is_verified" gorm:"default:false"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt   time.Time  `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"type:timestamptz;default:now()"`
}

func (User) TableName() string {
	return "users"
}
