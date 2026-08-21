package http

import (
	stdhttp "net/http"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/Jonathan0823/auth-go/internal/platform"
)

// OAuthFlow owns the HTTP-specific provider handshake outside the core service.
type OAuthFlow interface {
	BeginAuth(stdhttp.ResponseWriter, *stdhttp.Request, string)
	CompleteAuth(stdhttp.ResponseWriter, *stdhttp.Request, string) (domain.OAuthProfile, error)
}

type HandlerConfig struct {
	CookieDomain string
	SecureCookie bool
}

type Handler struct {
	Svc          port.Service
	Tokens       port.TokenService // used by middleware
	OAuth        OAuthFlow
	CookieDomain string
	SecureCookie bool
	Audit        *platform.AuditLogger
	RateLimiter  *platform.RateLimiter
}

func NewHandler(svc port.Service, tokens port.TokenService, oauth OAuthFlow, cfg HandlerConfig) *Handler {
	return &Handler{
		Svc:          svc,
		Tokens:       tokens,
		OAuth:        oauth,
		CookieDomain: cfg.CookieDomain,
		SecureCookie: cfg.SecureCookie,
	}
}
