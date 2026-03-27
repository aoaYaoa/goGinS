package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aoaYaoa/go-gin-starter/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func testRouter(mw gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(mw)
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func TestIPWhitelist_ExactMatch(t *testing.T) {
	r := testRouter(middleware.IPWhitelist([]string{"127.0.0.1"}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:9999"
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIPWhitelist_CIDRMatch(t *testing.T) {
	r := testRouter(middleware.IPWhitelist([]string{"10.0.0.0/8"}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.1.2.3:9999"
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIPWhitelist_CIDRBlock(t *testing.T) {
	r := testRouter(middleware.IPWhitelist([]string{"10.0.0.0/8"}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:9999"
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestIPWhitelist_Empty_AllowsAll(t *testing.T) {
	r := testRouter(middleware.IPWhitelist([]string{}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:9999"
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIPBlacklist_ExactBlock(t *testing.T) {
	r := testRouter(middleware.IPBlacklist([]string{"192.168.1.1"}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:9999"
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestIPBlacklist_CIDRBlock(t *testing.T) {
	r := testRouter(middleware.IPBlacklist([]string{"192.168.0.0/16"}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.5.5:9999"
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestIPBlacklist_CIDRAllow(t *testing.T) {
	r := testRouter(middleware.IPBlacklist([]string{"192.168.0.0/16"}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
