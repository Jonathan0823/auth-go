package platform

import (
	"context"
	"log/slog"
)

const (
	EventUserRegister             = "user.register"
	EventAuthLogin                = "auth.login"
	EventAuthLogout               = "auth.logout"
	EventAuthRefresh              = "auth.refresh"
	EventAuthRefreshReplay        = "auth.refresh_replay"
	EventAuthPasswordResetRequest = "auth.password_reset_request"
	EventAuthPasswordReset        = "auth.password_reset"
	EventAuthEmailVerification    = "auth.email_verification"
	EventAuthOAuth                = "auth.oauth"

	OutcomeSuccess  = "success"
	OutcomeFailure  = "failure"
	OutcomeDetected = "detected"

	ReasonNone               = "none"
	ReasonInvalidCredentials = "invalid_credentials"
	ReasonValidation         = "validation"
	ReasonUnauthenticated    = "unauthenticated"
	ReasonNotFound           = "not_found"
	ReasonConflict           = "conflict"
	ReasonInvalidToken       = "invalid_token"
	ReasonProviderError      = "provider_error"
	ReasonReplayDetected     = "replay_detected"
	ReasonInternal           = "internal"
)

type AuditEvent struct {
	Name      string
	Outcome   string
	RequestID string
	UserID    int
	Provider  string
	Reason    string
}

type AuditLogger struct {
	logger  *slog.Logger
	metrics *Metrics
}

func NewAuditLogger(logger *slog.Logger, metrics *Metrics) *AuditLogger {
	return &AuditLogger{logger: logger, metrics: metrics}
}

func (a *AuditLogger) Log(ctx context.Context, event AuditEvent) {
	if a == nil || a.logger == nil {
		return
	}

	attrs := []slog.Attr{
		Redact("event", normalizeEvent(event.Name)),
		Redact("outcome", normalizeOutcome(event.Outcome)),
		Redact("reason", normalizeReason(event.Reason)),
		Redact("provider", normalizeProvider(event.Provider)),
	}
	if event.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", event.RequestID))
	}
	if event.UserID > 0 {
		attrs = append(attrs, slog.Int("user_id", event.UserID))
	}

	a.logger.LogAttrs(ctx, slog.LevelInfo, "security audit", attrs...)
	if a.metrics != nil {
		a.metrics.RecordSecurityEvent(event.Name, event.Outcome, event.Reason, event.Provider)
	}
}
