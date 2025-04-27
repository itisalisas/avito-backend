package middleware

import (
	"net/http"
	"time"

	"github.com/itisalisas/avito-backend/pkg/metrics"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start).Seconds()
		status := http.StatusText(lrw.statusCode)

		metrics.HttpRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			status,
		).Inc()

		metrics.HttpResponseTime.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(duration)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
