package services

import (
	"context"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/models"
	"github.com/aoaYaoa/go-gin-starter/internal/repositories"
	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
	"github.com/google/uuid"
)

type TaskService interface {
	Create(ctx context.Context, userID uuid.UUID, req *dto.CreateTaskRequest) (*dto.TaskResponse, error)
	GetByID(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) (*dto.TaskResponse, error)
	List(ctx context.Context, userID uuid.UUID, page, size int) (*dto.TaskListResponse, error)
	Update(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error)
	Delete(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error
	// Admin
	ListAll(ctx context.Context, page, size int) (*dto.TaskListResponse, error)
	AdminDelete(ctx context.Context, taskID uuid.UUID) error
}

type taskService struct {
	taskRepo repositories.TaskRepository
}

func NewTaskService(taskRepo repositories.TaskRepository) TaskService {
	return &taskService{taskRepo: taskRepo}
}

func (s *taskService) Create(ctx context.Context, userID uuid.UUID, req *dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	task := &models.Task{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      models.TaskStatusPending,
	}
	created, err := s.taskRepo.Create(ctx, task)
	if err != nil {
		return nil, apperr.ErrInternalServer
	}
	return toTaskResponse(created), nil
}

func (s *taskService) GetByID(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if task.UserID != userID {
		return nil, apperr.ErrForbidden
	}
	return toTaskResponse(task), nil
}

func (s *taskService) List(ctx context.Context, userID uuid.UUID, page, size int) (*dto.TaskListResponse, error) {
	offset := (page - 1) * size
	tasks, total, err := s.taskRepo.ListByUserID(ctx, userID, offset, size)
	if err != nil {
		return nil, apperr.ErrInternalServer
	}
	items := make([]dto.TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, *toTaskResponse(t))
	}
	return &dto.TaskListResponse{Total: total, Page: page, Size: size, Items: items}, nil
}

func (s *taskService) Update(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if task.UserID != userID {
		return nil, apperr.ErrForbidden
	}
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	updated, err := s.taskRepo.Update(ctx, task)
	if err != nil {
		return nil, apperr.ErrInternalServer
	}
	return toTaskResponse(updated), nil
}

func (s *taskService) Delete(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return apperr.ErrNotFound
	}
	if task.UserID != userID {
		return apperr.ErrForbidden
	}
	return s.taskRepo.Delete(ctx, taskID)
}

func (s *taskService) ListAll(ctx context.Context, page, size int) (*dto.TaskListResponse, error) {
	offset := (page - 1) * size
	tasks, total, err := s.taskRepo.ListAll(ctx, offset, size)
	if err != nil {
		return nil, apperr.ErrInternalServer
	}
	items := make([]dto.TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, *toTaskResponse(t))
	}
	return &dto.TaskListResponse{Total: total, Page: page, Size: size, Items: items}, nil
}

func (s *taskService) AdminDelete(ctx context.Context, taskID uuid.UUID) error {
	_, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return apperr.ErrNotFound
	}
	return s.taskRepo.Delete(ctx, taskID)
}

func toTaskResponse(t *models.Task) *dto.TaskResponse {
	return &dto.TaskResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
