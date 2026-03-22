package middleware

import (
	"net/http"
	"time"

	"posts/internal/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 200}

		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()
		endpoint := r.URL.Path

		metrics.RequestsTotal.WithLabelValues(endpoint, r.Method, http.StatusText(rec.status)).Inc()
		metrics.RequestDuration.WithLabelValues(endpoint, r.Method).Observe(duration)
	})
}
