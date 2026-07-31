package platform

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	Registry       *prometheus.Registry
	httpRequests   *prometheus.CounterVec
	httpDuration   *prometheus.HistogramVec
	httpInFlight   prometheus.Gauge
	securityEvents *prometheus.CounterVec
}

func NewMetrics(pool *pgxpool.Pool) *Metrics {
	metrics := &Metrics{
		Registry: prometheus.NewRegistry(),
		httpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "auth_go",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		}, []string{"method", "route", "status"}),
		httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "auth_go",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "route", "status"}),
		httpInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "auth_go",
			Subsystem: "http",
			Name:      "in_flight_requests",
			Help:      "Current number of in-flight HTTP requests.",
		}),
		securityEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "auth_go",
			Name:      "security_events_total",
			Help:      "Total number of security audit events.",
		}, []string{"event", "outcome", "reason", "provider"}),
	}

	metrics.Registry.MustRegister(metrics.httpRequests, metrics.httpDuration, metrics.httpInFlight, metrics.securityEvents)
	if pool != nil {
		metrics.Registry.MustRegister(newPoolCollector(pool))
	}
	return metrics
}

func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.httpInFlight.Inc()
		start := prometheus.NewTimer(prometheus.ObserverFunc(func(duration float64) {
			method, route, status := requestLabels(c)
			m.httpRequests.WithLabelValues(method, route, status).Inc()
			m.httpDuration.WithLabelValues(method, route, status).Observe(duration)
		}))
		c.Next()
		start.ObserveDuration()
		m.httpInFlight.Dec()
	}
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}

func RegisterMetricsRoute(r *gin.Engine, metrics *Metrics, enabled bool) {
	if !enabled || metrics == nil {
		return
	}
	r.GET("/metrics", gin.WrapH(metrics.Handler()))
}

func (m *Metrics) RecordSecurityEvent(event, outcome, reason, provider string) {
	m.securityEvents.WithLabelValues(
		normalizeEvent(event),
		normalizeOutcome(outcome),
		normalizeReason(reason),
		normalizeProvider(provider),
	).Inc()
}

func requestLabels(c *gin.Context) (string, string, string) {
	route := c.FullPath()
	if route == "" {
		route = "/__unmatched__"
	}
	return c.Request.Method, route, strconv.Itoa(c.Writer.Status())
}

func normalizeEvent(event string) string {
	switch event {
	case EventUserRegister, EventAuthLogin, EventAuthLogout, EventAuthRefresh,
		EventAuthRefreshReplay, EventAuthPasswordResetRequest, EventAuthPasswordReset,
		EventAuthEmailVerification, EventAuthOAuth:
		return event
	default:
		return "unknown"
	}
}

func normalizeOutcome(outcome string) string {
	switch outcome {
	case OutcomeSuccess, OutcomeFailure, OutcomeDetected:
		return outcome
	default:
		return "unknown"
	}
}

func normalizeReason(reason string) string {
	switch reason {
	case ReasonNone, ReasonInvalidCredentials, ReasonValidation, ReasonUnauthenticated,
		ReasonNotFound, ReasonConflict, ReasonInvalidToken, ReasonProviderError,
		ReasonReplayDetected, ReasonInternal:
		return reason
	default:
		return "unknown"
	}
}

func normalizeProvider(provider string) string {
	switch provider {
	case "github", "google":
		return provider
	default:
		return "none"
	}
}

type poolCollector struct {
	pool  *pgxpool.Pool
	descs []*prometheus.Desc
}

func newPoolCollector(pool *pgxpool.Pool) prometheus.Collector {
	return &poolCollector{
		pool: pool,
		descs: []*prometheus.Desc{
			prometheus.NewDesc("auth_go_db_pool_max_connections", "Maximum PostgreSQL pool connections.", nil, nil),
			prometheus.NewDesc("auth_go_db_pool_total_connections", "Total PostgreSQL pool connections.", nil, nil),
			prometheus.NewDesc("auth_go_db_pool_acquired_connections", "Acquired PostgreSQL pool connections.", nil, nil),
			prometheus.NewDesc("auth_go_db_pool_idle_connections", "Idle PostgreSQL pool connections.", nil, nil),
			prometheus.NewDesc("auth_go_db_pool_constructing_connections", "PostgreSQL pool connections being constructed.", nil, nil),
			prometheus.NewDesc("auth_go_db_pool_acquires_total", "Total PostgreSQL pool acquire operations.", nil, nil),
			prometheus.NewDesc("auth_go_db_pool_canceled_acquires_total", "Total canceled PostgreSQL pool acquire operations.", nil, nil),
			prometheus.NewDesc("auth_go_db_pool_empty_acquires_total", "Total PostgreSQL pool acquire operations that found no idle connection.", nil, nil),
			prometheus.NewDesc("auth_go_db_pool_acquire_duration_seconds_total", "Total time spent acquiring PostgreSQL pool connections.", nil, nil),
		},
	}
}

func (c *poolCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.descs {
		ch <- desc
	}
}

func (c *poolCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.pool.Stat()
	values := []float64{
		float64(stats.MaxConns()),
		float64(stats.TotalConns()),
		float64(stats.AcquiredConns()),
		float64(stats.IdleConns()),
		float64(stats.ConstructingConns()),
		float64(stats.AcquireCount()),
		float64(stats.CanceledAcquireCount()),
		float64(stats.EmptyAcquireCount()),
		stats.AcquireDuration().Seconds(),
	}
	for i, value := range values {
		metricType := prometheus.GaugeValue
		if i >= 5 {
			metricType = prometheus.CounterValue
		}
		ch <- prometheus.MustNewConstMetric(c.descs[i], metricType, value)
	}
}
