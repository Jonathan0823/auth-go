package http

import (
	"errors"

	inhttp "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/middleware"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/service"
	"github.com/Jonathan0823/auth-go/internal/observability"
	"github.com/gin-gonic/gin"
)

func (h *Handler) audit(c *gin.Context, name, outcome, reason, provider string, userID int) {
	if h.Audit == nil {
		return
	}
	h.Audit.Log(c.Request.Context(), observability.AuditEvent{
		Name:      name,
		Outcome:   outcome,
		RequestID: inhttp.RequestIDFromContext(c.Request.Context()),
		UserID:    userID,
		Provider:  provider,
		Reason:    reason,
	})
}

func (h *Handler) auditFailure(c *gin.Context, name, provider string, err error, userID int) {
	var rateLimitErr *service.RateLimitError
	if h.Audit != nil && errors.As(err, &rateLimitErr) {
		event, outcome, reason := observability.EventAuthRateLimited, observability.OutcomeDetected, observability.ReasonRateLimited
		if rateLimitErr.Unavailable {
			event, outcome, reason = observability.EventRateLimiterUnavailable, observability.OutcomeFailure, observability.ReasonBackendUnavailable
		}
		h.Audit.Log(c.Request.Context(), observability.AuditEvent{
			Name: event, Outcome: outcome, RequestID: inhttp.RequestIDFromContext(c.Request.Context()),
			UserID: userID, Provider: provider, Reason: reason, Route: c.FullPath(),
			Policy: rateLimitErr.Policy, Backend: rateLimitErr.Backend,
		})
		return
	}
	h.audit(c, name, observability.OutcomeFailure, auditReason(err), provider, userID)
}

func (h *Handler) auditSuccess(c *gin.Context, name, provider string, userID int) {
	h.audit(c, name, observability.OutcomeSuccess, observability.ReasonNone, provider, userID)
}

func isRefreshReplay(err error) bool {
	return errors.Is(err, service.ErrRefreshTokenReused)
}

func auditReason(err error) string {
	switch {
	case errors.Is(err, service.ErrRefreshTokenReused):
		return observability.ReasonReplayDetected
	case errors.Is(err, domain.ErrInvalidInput):
		return observability.ReasonValidation
	case errors.Is(err, domain.ErrUnauthenticated):
		return observability.ReasonUnauthenticated
	case errors.Is(err, domain.ErrNotFound):
		return observability.ReasonNotFound
	case errors.Is(err, domain.ErrConflict):
		return observability.ReasonConflict
	default:
		return observability.ReasonInternal
	}
}
