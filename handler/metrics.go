package handler

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsHandler struct{}

func NewMetricsHandler() MetricsHandler {
	return MetricsHandler{}
}

// Handle returns the Prometheus metrics handler
// This bypasses the normal API error handling since Prometheus uses its own format
func (h MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}
