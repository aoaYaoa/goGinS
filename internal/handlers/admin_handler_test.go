package handlers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/handlers"
	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAdminHandlerListUsers_Success(t *testing.T) {
	svc := &mockUserService{
		listUsersFn: func(_ context.Context, page, size int) (*dto.UserListResponse, error) {
			assert.Equal(t, 1, page)
			assert.Equal(t, 20, size)
			return &dto.UserListResponse{
				Total: 1,
				Page:  page,
				Size:  size,
				Items: []*dto.UserResponse{{ID: uuid.New(), Username: "alice", Role: "user", Status: "active"}},
			}, nil
		},
	}
	taskSvc := &mockTaskService{}
	h := handlers.NewAdminHandler(svc, taskSvc)
	r := gin.New()
	r.GET("/admin/users", h.ListUsers)

	w := getReq(r, "/admin/users")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandlerGetUser_InvalidID(t *testing.T) {
	h := handlers.NewAdminHandler(&mockUserService{}, &mockTaskService{})
	r := gin.New()
	r.GET("/admin/users/:id", h.GetUser)

	w := getReq(r, "/admin/users/not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandlerBanUser_Success(t *testing.T) {
	userID := uuid.New()
	svc := &mockUserService{
		setStatusFn: func(_ context.Context, id uuid.UUID, status string) error {
			assert.Equal(t, userID, id)
			assert.Equal(t, "banned", status)
			return nil
		},
	}
	h := handlers.NewAdminHandler(svc, &mockTaskService{})
	r := gin.New()
	r.PUT("/admin/users/:id/ban", h.BanUser)

	w := putReq(r, "/admin/users/"+userID.String()+"/ban")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandlerUnbanUser_Success(t *testing.T) {
	userID := uuid.New()
	svc := &mockUserService{
		setStatusFn: func(_ context.Context, id uuid.UUID, status string) error {
			assert.Equal(t, userID, id)
			assert.Equal(t, "active", status)
			return nil
		},
	}
	h := handlers.NewAdminHandler(svc, &mockTaskService{})
	r := gin.New()
	r.PUT("/admin/users/:id/unban", h.UnbanUser)

	w := putReq(r, "/admin/users/"+userID.String()+"/unban")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandlerListAllTasks_Success(t *testing.T) {
	taskSvc := &mockTaskService{
		listAllFn: func(_ context.Context, page, size int) (*dto.TaskListResponse, error) {
			assert.Equal(t, 1, page)
			assert.Equal(t, 20, size)
			return &dto.TaskListResponse{
				Total: 1,
				Page:  page,
				Size:  size,
				Items: []dto.TaskResponse{{ID: uuid.New(), UserID: uuid.New(), Title: "task", CreatedAt: time.Now(), UpdatedAt: time.Now()}},
			}, nil
		},
	}
	h := handlers.NewAdminHandler(&mockUserService{}, taskSvc)
	r := gin.New()
	r.GET("/admin/tasks", h.ListAllTasks)

	w := getReq(r, "/admin/tasks")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandlerDeleteTask_InvalidID(t *testing.T) {
	h := handlers.NewAdminHandler(&mockUserService{}, &mockTaskService{})
	r := gin.New()
	r.DELETE("/admin/tasks/:id", h.DeleteTask)

	w := deleteReq(r, "/admin/tasks/not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandlerPromoteUser_Success(t *testing.T) {
	userID := uuid.New()
	svc := &mockUserService{
		setRoleFn: func(_ context.Context, id uuid.UUID, role string) error {
			assert.Equal(t, userID, id)
			assert.Equal(t, "admin", role)
			return nil
		},
	}
	h := handlers.NewAdminHandler(svc, &mockTaskService{})
	r := gin.New()
	r.PUT("/admin/users/:id/promote", h.PromoteUser)

	w := putReq(r, "/admin/users/"+userID.String()+"/promote")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandlerDemoteUser_ServiceError_Returns500(t *testing.T) {
	userID := uuid.New()
	svc := &mockUserService{
		setRoleFn: func(_ context.Context, id uuid.UUID, role string) error {
			assert.Equal(t, userID, id)
			assert.Equal(t, "user", role)
			return apperr.ErrForbidden
		},
	}
	h := handlers.NewAdminHandler(svc, &mockTaskService{})
	r := gin.New()
	r.PUT("/admin/users/:id/demote", h.DemoteUser)

	w := putReq(r, "/admin/users/"+userID.String()+"/demote")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
