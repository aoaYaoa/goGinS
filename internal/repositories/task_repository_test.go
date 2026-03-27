package repositories_test

import (
	"context"
	"testing"

	"github.com/aoaYaoa/go-gin-starter/internal/models"
	"github.com/aoaYaoa/go-gin-starter/internal/repositories"
	"github.com/aoaYaoa/go-gin-starter/internal/testhelper"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tenantctx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskRepository_CRUD(t *testing.T) {
	db, cleanup := testhelper.SetupPostgres(t)
	defer cleanup()
	require.NoError(t, db.AutoMigrate(&models.Task{}))

	repo := repositories.NewTaskRepository(db)
	ctx := context.Background()
	userID := uuid.New()

	task := &models.Task{
		ID:     uuid.New(),
		UserID: userID,
		Title:  "测试任务",
		Status: models.TaskStatusPending,
	}

	// Create
	created, err := repo.Create(ctx, task)
	require.NoError(t, err)
	assert.Equal(t, task.Title, created.Title)

	// FindByID
	found, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)

	// ListByUserID
	tasks, total, err := repo.ListByUserID(ctx, userID, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, tasks, 1)

	// Update
	created.Title = "已更新"
	created.Status = models.TaskStatusCompleted
	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	assert.Equal(t, "已更新", updated.Title)
	assert.Equal(t, models.TaskStatusCompleted, updated.Status)

	// ListAll
	all, total2, err := repo.ListAll(ctx, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total2)
	assert.Len(t, all, 1)

	// Delete
	err = repo.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, created.ID)
	assert.Error(t, err)
}

func TestTaskRepository_ListByUserID_Isolation(t *testing.T) {
	db, cleanup := testhelper.SetupPostgres(t)
	defer cleanup()
	require.NoError(t, db.AutoMigrate(&models.Task{}))

	repo := repositories.NewTaskRepository(db)
	ctx := context.Background()

	userA := uuid.New()
	userB := uuid.New()

	for i := 0; i < 3; i++ {
		_, err := repo.Create(ctx, &models.Task{
			ID: uuid.New(), UserID: userA, Title: "A的任务", Status: 0,
		})
		require.NoError(t, err)
	}
	_, err := repo.Create(ctx, &models.Task{
		ID: uuid.New(), UserID: userB, Title: "B的任务", Status: 0,
	})
	require.NoError(t, err)

	tasks, total, err := repo.ListByUserID(ctx, userA, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, tasks, 3)

	tasksB, totalB, err := repo.ListByUserID(ctx, userB, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalB)
	assert.Len(t, tasksB, 1)
}

func TestTaskRepository_TenantIsolation(t *testing.T) {
	db, cleanup := testhelper.SetupPostgres(t)
	defer cleanup()

	require.NoError(t, db.AutoMigrate(&models.Task{}))

	repo := repositories.NewTaskRepository(db)
	tenantA := uuid.New()
	tenantB := uuid.New()
	userID := uuid.New()

	ctxA := tenantctx.WithTenantID(context.Background(), tenantA)
	ctxB := tenantctx.WithTenantID(context.Background(), tenantB)

	taskA, err := repo.Create(ctxA, &models.Task{
		ID: uuid.New(), UserID: userID, Title: "tenant-a", Status: models.TaskStatusPending,
	})
	require.NoError(t, err)
	require.Equal(t, tenantA, taskA.TenantID)

	taskB, err := repo.Create(ctxB, &models.Task{
		ID: uuid.New(), UserID: userID, Title: "tenant-b", Status: models.TaskStatusPending,
	})
	require.NoError(t, err)
	require.Equal(t, tenantB, taskB.TenantID)

	tasksA, totalA, err := repo.ListAll(ctxA, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalA)
	assert.Len(t, tasksA, 1)
	assert.Equal(t, tenantA, tasksA[0].TenantID)

	_, err = repo.FindByID(ctxA, taskB.ID)
	assert.Error(t, err)
}
