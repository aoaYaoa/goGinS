package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/handlers"
	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockTaskService struct {
	createFn      func(ctx context.Context, userID uuid.UUID, req *dto.CreateTaskRequest) (*dto.TaskResponse, error)
	getByIDFn     func(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) (*dto.TaskResponse, error)
	listFn        func(ctx context.Context, userID uuid.UUID, page, size int) (*dto.TaskListResponse, error)
	updateFn      func(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error)
	deleteFn      func(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error
	listAllFn     func(ctx context.Context, page, size int) (*dto.TaskListResponse, error)
	adminDeleteFn func(ctx context.Context, taskID uuid.UUID) error
}

func (m *mockTaskService) Create(ctx context.Context, userID uuid.UUID, req *dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	return m.createFn(ctx, userID, req)
}

func (m *mockTaskService) GetByID(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) (*dto.TaskResponse, error) {
	return m.getByIDFn(ctx, userID, taskID)
}

func (m *mockTaskService) List(ctx context.Context, userID uuid.UUID, page, size int) (*dto.TaskListResponse, error) {
	return m.listFn(ctx, userID, page, size)
}

func (m *mockTaskService) Update(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	return m.updateFn(ctx, userID, taskID, req)
}

func (m *mockTaskService) Delete(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	return m.deleteFn(ctx, userID, taskID)
}

func (m *mockTaskService) ListAll(ctx context.Context, page, size int) (*dto.TaskListResponse, error) {
	return m.listAllFn(ctx, page, size)
}

func (m *mockTaskService) AdminDelete(ctx context.Context, taskID uuid.UUID) error {
	return m.adminDeleteFn(ctx, taskID)
}

func TestTaskHandlerCreate_Success(t *testing.T) {
	userID := uuid.New()
	svc := &mockTaskService{
		createFn: func(_ context.Context, gotUserID uuid.UUID, req *dto.CreateTaskRequest) (*dto.TaskResponse, error) {
			assert.Equal(t, userID, gotUserID)
			assert.Equal(t, "task", req.Title)
			return &dto.TaskResponse{ID: uuid.New(), UserID: userID, Title: req.Title}, nil
		},
	}
	h := handlers.NewTaskHandler(svc)
	r := gin.New()
	r.POST("/tasks", injectClaims(userID, "user"), h.Create)

	w := postJSON(r, "/tasks", map[string]string{"title": "task", "description": "desc"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandlerCreate_BadRequest(t *testing.T) {
	userID := uuid.New()
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)
	r := gin.New()
	r.POST("/tasks", injectClaims(userID, "user"), h.Create)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandlerGetByID_Success(t *testing.T) {
	userID := uuid.New()
	taskID := uuid.New()
	svc := &mockTaskService{
		getByIDFn: func(_ context.Context, gotUserID uuid.UUID, gotTaskID uuid.UUID) (*dto.TaskResponse, error) {
			assert.Equal(t, userID, gotUserID)
			assert.Equal(t, taskID, gotTaskID)
			return &dto.TaskResponse{ID: taskID, UserID: userID, Title: "task"}, nil
		},
	}
	h := handlers.NewTaskHandler(svc)
	r := gin.New()
	r.GET("/tasks/:id", injectClaims(userID, "user"), h.GetByID)

	w := getReq(r, "/tasks/"+taskID.String())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandlerGetByID_InvalidID(t *testing.T) {
	userID := uuid.New()
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)
	r := gin.New()
	r.GET("/tasks/:id", injectClaims(userID, "user"), h.GetByID)

	w := getReq(r, "/tasks/not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandlerList_Success(t *testing.T) {
	userID := uuid.New()
	svc := &mockTaskService{
		listFn: func(_ context.Context, gotUserID uuid.UUID, page, size int) (*dto.TaskListResponse, error) {
			assert.Equal(t, userID, gotUserID)
			assert.Equal(t, 1, page)
			assert.Equal(t, 20, size)
			return &dto.TaskListResponse{
				Total: 1,
				Page:  page,
				Size:  size,
				Items: []dto.TaskResponse{{ID: uuid.New(), UserID: userID, Title: "task", CreatedAt: time.Now(), UpdatedAt: time.Now()}},
			}, nil
		},
	}
	h := handlers.NewTaskHandler(svc)
	r := gin.New()
	r.GET("/tasks", injectClaims(userID, "user"), h.List)

	w := getReq(r, "/tasks")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandlerUpdate_Success(t *testing.T) {
	userID := uuid.New()
	taskID := uuid.New()
	svc := &mockTaskService{
		updateFn: func(_ context.Context, gotUserID uuid.UUID, gotTaskID uuid.UUID, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
			assert.Equal(t, userID, gotUserID)
			assert.Equal(t, taskID, gotTaskID)
			assert.NotNil(t, req.Title)
			return &dto.TaskResponse{ID: taskID, UserID: userID, Title: *req.Title}, nil
		},
	}
	h := handlers.NewTaskHandler(svc)
	r := gin.New()
	r.PUT("/tasks/:id", injectClaims(userID, "user"), h.Update)

	w := putJSON(r, "/tasks/"+taskID.String(), map[string]string{"title": "updated"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandlerDelete_ServiceError(t *testing.T) {
	userID := uuid.New()
	taskID := uuid.New()
	svc := &mockTaskService{
		deleteFn: func(_ context.Context, gotUserID uuid.UUID, gotTaskID uuid.UUID) error {
			assert.Equal(t, userID, gotUserID)
			assert.Equal(t, taskID, gotTaskID)
			return apperr.ErrForbidden
		},
	}
	h := handlers.NewTaskHandler(svc)
	r := gin.New()
	r.DELETE("/tasks/:id", injectClaims(userID, "user"), h.Delete)

	w := deleteReq(r, "/tasks/"+taskID.String())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTaskHandlerDelete_InvalidID(t *testing.T) {
	userID := uuid.New()
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)
	r := gin.New()
	r.DELETE("/tasks/:id", injectClaims(userID, "user"), h.Delete)

	w := deleteReq(r, "/tasks/not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandlerList_BadQuery(t *testing.T) {
	userID := uuid.New()
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)
	r := gin.New()
	r.GET("/tasks", injectClaims(userID, "user"), h.List)

	w := getReq(r, "/tasks?size=101")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandlerUpdate_InvalidBody(t *testing.T) {
	userID := uuid.New()
	taskID := uuid.New()
	svc := &mockTaskService{}
	h := handlers.NewTaskHandler(svc)
	r := gin.New()
	r.PUT("/tasks/:id", injectClaims(userID, "user"), h.Update)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID.String(), bytes.NewReader([]byte(`{"title":`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandlerCreate_ResponseBodyContainsData(t *testing.T) {
	userID := uuid.New()
	taskID := uuid.New()
	svc := &mockTaskService{
		createFn: func(_ context.Context, _ uuid.UUID, req *dto.CreateTaskRequest) (*dto.TaskResponse, error) {
			return &dto.TaskResponse{ID: taskID, UserID: userID, Title: req.Title}, nil
		},
	}
	h := handlers.NewTaskHandler(svc)
	r := gin.New()
	r.POST("/tasks", injectClaims(userID, "user"), h.Create)

	w := postJSON(r, "/tasks", map[string]string{"title": "task"})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	assert.Equal(t, "task", data["title"])
}
