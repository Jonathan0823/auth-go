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
	EventAuthRateLimited          = "auth.rate_limited"
	EventRateLimiterUnavailable   = "auth.rate_limiter_unavailable"

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
	ReasonRateLimited        = "rate_limited"
	ReasonBackendUnavailable = "backend_unavailable"
)

type AuditEvent struct {
	Name      string
	Outcome   string
	RequestID string
	UserID    int
	Provider  string
	Reason    string
	Route     string
	Policy    string
	Backend   string
}

type AuditLogger struct {
	logger  *slog.Logger
	metrics *Metrics
}

func NewAuditLogger(logger *slog.Logger, metrics *Metrics) *AuditLogger {
	return &AuditLogger{logger: logger, metrics: metrics}
}

func (a *AuditLogger) Log(ctx context.Context, event AuditEvent) {
	if a == nil {
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
	if event.Route != "" {
		attrs = append(attrs, Redact("route", normalizeRoute(event.Route)))
	}
	if event.Policy != "" {
		attrs = append(attrs, Redact("policy", normalizePolicy(event.Policy)))
	}
	if event.Backend != "" {
		attrs = append(attrs, Redact("backend", normalizeBackend(event.Backend)))
	}

	if a.logger != nil {
		a.logger.LogAttrs(ctx, slog.LevelInfo, "security audit", attrs...)
	}
	if a.metrics != nil {
		a.metrics.RecordSecurityEvent(event.Name, event.Outcome, event.Reason, event.Provider)
		if event.Name == EventAuthRateLimited || event.Name == EventRateLimiterUnavailable {
			a.metrics.RecordRateLimitEvent(event.Route, event.Policy, event.Backend, event.Outcome, event.Reason)
		}
	}
}
