package services

import (
	"context"
	"errors"
	"testing"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/models"
	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
	"github.com/google/uuid"
)

// mockTaskRepository implements repositories.TaskRepository
type mockTaskRepository struct {
	tasks map[uuid.UUID]*models.Task
	err   error
}

func newMockTaskRepo() *mockTaskRepository {
	return &mockTaskRepository{tasks: make(map[uuid.UUID]*models.Task)}
}

func (m *mockTaskRepository) Create(_ context.Context, task *models.Task) (*models.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskRepository) FindByID(_ context.Context, id uuid.UUID) (*models.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	t, ok := m.tasks[id]
	if !ok {
		return nil, errors.New("任务不存在")
	}
	return t, nil
}

func (m *mockTaskRepository) ListByUserID(_ context.Context, userID uuid.UUID, offset, limit int) ([]*models.Task, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	var result []*models.Task
	for _, t := range m.tasks {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	total := int64(len(result))
	start := offset
	if start > len(result) {
		start = len(result)
	}
	end := start + limit
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

func (m *mockTaskRepository) Update(_ context.Context, task *models.Task) (*models.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskRepository) Delete(_ context.Context, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.tasks[id]; !ok {
		return errors.New("任务不存在")
	}
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepository) ListAll(_ context.Context, offset, limit int) ([]*models.Task, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	var result []*models.Task
	for _, t := range m.tasks {
		result = append(result, t)
	}
	total := int64(len(result))
	start := offset
	if start > len(result) {
		start = len(result)
	}
	end := start + limit
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

// --- Tests ---

func TestTaskCreate_Success(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	userID := uuid.New()

	resp, err := svc.Create(context.Background(), userID, &dto.CreateTaskRequest{
		Title:       "test task",
		Description: "desc",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Title != "test task" {
		t.Errorf("expected title 'test task', got %q", resp.Title)
	}
	if resp.Status != models.TaskStatusPending {
		t.Errorf("expected status pending, got %d", resp.Status)
	}
}

func TestTaskCreate_RepoError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.err = errors.New("db error")
	svc := NewTaskService(repo)

	_, err := svc.Create(context.Background(), uuid.New(), &dto.CreateTaskRequest{Title: "x"})
	if err != apperr.ErrInternalServer {
		t.Errorf("expected ErrInternalServer, got %v", err)
	}
}

func TestTaskGetByID_Success(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	userID := uuid.New()
	taskID := uuid.New()
	repo.tasks[taskID] = &models.Task{ID: taskID, UserID: userID, Title: "t1", Status: models.TaskStatusPending}

	resp, err := svc.GetByID(context.Background(), userID, taskID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != taskID {
		t.Errorf("wrong task ID")
	}
}

func TestTaskGetByID_WrongUser(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	taskID := uuid.New()
	repo.tasks[taskID] = &models.Task{ID: taskID, UserID: uuid.New(), Title: "t1"}

	_, err := svc.GetByID(context.Background(), uuid.New(), taskID)
	if err != apperr.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskGetByID_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)

	_, err := svc.GetByID(context.Background(), uuid.New(), uuid.New())
	if err != apperr.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskList_Success(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	userID := uuid.New()
	for i := 0; i < 3; i++ {
		id := uuid.New()
		repo.tasks[id] = &models.Task{ID: id, UserID: userID, Title: "task"}
	}

	resp, err := svc.List(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 3 {
		t.Errorf("expected total 3, got %d", resp.Total)
	}
}

func TestTaskUpdate_Success(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	userID := uuid.New()
	taskID := uuid.New()
	repo.tasks[taskID] = &models.Task{ID: taskID, UserID: userID, Title: "old", Status: models.TaskStatusPending}

	newTitle := "new title"
	resp, err := svc.Update(context.Background(), userID, taskID, &dto.UpdateTaskRequest{Title: &newTitle})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Title != "new title" {
		t.Errorf("expected updated title, got %q", resp.Title)
	}
}

func TestTaskUpdate_WrongUser(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	taskID := uuid.New()
	repo.tasks[taskID] = &models.Task{ID: taskID, UserID: uuid.New(), Title: "t"}

	title := "x"
	_, err := svc.Update(context.Background(), uuid.New(), taskID, &dto.UpdateTaskRequest{Title: &title})
	if err != apperr.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskDelete_Success(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	userID := uuid.New()
	taskID := uuid.New()
	repo.tasks[taskID] = &models.Task{ID: taskID, UserID: userID}

	if err := svc.Delete(context.Background(), userID, taskID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.tasks[taskID]; ok {
		t.Error("task should have been deleted")
	}
}

func TestTaskDelete_WrongUser(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	taskID := uuid.New()
	repo.tasks[taskID] = &models.Task{ID: taskID, UserID: uuid.New()}

	err := svc.Delete(context.Background(), uuid.New(), taskID)
	if err != apperr.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskAdminDelete_Success(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	taskID := uuid.New()
	repo.tasks[taskID] = &models.Task{ID: taskID, UserID: uuid.New()}

	if err := svc.AdminDelete(context.Background(), taskID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskListAll_Success(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	for i := 0; i < 5; i++ {
		id := uuid.New()
		repo.tasks[id] = &models.Task{ID: id, UserID: uuid.New(), Title: "t"}
	}

	resp, err := svc.ListAll(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 5 {
		t.Errorf("expected total 5, got %d", resp.Total)
	}
}
