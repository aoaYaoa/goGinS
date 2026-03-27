package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aoaYaoa/go-gin-starter/internal/middleware"
	"github.com/aoaYaoa/go-gin-starter/internal/models"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// --- mock UserRepository ---

type mockUserRepo struct {
	users map[uuid.UUID]*models.User
}

func (m *mockUserRepo) Create(_ context.Context, u *models.User) (*models.User, error) {
	m.users[u.ID] = u
	return u, nil
}
func (m *mockUserRepo) FindByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}
func (m *mockUserRepo) FindByUsername(_ context.Context, username string) (*models.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}
func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}
func (m *mockUserRepo) Update(_ context.Context, u *models.User) (*models.User, error) {
	m.users[u.ID] = u
	return u, nil
}
func (m *mockUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}
func (m *mockUserRepo) List(_ context.Context, _, _ int) ([]*models.User, int64, error) {
	all := make([]*models.User, 0, len(m.users))
	for _, u := range m.users {
		all = append(all, u)
	}
	return all, int64(len(all)), nil
}

func makeUser(status string) *models.User {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.MinCost)
	return &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Password: string(hashed),
		Role:     "user",
		Status:   status,
	}
}

func setupJWTSecret(t *testing.T) {
	t.Helper()
	require.NoError(t, jwt.SetDefaultSecret("test-secret-that-is-long-enough-32chars"))
}

func makeToken(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	tok, err := jwt.GenerateAccessToken(userID, "testuser", "user")
	require.NoError(t, err)
	return tok
}

func authRouter(repo *mockUserRepo) *gin.Engine {
	r := gin.New()
	r.Use(middleware.JWTAuthWithStatusCheck(repo))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func TestJWTAuthWithStatusCheck_NoToken(t *testing.T) {
	setupJWTSecret(t)
	repo := &mockUserRepo{users: map[uuid.UUID]*models.User{}}
	r := authRouter(repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthWithStatusCheck_InvalidToken(t *testing.T) {
	setupJWTSecret(t)
	repo := &mockUserRepo{users: map[uuid.UUID]*models.User{}}
	r := authRouter(repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthWithStatusCheck_ActiveUser_Allowed(t *testing.T) {
	setupJWTSecret(t)
	u := makeUser("active")
	repo := &mockUserRepo{users: map[uuid.UUID]*models.User{u.ID: u}}
	r := authRouter(repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+makeToken(t, u.ID))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuthWithStatusCheck_BannedUser_Rejected(t *testing.T) {
	setupJWTSecret(t)
	u := makeUser("banned")
	repo := &mockUserRepo{users: map[uuid.UUID]*models.User{u.ID: u}}
	r := authRouter(repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+makeToken(t, u.ID))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestJWTAuthWithStatusCheck_DeletedUser_Rejected(t *testing.T) {
	setupJWTSecret(t)
	u := makeUser("active")
	// user not in repo (deleted/not found)
	repo := &mockUserRepo{users: map[uuid.UUID]*models.User{}}
	r := authRouter(repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+makeToken(t, u.ID))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_DoesNotCheckStatus(t *testing.T) {
	setupJWTSecret(t)
	// Plain JWTAuth should pass a banned user's valid token through (no DB lookup)
	u := makeUser("banned")
	r := gin.New()
	r.Use(middleware.JWTAuth())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+makeToken(t, u.ID))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Verify refresh token store selection: in-memory store works without Redis.
func TestMemoryRefreshStore_RoundTrip(t *testing.T) {
	store := jwt.NewMemoryRefreshStore()
	ctx := context.Background()
	tok := "test-refresh-token"

	require.NoError(t, store.Save(ctx, tok, time.Minute))

	ok, err := store.Exists(ctx, tok)
	require.NoError(t, err)
	assert.True(t, ok)

	require.NoError(t, store.Delete(ctx, tok))

	ok, err = store.Exists(ctx, tok)
	require.NoError(t, err)
	assert.False(t, ok)
}
