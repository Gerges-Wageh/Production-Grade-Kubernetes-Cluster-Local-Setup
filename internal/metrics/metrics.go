package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "posts_service_requests_total",
			Help: "Total number of requests",
		},
		[]string{"endpoint", "method", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "posts_service_request_duration_seconds",
			Help:    "Request latency in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms -> ~16s
		},
		[]string{"endpoint", "method"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(RequestsTotal, RequestDuration)
}
