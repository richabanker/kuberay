package observability

import (
	"context"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	clientmetrics "k8s.io/client-go/tools/metrics"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// Define your custom Histogram matching upstream standards
	rayRestClientLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rest_client_request_duration_seconds",
			Help:    "Request latency in seconds. Broken down by verb and host.",
			Buckets: []float64{0.005, 0.025, 0.1, 0.25, 0.5, 1.0, 2.0, 4.0, 8.0, 15.0, 30.0, 60.0},
		},
		[]string{"verb", "host"},
	)
)

// Implement the LatencyMetric interface from client-go
type rayLatencyAdapter struct {
	metric *prometheus.HistogramVec
}

func (l *rayLatencyAdapter) Observe(ctx context.Context, verb string, u url.URL, latency time.Duration) {
	l.metric.WithLabelValues(verb, u.Host).Observe(latency.Seconds())
}

func init() {
	// 1. Register your metric with controller-runtime's global prometheus registry
	ctrlmetrics.Registry.MustRegister(rayRestClientLatency)
	// 2. Overwrite the global variable directly (Bypasses clientmetrics.Register lock)
	clientmetrics.RequestLatency = &rayLatencyAdapter{metric: rayRestClientLatency}
}
