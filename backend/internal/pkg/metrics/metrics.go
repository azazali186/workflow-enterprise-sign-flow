// Package metrics defines Prometheus collectors for the application.
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collectors holds all custom metric vectors.
type Collectors struct {
	HTTPRequests   *prometheus.CounterVec
	HTTPDuration   *prometheus.HistogramVec
	OutboxEvents   *prometheus.CounterVec
	RBACChecks     *prometheus.CounterVec
	LockContention *prometheus.CounterVec
}

var (
	mOnce     sync.Once
	mInstance *Collectors
)

// New registers and returns the application metrics. Registration is
// idempotent so tests and restarts never panic on duplicate collectors.
func New() *Collectors {
	mOnce.Do(func() { mInstance = newCollectors() })
	return mInstance
}

func newCollectors() *Collectors {
	return &Collectors{
		HTTPRequests: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "signflow_http_requests_total",
			Help: "Total HTTP requests by method, path and status.",
		}, []string{"method", "path", "status"}),
		HTTPDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "signflow_http_request_duration_seconds",
			Help:    "HTTP request latency histogram.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),
		OutboxEvents: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "signflow_outbox_events_total",
			Help: "Outbox events by state.",
		}, []string{"state"}),
		RBACChecks: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "signflow_rbac_checks_total",
			Help: "RBAC decisions.",
		}, []string{"result"}),
		LockContention: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "signflow_lock_contention_total",
			Help: "Failed distributed lock acquisitions by resource.",
		}, []string{"resource"}),
	}
}
