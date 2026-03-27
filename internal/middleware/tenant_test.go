package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aoaYaoa/go-gin-starter/internal/middleware"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tenantctx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTenantResolver_UsesClaimsTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tenantID := uuid.New()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("claims", &jwt.Claims{TenantID: tenantID})
		c.Next()
	})
	r.Use(middleware.TenantResolver())
	r.GET("/", func(c *gin.Context) {
		resolved, ok := tenantctx.FromContext(c.Request.Context())
		require.True(t, ok)
		c.String(http.StatusOK, resolved.String())
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, tenantID.String(), w.Body.String())
}

func TestTenantResolver_UsesHeaderFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tenantID := uuid.New()

	r := gin.New()
	r.Use(middleware.TenantResolver())
	r.GET("/", func(c *gin.Context) {
		resolved, ok := tenantctx.FromContext(c.Request.Context())
		require.True(t, ok)
		c.String(http.StatusOK, resolved.String())
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.HeaderTenantID, tenantID.String())
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, tenantID.String(), w.Body.String())
}
