package portal

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/oullin/metal/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TracerProvider wraps the OpenTelemetry tracer provider
type TracerProvider struct {
	Provider *sdktrace.TracerProvider
	Env      *env.Environment
}

// NewTracerProvider initializes OpenTelemetry with OTLP HTTP exporter
func NewTracerProvider(environment *env.Environment) (*TracerProvider, error) {
	if !environment.Tracing.Enabled {
		log.Println("OpenTelemetry tracing is disabled")
		return &TracerProvider{
			Provider: nil,
			Env:      environment,
		}, nil
	}

	ctx := context.Background()

	// Create OTLP HTTP exporter with environment-appropriate security settings
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(getEndpointHost(environment.Tracing.Endpoint)),
	}

	// Only use insecure connections for local and staging environments
	if environment.App.IsLocal() || environment.App.IsStaging() {
		opts = append(opts, otlptracehttp.WithInsecure())
		log.Printf("Using insecure OTLP connection for %s environment", environment.App.Type)
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("service.name", environment.App.Name),
			attribute.String("service.version", "1.0.0"),
			attribute.String("deployment.environment", environment.App.Type),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for distributed tracing
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	log.Printf("OpenTelemetry tracing initialized with endpoint: %s", environment.Tracing.Endpoint)

	return &TracerProvider{
		Provider: tp,
		Env:      environment,
	}, nil
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown() error {
	if tp.Provider == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tp.Provider.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown tracer provider: %w", err)
	}

	log.Println("OpenTelemetry tracer provider shutdown successfully")
	return nil
}

// getEndpointHost extracts the host:port from a full URL
// e.g., "http://localhost:4318" -> "localhost:4318"
func getEndpointHost(endpoint string) string {
	// Remove http:// or https:// prefix
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		return endpoint[7:]
	}
	if len(endpoint) > 8 && endpoint[:8] == "https://" {
		return endpoint[8:]
	}
	return endpoint
}
