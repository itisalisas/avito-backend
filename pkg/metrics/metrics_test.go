package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestHttpRequestsTotalCounter(t *testing.T) {
	HttpRequestsTotal.WithLabelValues("GET", "/test", "200").Inc()

	metric := HttpRequestsTotal.WithLabelValues("GET", "/test", "200")
	metricValue := testutil.ToFloat64(metric)

	assert.Equal(t, float64(1), metricValue)
}
