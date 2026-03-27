package tracing_test

import (
	"context"
	"testing"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/tracing"
	"github.com/stretchr/testify/require"
)

func TestInit_UsesStdoutExporterWhenEndpointMissing(t *testing.T) {
	shutdown, err := tracing.Init(tracing.Config{
		ServiceName: "go-gin-starter-test",
	})
	require.NoError(t, err)
	require.NotNil(t, shutdown)
	require.NoError(t, shutdown(context.Background()))
}
