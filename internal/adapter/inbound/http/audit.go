package http

import (
	"errors"

	inhttp "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/middleware"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/service"
	"github.com/Jonathan0823/auth-go/internal/platform"
	"github.com/gin-gonic/gin"
)

func (h *Handler) audit(c *gin.Context, name, outcome, reason, provider string, userID int) {
	if h.Audit == nil {
		return
	}
	h.Audit.Log(c.Request.Context(), platform.AuditEvent{
		Name:      name,
		Outcome:   outcome,
		RequestID: inhttp.RequestIDFromContext(c.Request.Context()),
		UserID:    userID,
		Provider:  provider,
		Reason:    reason,
	})
}

func (h *Handler) auditFailure(c *gin.Context, name, provider string, err error, userID int) {
	h.audit(c, name, platform.OutcomeFailure, auditReason(err), provider, userID)
}

func (h *Handler) auditSuccess(c *gin.Context, name, provider string, userID int) {
	h.audit(c, name, platform.OutcomeSuccess, platform.ReasonNone, provider, userID)
}

func isRefreshReplay(err error) bool {
	return errors.Is(err, service.ErrRefreshTokenReused)
}

func auditReason(err error) string {
	switch {
	case errors.Is(err, service.ErrRefreshTokenReused):
		return platform.ReasonReplayDetected
	case errors.Is(err, domain.ErrInvalidInput):
		return platform.ReasonValidation
	case errors.Is(err, domain.ErrUnauthenticated):
		return platform.ReasonUnauthenticated
	case errors.Is(err, domain.ErrNotFound):
		return platform.ReasonNotFound
	case errors.Is(err, domain.ErrConflict):
		return platform.ReasonConflict
	default:
		return platform.ReasonInternal
	}
}
