package http

import (
	"net/http"
	"time"

	inhttp "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/middleware"
	"github.com/Jonathan0823/auth-go/internal/core/ratelimit"
	"github.com/Jonathan0823/auth-go/internal/observability"
	"github.com/gin-gonic/gin"
)

const backendRetryAfter = 30 * time.Second

func ipRateLimit(policy string, c *gin.Context) ratelimit.RateLimitCheck {
	return ratelimit.RateLimitCheck{Policy: policy, Dimension: "ip", Value: c.ClientIP()}
}

func (h *Handler) allowRateLimit(c *gin.Context, checks ...ratelimit.RateLimitCheck) bool {
	if h.RateLimiter == nil {
		return true
	}
	decision, err := h.RateLimiter.Allow(c.Request.Context(), checks...)
	if err != nil {
		h.logRateLimitEvent(c, observability.EventRateLimiterUnavailable, observability.OutcomeFailure, observability.ReasonBackendUnavailable, checks)
		c.Header("Retry-After", inhttp.RetryAfterHeader(backendRetryAfter))
		c.Header("Cache-Control", "no-store")
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "authentication temporarily unavailable"})
		return false
	}
	if decision.Allowed {
		return true
	}

	h.logRateLimitEvent(c, observability.EventAuthRateLimited, observability.OutcomeDetected, observability.ReasonRateLimited, checks)
	c.Header("Retry-After", inhttp.RetryAfterHeader(decision.RetryAfter))
	c.Header("Cache-Control", "no-store")
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
	return false
}

func (h *Handler) logRateLimitEvent(c *gin.Context, event, outcome, reason string, checks []ratelimit.RateLimitCheck) {
	if h.Audit == nil || len(checks) == 0 {
		return
	}
	h.Audit.Log(c.Request.Context(), observability.AuditEvent{
		Name:      event,
		Outcome:   outcome,
		RequestID: requestID(c),
		Reason:    reason,
		Route:     c.FullPath(),
		Policy:    checks[0].Policy,
		Backend:   h.RateLimiter.Backend(),
	})
}

func requestID(c *gin.Context) string {
	return inhttp.RequestIDFromContext(c.Request.Context())
}
