package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	inhttp "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http"
	inhttpmw "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/middleware"
	outemail "github.com/Jonathan0823/auth-go/internal/adapter/outbound/email"
	outjwt "github.com/Jonathan0823/auth-go/internal/adapter/outbound/jwt"
	outoauth "github.com/Jonathan0823/auth-go/internal/adapter/outbound/oauth"
	outpassword "github.com/Jonathan0823/auth-go/internal/adapter/outbound/password"
	outpostgres "github.com/Jonathan0823/auth-go/internal/adapter/outbound/postgres"
	"github.com/Jonathan0823/auth-go/internal/config"
	"github.com/Jonathan0823/auth-go/internal/core/ratelimit"
	"github.com/Jonathan0823/auth-go/internal/core/service"
	"github.com/Jonathan0823/auth-go/internal/observability"
)

func Run(ctx context.Context, cfg config.Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate configuration: %w", err)
	}

	startupCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	pool, err := outpostgres.NewPool(startupCtx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := outpostgres.NewRepository(pool)
	tokens := outjwt.NewTokenService(cfg.JWTAccessSecret, cfg.RefreshTokenHashKey)
	email := outemail.NewSender(cfg.EmailAddress, cfg.EmailPassword)
	hasher := outpassword.NewHasher()
	oauth := outoauth.New(outoauth.Config{
		BaseURL:            cfg.BaseURL,
		SessionSecret:      cfg.SessionSecret,
		GitHubClientID:     cfg.GitHubClientID,
		GitHubClientSecret: cfg.GitHubClientSecret,
		GoogleClientID:     cfg.GoogleClientID,
		GoogleClientSecret: cfg.GoogleClientSecret,
		SecureCookies:      cfg.Environment == "production",
	})

	r := gin.New()
	if err := r.SetTrustedProxies(cfg.RateLimit.TrustedProxies); err != nil {
		return fmt.Errorf("configure trusted proxies: %w", err)
	}
	logger := observability.NewLogger(cfg.LogLevel)
	metrics := observability.NewMetrics(pool)
	audit := observability.NewAuditLogger(logger, metrics)
	rateLimitStore, redisClient, err := newRateLimitStore(cfg.RateLimit, pool)
	if err != nil {
		return fmt.Errorf("create rate-limit backend: %w", err)
	}
	defer func() { _ = rateLimitStore.Close() }()
	if redisClient != nil {
		defer func() { _ = redisClient.Close() }()
	}
	rateLimiter := ratelimit.NewRateLimiter(rateLimitStore, cfg.RateLimit.Key, cfg.RateLimit.Policies)
	svc := service.New(repo, tokens, email, hasher, cfg.BaseURL, rateLimiter)

	r.Use(inhttp.CORS(cfg))
	r.Use(inhttpmw.RequestID())
	if cfg.EnableMetrics {
		r.Use(inhttpmw.Metrics(metrics))
	}
	r.Use(inhttpmw.RequestLogger(logger))

	handler := inhttp.NewHandler(svc, tokens, oauth, inhttp.HandlerConfig{
		CookieDomain: cfg.Domain,
		SecureCookie: cfg.Environment == "production",
	})
	handler.Audit = audit
	handler.RateLimiter = rateLimiter
	inhttp.RegisterRoutes(r, handler, logger)
	inhttp.RegisterSwaggerRoutes(r, cfg.EnableSwagger, cfg.Environment)
	inhttp.RegisterHealthRoutes(r, pool)
	inhttp.RegisterMetricsRoute(r, metrics, cfg.EnableMetrics)

	return inhttp.RunServer(ctx, r, cfg)
}
