package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/handlers"
	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/captcha"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
	_ = jwt.SetDefaultSecret("test-secret-for-handler-tests-32bytes!")
	// 初始化内存验证码 store，供 Login/Register handler 使用
	captcha.InitStore(captcha.StoreConfig{})
}

// --- mock UserService ---

type mockUserService struct {
	registerFn      func(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	loginFn         func(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	refreshFn       func(ctx context.Context, token string) (*dto.AuthResponse, error)
	logoutFn        func(ctx context.Context, token string) error
	requestResetFn  func(ctx context.Context, email string) error
	resetPasswordFn func(ctx context.Context, token, newPassword string) error
	getProfileFn    func(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error)
	updateProfileFn func(ctx context.Context, id uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
	listUsersFn     func(ctx context.Context, page, size int) (*dto.UserListResponse, error)
	setStatusFn     func(ctx context.Context, id uuid.UUID, status string) error
	setRoleFn       func(ctx context.Context, id uuid.UUID, role string) error
}

func (m *mockUserService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	return m.registerFn(ctx, req)
}
func (m *mockUserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	return m.loginFn(ctx, req)
}
func (m *mockUserService) RefreshToken(ctx context.Context, token string) (*dto.AuthResponse, error) {
	return m.refreshFn(ctx, token)
}
func (m *mockUserService) Logout(ctx context.Context, token string) error {
	return m.logoutFn(ctx, token)
}
func (m *mockUserService) RequestPasswordReset(ctx context.Context, email string) error {
	return m.requestResetFn(ctx, email)
}
func (m *mockUserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	return m.resetPasswordFn(ctx, token, newPassword)
}
func (m *mockUserService) GetProfile(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	return m.getProfileFn(ctx, id)
}
func (m *mockUserService) UpdateProfile(ctx context.Context, id uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	return m.updateProfileFn(ctx, id, req)
}
func (m *mockUserService) ListUsers(ctx context.Context, page, size int) (*dto.UserListResponse, error) {
	return m.listUsersFn(ctx, page, size)
}
func (m *mockUserService) SetUserStatus(ctx context.Context, id uuid.UUID, status string) error {
	return m.setStatusFn(ctx, id, status)
}
func (m *mockUserService) SetUserRole(ctx context.Context, id uuid.UUID, role string) error {
	return m.setRoleFn(ctx, id, role)
}

// injectClaims 模拟 JWTAuth 中间件，把 claims 注入 context
func injectClaims(userID uuid.UUID, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("claims", &jwt.Claims{
			UserID: userID,
			Role:   role,
		})
		c.Next()
	}
}

func postJSON(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func getReq(r *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r.ServeHTTP(w, req)
	return w
}

func putJSON(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func putReq(r *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, path, nil)
	r.ServeHTTP(w, req)
	return w
}

func deleteReq(r *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	r.ServeHTTP(w, req)
	return w
}

// --- Captcha ---

func TestGetCaptcha_Success_Returns200(t *testing.T) {
	svc := &mockUserService{}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.GET("/captcha", h.GetCaptcha)

	w := getReq(r, "/captcha")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	assert.NotEmpty(t, data["captcha_id"])
	assert.NotEmpty(t, data["captcha_image"])
}

// --- Register ---

func TestRegister_InvalidCaptcha_Returns400(t *testing.T) {
	svc := &mockUserService{}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.POST("/register", h.Register)

	w := postJSON(r, "/register", map[string]string{
		"username":     "alice",
		"email":        "alice@example.com",
		"password":     "password123",
		"captcha_id":   "missing",
		"captcha_code": "wrong",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_Success_Returns200(t *testing.T) {
	_ = captcha.GetStore().Set("cid-register", "ans-register")
	svc := &mockUserService{
		registerFn: func(_ context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
			assert.Equal(t, "alice", req.Username)
			return &dto.AuthResponse{
				AccessToken:  "access",
				RefreshToken: "refresh",
				User:         dto.UserResponse{Username: "alice"},
			}, nil
		},
	}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.POST("/register", h.Register)

	w := postJSON(r, "/register", map[string]string{
		"username":     "alice",
		"email":        "alice@example.com",
		"password":     "password123",
		"captcha_id":   "cid-register",
		"captcha_code": "ans-register",
	})

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- Login ---

func TestLogin_MissingBody_Returns400(t *testing.T) {
	svc := &mockUserService{}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.POST("/login", h.Login)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_ServiceError_Returns401(t *testing.T) {
	_ = captcha.GetStore().Set("cid1", "ans1")
	svc := &mockUserService{
		loginFn: func(_ context.Context, _ *dto.LoginRequest) (*dto.AuthResponse, error) {
			return nil, apperr.ErrInvalidPassword
		},
	}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.POST("/login", h.Login)

	w := postJSON(r, "/login", map[string]string{
		"username":     "alice",
		"password":     "wrong",
		"captcha_id":   "cid1",
		"captcha_code": "ans1",
	})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_Success_Returns200(t *testing.T) {
	_ = captcha.GetStore().Set("cid2", "ans2")
	svc := &mockUserService{
		loginFn: func(_ context.Context, _ *dto.LoginRequest) (*dto.AuthResponse, error) {
			return &dto.AuthResponse{
				AccessToken:  "access",
				RefreshToken: "refresh",
				User:         dto.UserResponse{Username: "alice"},
			}, nil
		},
	}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.POST("/login", h.Login)

	w := postJSON(r, "/login", map[string]string{
		"username":     "alice",
		"password":     "pass",
		"captcha_id":   "cid2",
		"captcha_code": "ans2",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	assert.Equal(t, "access", data["access_token"])
}

// --- RefreshToken ---

func TestRefreshToken_MissingBody_Returns400(t *testing.T) {
	svc := &mockUserService{}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.POST("/refresh", h.RefreshToken)

	w := postJSON(r, "/refresh", map[string]string{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefreshToken_InvalidToken_Returns401(t *testing.T) {
	svc := &mockUserService{
		refreshFn: func(_ context.Context, _ string) (*dto.AuthResponse, error) {
			return nil, apperr.ErrRefreshToken
		},
	}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.POST("/refresh", h.RefreshToken)

	w := postJSON(r, "/refresh", map[string]string{"refresh_token": "bad"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- GetProfile ---

func TestGetProfile_Success_Returns200(t *testing.T) {
	userID := uuid.New()
	svc := &mockUserService{
		getProfileFn: func(_ context.Context, id uuid.UUID) (*dto.UserResponse, error) {
			assert.Equal(t, userID, id)
			return &dto.UserResponse{ID: id, Username: "alice", Role: "user", Status: "active"}, nil
		},
	}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.GET("/profile", injectClaims(userID, "user"), h.GetProfile)

	w := getReq(r, "/profile")
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	assert.Equal(t, "alice", data["username"])
}

func TestGetProfile_NotFound_Returns404(t *testing.T) {
	userID := uuid.New()
	svc := &mockUserService{
		getProfileFn: func(_ context.Context, _ uuid.UUID) (*dto.UserResponse, error) {
			return nil, apperr.ErrUserNotFound
		},
	}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.GET("/profile", injectClaims(userID, "user"), h.GetProfile)

	w := getReq(r, "/profile")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- UpdateProfile ---

func TestUpdateProfile_Success_Returns200(t *testing.T) {
	userID := uuid.New()
	svc := &mockUserService{
		updateProfileFn: func(_ context.Context, id uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
			assert.Equal(t, userID, id)
			assert.NotNil(t, req.FullName)
			return &dto.UserResponse{ID: id, Username: "alice", Role: "user", Status: "active"}, nil
		},
	}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.PUT("/profile", injectClaims(userID, "user"), h.UpdateProfile)

	w := putJSON(r, "/profile", map[string]string{"full_name": "Alice"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateProfile_BadRequest_Returns400(t *testing.T) {
	userID := uuid.New()
	svc := &mockUserService{}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.PUT("/profile", injectClaims(userID, "user"), h.UpdateProfile)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/profile", bytes.NewReader([]byte(`{"full_name":`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Logout ---

func TestLogout_Success_Returns200(t *testing.T) {
	svc := &mockUserService{
		logoutFn: func(_ context.Context, token string) error {
			assert.Equal(t, "mytoken", token)
			return nil
		},
	}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.POST("/logout", h.Logout)

	w := postJSON(r, "/logout", map[string]string{"refresh_token": "mytoken"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogout_ServiceError_Returns500(t *testing.T) {
	svc := &mockUserService{
		logoutFn: func(_ context.Context, _ string) error {
			return errors.New("store error")
		},
	}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.POST("/logout", h.Logout)

	w := postJSON(r, "/logout", map[string]string{"refresh_token": "tok"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRequestPasswordReset_Success_Returns200(t *testing.T) {
	svc := &mockUserService{
		requestResetFn: func(_ context.Context, email string) error {
			assert.Equal(t, "alice@example.com", email)
			return nil
		},
	}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.POST("/forgot-password", h.RequestPasswordReset)

	w := postJSON(r, "/forgot-password", map[string]string{"email": "alice@example.com"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResetPassword_Success_Returns200(t *testing.T) {
	svc := &mockUserService{
		resetPasswordFn: func(_ context.Context, token, newPassword string) error {
			assert.Equal(t, "reset-token", token)
			assert.Equal(t, "new-password", newPassword)
			return nil
		},
	}
	h := handlers.NewUserHandler(svc)
	r := gin.New()
	r.POST("/reset-password", h.ResetPassword)

	w := postJSON(r, "/reset-password", map[string]string{
		"token":        "reset-token",
		"new_password": "new-password",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}
