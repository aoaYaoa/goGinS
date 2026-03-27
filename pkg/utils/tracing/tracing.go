// Package tracing initializes the OpenTelemetry TracerProvider.
// When OTEL_ENDPOINT is set, spans are exported via OTLP gRPC;
// otherwise they are printed to stdout (useful for local development).
package tracing

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

type Config struct {
	Endpoint    string
	ServiceName string
}

// Init sets up the global TracerProvider and TextMapPropagator.
// It returns a shutdown function that must be called on application exit
// to flush and export any remaining spans.
func Init(cfg Config) (func(context.Context) error, error) {
	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "go-gin-starter"
	}

	exporter, err := newExporter(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

func newExporter(endpoint string) (sdktrace.SpanExporter, error) {
	ctx := context.Background()
	if endpoint == "" {
		return stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	}

	opts := []otlptracegrpc.Option{}
	if strings.Contains(endpoint, "://") {
		opts = append(opts, otlptracegrpc.WithEndpointURL(endpoint))
		if strings.HasPrefix(endpoint, "http://") {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
	} else {
		opts = append(opts, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure())
	}

	return otlptracegrpc.New(ctx, opts...)
}
