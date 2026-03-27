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

func TestUserRepository_CreateAndFind(t *testing.T) {
	db, cleanup := testhelper.SetupPostgres(t)
	defer cleanup()

	require.NoError(t, db.AutoMigrate(&models.User{}))

	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashed",
		Role:     "user",
		Status:   "active",
	}

	created, err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.Equal(t, user.Username, created.Username)

	found, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)

	byUsername, err := repo.FindByUsername(ctx, "testuser")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byUsername.ID)

	byEmail, err := repo.FindByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byEmail.ID)
}

func TestUserRepository_UpdateListAndDelete(t *testing.T) {
	db, cleanup := testhelper.SetupPostgres(t)
	defer cleanup()

	require.NoError(t, db.AutoMigrate(&models.User{}))

	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		ID:       uuid.New(),
		Username: "alice",
		Email:    "alice@example.com",
		Password: "hashed",
		Role:     "user",
		Status:   "active",
	}

	created, err := repo.Create(ctx, user)
	require.NoError(t, err)

	fullName := "Alice Doe"
	created.FullName = &fullName
	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	assert.Equal(t, &fullName, updated.FullName)

	users, total, err := repo.List(ctx, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)

	err = repo.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, created.ID)
	assert.Error(t, err)
}

func TestUserRepository_DuplicateUsername(t *testing.T) {
	db, cleanup := testhelper.SetupPostgres(t)
	defer cleanup()

	require.NoError(t, db.AutoMigrate(&models.User{}))

	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

	user1 := &models.User{
		ID:       uuid.New(),
		Username: "dup",
		Email:    "a@example.com",
		Password: "x",
		Role:     "user",
		Status:   "active",
	}
	user2 := &models.User{
		ID:       uuid.New(),
		Username: "dup",
		Email:    "b@example.com",
		Password: "x",
		Role:     "user",
		Status:   "active",
	}

	_, err := repo.Create(ctx, user1)
	require.NoError(t, err)

	_, err = repo.Create(ctx, user2)
	assert.Error(t, err)
}

func TestUserRepository_DeleteNotFound(t *testing.T) {
	db, cleanup := testhelper.SetupPostgres(t)
	defer cleanup()

	require.NoError(t, db.AutoMigrate(&models.User{}))

	repo := repositories.NewUserRepository(db)
	err := repo.Delete(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestUserRepository_TenantIsolation(t *testing.T) {
	db, cleanup := testhelper.SetupPostgres(t)
	defer cleanup()

	require.NoError(t, db.AutoMigrate(&models.User{}))

	repo := repositories.NewUserRepository(db)
	tenantA := uuid.New()
	tenantB := uuid.New()

	ctxA := tenantctx.WithTenantID(context.Background(), tenantA)
	ctxB := tenantctx.WithTenantID(context.Background(), tenantB)

	userA := &models.User{
		ID:       uuid.New(),
		Username: "alice-a",
		Email:    "alice-a@example.com",
		Password: "hashed",
		Role:     "user",
		Status:   "active",
	}
	userB := &models.User{
		ID:       uuid.New(),
		Username: "alice-b",
		Email:    "alice-b@example.com",
		Password: "hashed",
		Role:     "user",
		Status:   "active",
	}

	createdA, err := repo.Create(ctxA, userA)
	require.NoError(t, err)
	require.Equal(t, tenantA, createdA.TenantID)

	createdB, err := repo.Create(ctxB, userB)
	require.NoError(t, err)
	require.Equal(t, tenantB, createdB.TenantID)

	usersA, totalA, err := repo.List(ctxA, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalA)
	assert.Len(t, usersA, 1)
	assert.Equal(t, tenantA, usersA[0].TenantID)

	_, err = repo.FindByID(ctxA, createdB.ID)
	assert.Error(t, err)
}
