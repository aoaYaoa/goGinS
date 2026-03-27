package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aoaYaoa/go-gin-starter/internal/middleware"
	authzutil "github.com/aoaYaoa/go-gin-starter/pkg/utils/authz"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubAuthorizer struct {
	allow bool
}

func (s *stubAuthorizer) Check(_, _, _ string) (bool, error) {
	return s.allow, nil
}

type reloadableStubAuthorizer struct {
	allow      bool
	reloadCall int
}

func (s *reloadableStubAuthorizer) Reload() error {
	s.reloadCall++
	s.allow = true
	return nil
}

func (s *reloadableStubAuthorizer) Check(_, _, _ string) (bool, error) {
	return s.allow, nil
}

func TestAuthz_AllowsAuthorizedRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authzutil.SetDefault(&stubAuthorizer{allow: true})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("claims", &jwt.Claims{
			UserID: uuid.New(),
			Role:   "admin",
		})
		c.Next()
	})
	r.GET("/", middleware.Authz("users", "read"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthz_RejectsUnauthorizedRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authzutil.SetDefault(&stubAuthorizer{allow: false})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("claims", &jwt.Claims{
			UserID: uuid.New(),
			Role:   "user",
		})
		c.Next()
	})
	r.GET("/", middleware.Authz("users", "write"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestAuthz_ReloadsPolicyBeforeCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authorizer := &reloadableStubAuthorizer{allow: false}
	authzutil.SetDefault(authorizer)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("claims", &jwt.Claims{
			UserID: uuid.New(),
			Role:   "admin",
		})
		c.Next()
	})
	r.GET("/", middleware.Authz("users", "write"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 after reload, got %d", w.Code)
	}
	if authorizer.reloadCall != 1 {
		t.Fatalf("expected reload to be called once, got %d", authorizer.reloadCall)
	}
}
