package handlers_test

import (
	"net/http"
	"testing"

	"github.com/aoaYaoa/go-gin-starter/internal/database"
	"github.com/aoaYaoa/go-gin-starter/internal/handlers"
	"github.com/aoaYaoa/go-gin-starter/internal/testhelper"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandlerLiveness_Returns200(t *testing.T) {
	h := handlers.NewHealthHandler(nil)
	r := gin.New()
	r.GET("/healthz", h.Liveness)

	w := getReq(r, "/healthz")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
}

func TestHealthHandlerReadiness_Returns200WhenDatabaseReady(t *testing.T) {
	gormDB, cleanup := testhelper.SetupPostgres(t)
	defer cleanup()

	h := handlers.NewHealthHandler(&database.DB{DB: gormDB})
	r := gin.New()
	r.GET("/readyz", h.Readiness)

	w := getReq(r, "/readyz")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ready"`)
}

func TestHealthHandlerReadiness_Returns503WhenDatabaseUnavailable(t *testing.T) {
	gormDB, cleanup := testhelper.SetupPostgres(t)
	defer cleanup()

	wrapped := &database.DB{DB: gormDB}
	assert.NoError(t, wrapped.Close())

	h := handlers.NewHealthHandler(wrapped)
	r := gin.New()
	r.GET("/readyz", h.Readiness)

	w := getReq(r, "/readyz")
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "数据库不可用")
}
