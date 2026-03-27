package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	TopicUserLogin = "user.login"
)

// LoginEvent 用户登录事件
type LoginEvent struct {
	EventID   string    `json:"event_id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	IP        string    `json:"ip"`
	OccurredAt time.Time `json:"occurred_at"`
}
