package env

import (
	"log"
	"os"
)

// TracingEnvironment holds configuration for OpenTelemetry tracing
type TracingEnvironment struct {
	Enabled  bool   // No validation needed - bools are always true or false
	Endpoint string `validate:"omitempty,required_if=Enabled true,url"`
}

// NewTracingEnvironment loads tracing configuration from environment variables
func NewTracingEnvironment() *TracingEnvironment {
	enabled := os.Getenv("ENV_TRACING_ENABLED") == "true"
	endpoint := os.Getenv("ENV_TRACING_OTLP_ENDPOINT")

	// Default endpoint if tracing is enabled but no endpoint specified
	if enabled && endpoint == "" {
		endpoint = "http://localhost:4318"
		log.Printf("tracing enabled but ENV_TRACING_OTLP_ENDPOINT not set, using default: %s", endpoint)
	}

	return &TracingEnvironment{
		Enabled:  enabled,
		Endpoint: endpoint,
	}
}
