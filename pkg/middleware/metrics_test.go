package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/itisalisas/avito-backend/pkg/metrics"
)

func TestMetricsMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	mw := MetricsMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	rr := httptest.NewRecorder()

	// Serve the request
	mw.ServeHTTP(rr, req)

	// Check if the status code is correct
	assert.Equal(t, http.StatusOK, rr.Code)

	// Test if the HttpRequestsTotal counter has been incremented
	metric := metrics.HttpRequestsTotal.WithLabelValues(http.MethodGet, "/test-path", "OK")
	metricValue := testutil.ToFloat64(metric)
	assert.Equal(t, float64(1), metricValue)
}
