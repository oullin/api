package handler

import (
	"net/http"

	"github.com/oullin/pkg/endpoint"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsHandler struct{}

func NewMetricsHandler() MetricsHandler {
	return MetricsHandler{}
}

// Handle returns the Prometheus metrics handler (public endpoint for Prometheus scraping)
func (h MetricsHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	// Serve Prometheus metrics using the standard promhttp handler
	promhttp.Handler().ServeHTTP(w, r)
	return nil
}
