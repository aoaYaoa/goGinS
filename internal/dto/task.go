package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

type TaskResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaskListResponse struct {
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
	Items []TaskResponse `json:"items"`
}

// PageQuery 通用分页查询参数
type PageQuery struct {
	Page int `form:"page" binding:"omitempty,min=1"`
	Size int `form:"size" binding:"omitempty,min=1,max=100"`
}

func (q *PageQuery) Normalize() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 {
		q.Size = 20
	}
}

func (q *PageQuery) Offset() int {
	return (q.Page - 1) * q.Size
}
