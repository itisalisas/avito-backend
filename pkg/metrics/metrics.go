package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var Registry = prometheus.NewRegistry()

var (
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HttpResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_response_time_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: []float64{0.01, 0.03, 0.05, 0.1, 0.25, 0.5},
	}, []string{"method", "path"})

	PVZCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pvz_created_total",
		Help: "Total number of PVZ created",
	})

	OrderReceptionsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "order_receptions_created_total",
		Help: "Total number of order receptions created",
	})

	ProductsAdded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_added_total",
		Help: "Total number of products added",
	})
)

func init() {
	Registry.MustRegister(
		HttpRequestsTotal,
		HttpResponseTime,
		PVZCreated,
		OrderReceptionsCreated,
		ProductsAdded,
	)
}
