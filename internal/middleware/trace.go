package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

const HeaderTraceID = "X-Trace-ID"

// Trace installs otelgin request tracing.
func Trace(serviceName string) gin.HandlerFunc {
	if serviceName == "" {
		serviceName = "go-gin-starter"
	}
	return otelgin.Middleware(serviceName)
}

// TraceContext injects the active trace ID into Gin context and response headers.
func TraceContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		tid := c.GetHeader(HeaderTraceID)
		if spanCtx := trace.SpanContextFromContext(c.Request.Context()); spanCtx.IsValid() {
			tid = spanCtx.TraceID().String()
		}
		if strings.TrimSpace(tid) == "" {
			tid = uuid.New().String()
		}
		c.Set("trace_id", tid)
		c.Header(HeaderTraceID, tid)
		c.Next()
	}
}

// TraceID keeps the legacy name for the trace-ID propagation middleware.
func TraceID() gin.HandlerFunc {
	return TraceContext()
}
