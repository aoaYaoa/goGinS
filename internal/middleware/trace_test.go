package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aoaYaoa/go-gin-starter/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestTrace_AddsTraceIDHeaderAndContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	r := gin.New()
	r.Use(middleware.Trace("test-service"))
	r.Use(middleware.TraceContext())
	r.GET("/", func(c *gin.Context) {
		traceID, _ := c.Get("trace_id")
		c.String(http.StatusOK, traceID.(string))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	traceHeader := w.Header().Get(middleware.HeaderTraceID)
	if traceHeader == "" {
		t.Fatal("expected X-Trace-ID response header to be set")
	}
	if body := w.Body.String(); body == "" || body != traceHeader {
		t.Fatalf("expected body trace id to equal header, got body=%q header=%q", body, traceHeader)
	}
}
