package tracing

import (
	"gorm.io/gorm"
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
)

// NewGormPlugin returns a GORM plugin that creates an OTel child span for
// every database operation, linking it to the incoming request trace.
func NewGormPlugin() gorm.Plugin {
	return otelgorm.NewPlugin()
}
