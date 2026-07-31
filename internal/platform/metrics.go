package platform

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	Registry        *prometheus.Registry
	httpRequests    *prometheus.CounterVec
	httpDuration    *prometheus.HistogramVec
	httpInFlight    prometheus.Gauge
	securityEvents  *prometheus.CounterVec
	rateLimitEvents *prometheus.CounterVec
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
		rateLimitEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "auth_go",
			Name:      "rate_limit_events_total",
			Help:      "Total number of rate-limit events.",
		}, []string{"route", "policy", "backend", "outcome", "reason"}),
	}

	metrics.Registry.MustRegister(metrics.httpRequests, metrics.httpDuration, metrics.httpInFlight, metrics.securityEvents, metrics.rateLimitEvents)
	if pool != nil {
		metrics.Registry.MustRegister(newPoolCollector(pool))
	}
	return metrics
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}

func (m *Metrics) IncInFlight() {
	m.httpInFlight.Inc()
}

func (m *Metrics) DecInFlight() {
	m.httpInFlight.Dec()
}

func (m *Metrics) ObserveHTTPRequest(method, route, status string, duration time.Duration) {
	m.httpRequests.WithLabelValues(method, route, status).Inc()
	m.httpDuration.WithLabelValues(method, route, status).Observe(duration.Seconds())
}

func (m *Metrics) RecordSecurityEvent(event, outcome, reason, provider string) {
	m.securityEvents.WithLabelValues(
		normalizeEvent(event),
		normalizeOutcome(outcome),
		normalizeReason(reason),
		normalizeProvider(provider),
	).Inc()
}

func (m *Metrics) RecordRateLimitEvent(route, policy, backend, outcome, reason string) {
	m.rateLimitEvents.WithLabelValues(
		normalizeRoute(route),
		normalizePolicy(policy),
		normalizeBackend(backend),
		normalizeOutcome(outcome),
		normalizeReason(reason),
	).Inc()
}

func normalizeEvent(event string) string {
	switch event {
	case EventUserRegister, EventAuthLogin, EventAuthLogout, EventAuthRefresh,
		EventAuthRefreshReplay, EventAuthPasswordResetRequest, EventAuthPasswordReset,
		EventAuthEmailVerification, EventAuthOAuth, EventAuthRateLimited,
		EventRateLimiterUnavailable:
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
		ReasonReplayDetected, ReasonInternal, ReasonRateLimited, ReasonBackendUnavailable:
		return reason
	default:
		return "unknown"
	}
}

func normalizeRoute(route string) string {
	switch route {
	case "/api/auth/login", "/api/auth/register", "/api/auth/refresh", "/api/auth/forgot-password", "/api/auth/reset-password", "/api/auth/verify/email/resend", "/api/oauth/:provider/", "/api/oauth/:provider/callback":
		return route
	default:
		return "unknown"
	}
}

func normalizePolicy(policy string) string {
	switch policy {
	case "login_ip", "login_account", "register_ip", "register_email", "recovery_ip", "recovery_email", "recovery_token", "refresh_ip", "verify_ip", "verify_email", "oauth_ip":
		return policy
	default:
		return "unknown"
	}
}

func normalizeBackend(backend string) string {
	switch backend {
	case "memory", "redis", "postgres":
		return backend
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
