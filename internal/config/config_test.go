package config_test

import (
	"testing"

	"github.com/aoaYaoa/go-gin-starter/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestInit_LoadsOTelConfigFromEnv(t *testing.T) {
	t.Setenv("OTEL_ENDPOINT", "http://localhost:4318")
	t.Setenv("OTEL_SERVICE_NAME", "go-gin-starter-test")

	config.Init()

	assert.Equal(t, "http://localhost:4318", config.AppConfig.OTELEndpoint)
	assert.Equal(t, "go-gin-starter-test", config.AppConfig.OTELServiceName)
}
