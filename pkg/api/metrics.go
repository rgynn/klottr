package api

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricServedRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"method", "path", "query", "code"},
	)
	metricDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "How long did the processed HTTP Requests Take",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "code"})
)

func setupMetrics() {
	prometheus.MustRegister(metricServedRequests)
	prometheus.MustRegister(metricDurationSeconds)
}
