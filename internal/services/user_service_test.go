package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aoaYaoa/go-gin-starter/internal/dto"
	"github.com/aoaYaoa/go-gin-starter/internal/models"
	"github.com/aoaYaoa/go-gin-starter/pkg/apperr"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tenantctx"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// --- mock UserRepository ---

type mockUserRepo struct {
	usersByName  map[string]*models.User
	usersByEmail map[string]*models.User
	usersByID    map[uuid.UUID]*models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByName:  make(map[string]*models.User),
		usersByEmail: make(map[string]*models.User),
		usersByID:    make(map[uuid.UUID]*models.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *models.User) (*models.User, error) {
	if tenantID, ok := tenantctx.FromContext(ctx); ok {
		u.TenantID = tenantID
	}
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	m.usersByName[u.Username] = u
	if u.Email != "" {
		m.usersByEmail[u.Email] = u
	}
	m.usersByID[u.ID] = u
	return u, nil
}

func (m *mockUserRepo) FindByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return nil, errors.New("用户不存在")
	}
	return u, nil
}

func (m *mockUserRepo) FindByUsername(_ context.Context, username string) (*models.User, error) {
	u, ok := m.usersByName[username]
	if !ok {
		return nil, errors.New("用户不存在")
	}
	return u, nil
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*models.User, error) {
	u, ok := m.usersByEmail[email]
	if !ok {
		return nil, errors.New("用户不存在")
	}
	return u, nil
}

func (m *mockUserRepo) Update(_ context.Context, u *models.User) (*models.User, error) {
	m.usersByName[u.Username] = u
	m.usersByID[u.ID] = u
	if u.Email != "" {
		m.usersByEmail[u.Email] = u
	}
	return u, nil
}

func (m *mockUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	u, ok := m.usersByID[id]
	if !ok {
		return errors.New("用户不存在")
	}
	delete(m.usersByName, u.Username)
	delete(m.usersByID, id)
	return nil
}

func (m *mockUserRepo) List(_ context.Context, offset, limit int) ([]*models.User, int64, error) {
	all := make([]*models.User, 0, len(m.usersByID))
	for _, u := range m.usersByID {
		all = append(all, u)
	}
	total := int64(len(all))
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	if offset >= len(all) {
		return []*models.User{}, total, nil
	}
	return all[offset:end], total, nil
}

// --- mock RefreshTokenStore ---

type mockRefreshStore struct {
	tokens map[string]time.Time
}

func newMockRefreshStore() *mockRefreshStore {
	return &mockRefreshStore{tokens: make(map[string]time.Time)}
}

func (m *mockRefreshStore) Save(_ context.Context, token string, ttl time.Duration) error {
	m.tokens[token] = time.Now().Add(ttl)
	return nil
}

func (m *mockRefreshStore) Exists(_ context.Context, token string) (bool, error) {
	exp, ok := m.tokens[token]
	if !ok {
		return false, nil
	}
	return time.Now().Before(exp), nil
}

func (m *mockRefreshStore) Delete(_ context.Context, token string) error {
	delete(m.tokens, token)
	return nil
}

func (m *mockRefreshStore) Ping(_ context.Context) error { return nil }

type mockNotifier struct {
	sendCalls int
	lastTo    string
	lastSubj  string
	lastBody  string
}

func (m *mockNotifier) Send(_ context.Context, to, subject, body string) error {
	m.sendCalls++
	m.lastTo = to
	m.lastSubj = subject
	m.lastBody = body
	return nil
}

// --- helpers ---

func setupUserService(t *testing.T) (UserService, *mockUserRepo, *mockRefreshStore) {
	t.Helper()
	_ = jwt.SetDefaultSecret("test-secret-at-least-32-characters-long")
	repo := newMockUserRepo()
	store := newMockRefreshStore()
	svc := &userService{
		userRepo:     repo,
		refreshStore: store,
		notifier:     &mockNotifier{},
		pub:          nil,
		db:           nil,
	}
	return svc, repo, store
}

func makeActiveUser(t *testing.T, username, password string) *models.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return &models.User{
		ID:       uuid.New(),
		Username: username,
		Email:    username + "@example.com",
		Password: string(hash),
		Role:     "user",
		Status:   "active",
	}
}

// --- Register tests ---

func TestRegister_Success(t *testing.T) {
	svc, _, _ := setupUserService(t)
	resp, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("期望注册成功，got error: %v", err)
	}
	if resp.User.Username != "alice" {
		t.Errorf("期望 username=alice，got %s", resp.User.Username)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Error("期望返回 token，got empty")
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	svc, repo, _ := setupUserService(t)
	existing := makeActiveUser(t, "alice", "pass")
	repo.usersByName["alice"] = existing
	repo.usersByID[existing.ID] = existing

	_, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Username: "alice",
		Password: "newpass123",
	})
	if !errors.Is(err, apperr.ErrUserExists) {
		t.Errorf("期望 ErrUserExists，got %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, repo, _ := setupUserService(t)
	existing := makeActiveUser(t, "bob", "pass")
	repo.usersByName["bob"] = existing
	repo.usersByEmail["shared@example.com"] = existing
	repo.usersByID[existing.ID] = existing

	_, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Username: "carol",
		Email:    "shared@example.com",
		Password: "pass123",
	})
	if !errors.Is(err, apperr.ErrUserExists) {
		t.Errorf("期望 ErrUserExists，got %v", err)
	}
}

// --- Login tests ---

func TestLogin_Success(t *testing.T) {
	svc, repo, _ := setupUserService(t)
	u := makeActiveUser(t, "alice", "correct")
	repo.usersByName["alice"] = u
	repo.usersByID[u.ID] = u

	resp, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "alice",
		Password: "correct",
	})
	if err != nil {
		t.Fatalf("期望登录成功，got error: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("期望返回 access token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, repo, _ := setupUserService(t)
	u := makeActiveUser(t, "alice", "correct")
	repo.usersByName["alice"] = u
	repo.usersByID[u.ID] = u

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "alice",
		Password: "wrong",
	})
	if !errors.Is(err, apperr.ErrInvalidPassword) {
		t.Errorf("期望 ErrInvalidPassword，got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, _, _ := setupUserService(t)
	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "nobody",
		Password: "pass",
	})
	if !errors.Is(err, apperr.ErrInvalidPassword) {
		t.Errorf("期望 ErrInvalidPassword，got %v", err)
	}
}

func TestLogin_DisabledAccount(t *testing.T) {
	svc, repo, _ := setupUserService(t)
	u := makeActiveUser(t, "alice", "correct")
	u.Status = "banned"
	repo.usersByName["alice"] = u
	repo.usersByID[u.ID] = u

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "alice",
		Password: "correct",
	})
	if !errors.Is(err, apperr.ErrAccountDisabled) {
		t.Errorf("期望 ErrAccountDisabled，got %v", err)
	}
}

func TestLogin_TokenContainsTenantID(t *testing.T) {
	svc, repo, _ := setupUserService(t)
	u := makeActiveUser(t, "alice", "correct")
	u.TenantID = uuid.New()
	repo.usersByName["alice"] = u
	repo.usersByID[u.ID] = u

	resp, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "alice",
		Password: "correct",
	})
	if err != nil {
		t.Fatalf("期望登录成功，got error: %v", err)
	}

	claims, err := jwt.ParseToken(resp.AccessToken)
	if err != nil {
		t.Fatalf("期望 access token 可解析，got error: %v", err)
	}
	if claims.TenantID != u.TenantID {
		t.Errorf("期望 token 包含 tenant_id=%s，got %s", u.TenantID, claims.TenantID)
	}
}

// --- RefreshToken & Logout tests ---

func TestRefreshToken_Success(t *testing.T) {
	svc, repo, store := setupUserService(t)
	u := makeActiveUser(t, "alice", "pass")
	repo.usersByName["alice"] = u
	repo.usersByID[u.ID] = u

	loginResp, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "alice",
		Password: "pass",
	})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := svc.RefreshToken(context.Background(), loginResp.RefreshToken)
	if err != nil {
		t.Fatalf("期望刷新成功，got error: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("期望返回新的 access token")
	}
	// 旧 token 应被撤销
	exists, _ := store.Exists(context.Background(), loginResp.RefreshToken)
	if exists {
		t.Error("旧 refresh token 应已被撤销")
	}
}

func TestRefreshToken_Invalid(t *testing.T) {
	svc, _, _ := setupUserService(t)
	_, err := svc.RefreshToken(context.Background(), "invalid-token")
	if !errors.Is(err, apperr.ErrRefreshToken) {
		t.Errorf("期望 ErrRefreshToken，got %v", err)
	}
}

func TestRefreshToken_BannedAccount(t *testing.T) {
	svc, repo, _ := setupUserService(t)
	u := makeActiveUser(t, "alice", "pass")
	repo.usersByName["alice"] = u
	repo.usersByID[u.ID] = u

	loginResp, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "alice",
		Password: "pass",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Ban the user after login
	u.Status = "banned"

	_, err = svc.RefreshToken(context.Background(), loginResp.RefreshToken)
	if !errors.Is(err, apperr.ErrAccountDisabled) {
		t.Errorf("期望 ErrAccountDisabled，got %v", err)
	}
}

func TestLogout_Success(t *testing.T) {
	svc, repo, store := setupUserService(t)
	u := makeActiveUser(t, "alice", "pass")
	repo.usersByName["alice"] = u
	repo.usersByID[u.ID] = u

	loginResp, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "alice",
		Password: "pass",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.Logout(context.Background(), loginResp.RefreshToken); err != nil {
		t.Fatalf("期望 logout 成功，got error: %v", err)
	}
	exists, _ := store.Exists(context.Background(), loginResp.RefreshToken)
	if exists {
		t.Error("logout 后 refresh token 应已被删除")
	}
}

// --- Profile tests ---

func TestGetProfile_Success(t *testing.T) {
	svc, repo, _ := setupUserService(t)
	u := makeActiveUser(t, "alice", "pass")
	repo.usersByID[u.ID] = u

	profile, err := svc.GetProfile(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("期望获取 profile 成功，got error: %v", err)
	}
	if profile.Username != "alice" {
		t.Errorf("期望 username=alice，got %s", profile.Username)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	svc, _, _ := setupUserService(t)
	_, err := svc.GetProfile(context.Background(), uuid.New())
	if !errors.Is(err, apperr.ErrUserNotFound) {
		t.Errorf("期望 ErrUserNotFound，got %v", err)
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	svc, repo, _ := setupUserService(t)
	u := makeActiveUser(t, "alice", "pass")
	repo.usersByName["alice"] = u
	repo.usersByID[u.ID] = u

	fullName := "Alice Doe"
	phone := "13800138000"

	profile, err := svc.UpdateProfile(context.Background(), u.ID, &dto.UpdateProfileRequest{
		FullName: &fullName,
		Phone:    &phone,
	})
	if err != nil {
		t.Fatalf("期望更新 profile 成功，got error: %v", err)
	}
	if profile.Username != "alice" {
		t.Errorf("期望 username=alice，got %s", profile.Username)
	}
	if repo.usersByID[u.ID].FullName == nil || *repo.usersByID[u.ID].FullName != fullName {
		t.Error("期望仓库中的 full_name 被更新")
	}
	if repo.usersByID[u.ID].Phone == nil || *repo.usersByID[u.ID].Phone != phone {
		t.Error("期望仓库中的 phone 被更新")
	}
}

func TestUpdateProfile_NotFound(t *testing.T) {
	svc, _, _ := setupUserService(t)
	fullName := "Alice Doe"

	_, err := svc.UpdateProfile(context.Background(), uuid.New(), &dto.UpdateProfileRequest{
		FullName: &fullName,
	})
	if !errors.Is(err, apperr.ErrUserNotFound) {
		t.Errorf("期望 ErrUserNotFound，got %v", err)
	}
}

func TestListUsers_Success(t *testing.T) {
	svc, repo, _ := setupUserService(t)
	userA := makeActiveUser(t, "alice", "pass")
	userB := makeActiveUser(t, "bob", "pass")
	repo.usersByID[userA.ID] = userA
	repo.usersByID[userB.ID] = userB
	repo.usersByName[userA.Username] = userA
	repo.usersByName[userB.Username] = userB

	resp, err := svc.ListUsers(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("期望列出用户成功，got error: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("期望 total=2，got %d", resp.Total)
	}
	if len(resp.Items) != 2 {
		t.Errorf("期望 items 长度=2，got %d", len(resp.Items))
	}
}

func TestRegister_SendsWelcomeEmail(t *testing.T) {
	_ = jwt.SetDefaultSecret("test-secret-at-least-32-characters-long")
	repo := newMockUserRepo()
	store := newMockRefreshStore()
	notifier := &mockNotifier{}
	svc := &userService{
		userRepo:     repo,
		refreshStore: store,
		notifier:     notifier,
	}

	_, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("期望注册成功，got error: %v", err)
	}
	if notifier.sendCalls != 1 {
		t.Fatalf("期望发送 1 封欢迎邮件，got %d", notifier.sendCalls)
	}
	if notifier.lastTo != "alice@example.com" {
		t.Fatalf("期望邮件收件人为 alice@example.com，got %s", notifier.lastTo)
	}
}

func TestRequestPasswordReset_SendsResetEmail(t *testing.T) {
	_ = jwt.SetDefaultSecret("test-secret-at-least-32-characters-long")
	repo := newMockUserRepo()
	store := newMockRefreshStore()
	notifier := &mockNotifier{}
	user := makeActiveUser(t, "alice", "password123")
	user.Email = "alice@example.com"
	user.TenantID = uuid.New()
	repo.usersByName[user.Username] = user
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.ID] = user

	svc := &userService{
		userRepo:     repo,
		refreshStore: store,
		notifier:     notifier,
	}

	err := svc.RequestPasswordReset(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("期望发起重置成功，got error: %v", err)
	}
	if notifier.sendCalls != 1 {
		t.Fatalf("期望发送 1 封重置邮件，got %d", notifier.sendCalls)
	}
	if notifier.lastTo != user.Email {
		t.Fatalf("期望重置邮件收件人为 %s，got %s", user.Email, notifier.lastTo)
	}
	if notifier.lastBody == "" {
		t.Fatal("期望重置邮件正文包含 token")
	}
}

func TestResetPassword_UpdatesStoredPassword(t *testing.T) {
	_ = jwt.SetDefaultSecret("test-secret-at-least-32-characters-long")
	repo := newMockUserRepo()
	store := newMockRefreshStore()
	notifier := &mockNotifier{}
	user := makeActiveUser(t, "alice", "old-password")
	user.Email = "alice@example.com"
	user.TenantID = uuid.New()
	repo.usersByName[user.Username] = user
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.ID] = user

	svc := &userService{
		userRepo:     repo,
		refreshStore: store,
		notifier:     notifier,
	}

	err := svc.RequestPasswordReset(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("期望发起重置成功，got error: %v", err)
	}

	err = svc.ResetPassword(context.Background(), notifier.lastBody, "new-password")
	if err != nil {
		t.Fatalf("期望重置密码成功，got error: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(repo.usersByID[user.ID].Password), []byte("new-password"))
	if err != nil {
		t.Fatalf("期望密码已更新，got compare error: %v", err)
	}
}

// --- Admin tests ---

func TestSetUserStatus(t *testing.T) {
	svc, repo, _ := setupUserService(t)
	u := makeActiveUser(t, "alice", "pass")
	repo.usersByName["alice"] = u
	repo.usersByID[u.ID] = u

	if err := svc.SetUserStatus(context.Background(), u.ID, "banned"); err != nil {
		t.Fatalf("期望成功，got error: %v", err)
	}
	if repo.usersByID[u.ID].Status != "banned" {
		t.Error("期望 status=banned")
	}
}

func TestSetUserRole(t *testing.T) {
	svc, repo, _ := setupUserService(t)
	u := makeActiveUser(t, "alice", "pass")
	repo.usersByName["alice"] = u
	repo.usersByID[u.ID] = u

	if err := svc.SetUserRole(context.Background(), u.ID, "admin"); err != nil {
		t.Fatalf("期望成功，got error: %v", err)
	}
	if repo.usersByID[u.ID].Role != "admin" {
		t.Error("期望 role=admin")
	}
}
