package http

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	inhttp "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/middleware"
	"github.com/Jonathan0823/auth-go/internal/platform"
	"github.com/gin-gonic/gin"
)

const backendRetryAfter = 30 * time.Second

func ipRateLimit(policy string, c *gin.Context) platform.RateLimitCheck {
	return platform.RateLimitCheck{Policy: policy, Dimension: "ip", Value: c.ClientIP()}
}

func emailRateLimit(policy, email string) platform.RateLimitCheck {
	return platform.RateLimitCheck{Policy: policy, Dimension: "email", Value: strings.ToLower(strings.TrimSpace(email))}
}

func tokenRateLimit(policy, token string) platform.RateLimitCheck {
	return platform.RateLimitCheck{Policy: policy, Dimension: "token", Value: token}
}

func (h *Handler) allowRateLimit(c *gin.Context, checks ...platform.RateLimitCheck) bool {
	if h.RateLimiter == nil {
		return true
	}
	decision, err := h.RateLimiter.Allow(c.Request.Context(), checks...)
	if err != nil {
		h.logRateLimitEvent(c, platform.EventRateLimiterUnavailable, platform.OutcomeFailure, platform.ReasonBackendUnavailable, checks)
		c.Header("Retry-After", retryAfterHeader(backendRetryAfter))
		c.Header("Cache-Control", "no-store")
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "authentication temporarily unavailable"})
		return false
	}
	if decision.Allowed {
		return true
	}

	h.logRateLimitEvent(c, platform.EventAuthRateLimited, platform.OutcomeDetected, platform.ReasonRateLimited, checks)
	c.Header("Retry-After", retryAfterHeader(decision.RetryAfter))
	c.Header("Cache-Control", "no-store")
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
	return false
}

func (h *Handler) resetRateLimit(c *gin.Context, check platform.RateLimitCheck) {
	if h.RateLimiter == nil {
		return
	}
	if err := h.RateLimiter.Reset(c.Request.Context(), check); err != nil {
		h.logRateLimitEvent(c, platform.EventRateLimiterUnavailable, platform.OutcomeFailure, platform.ReasonBackendUnavailable, []platform.RateLimitCheck{check})
	}
}

func (h *Handler) logRateLimitEvent(c *gin.Context, event, outcome, reason string, checks []platform.RateLimitCheck) {
	if h.Audit == nil || len(checks) == 0 {
		return
	}
	h.Audit.Log(c.Request.Context(), platform.AuditEvent{
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

func retryAfterHeader(duration time.Duration) string {
	seconds := int64(math.Ceil(duration.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	if seconds > 3600 {
		seconds = 3600
	}
	return strconv.FormatInt(seconds, 10)
}
