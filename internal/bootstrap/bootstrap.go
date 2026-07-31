package bootstrap

import (
	"log"

	"github.com/gin-gonic/gin"

	inhttp "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http"
	inhttpmw "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/middleware"
	outemail "github.com/Jonathan0823/auth-go/internal/adapter/outbound/email"
	outjwt "github.com/Jonathan0823/auth-go/internal/adapter/outbound/jwt"
	outoauth "github.com/Jonathan0823/auth-go/internal/adapter/outbound/oauth"
	outpassword "github.com/Jonathan0823/auth-go/internal/adapter/outbound/password"
	outpostgres "github.com/Jonathan0823/auth-go/internal/adapter/outbound/postgres"
	"github.com/Jonathan0823/auth-go/internal/core/service"
	"github.com/Jonathan0823/auth-go/internal/platform"
)

func Run(cfg platform.Config) {
	if err := cfg.RateLimit.Validate(cfg.Environment); err != nil {
		log.Fatal(err)
	}
	pool := platform.NewPool()
	defer pool.Close()

	repo := outpostgres.NewRepository(pool)
	tokens := outjwt.NewTokenService()
	email := outemail.NewSender()
	hasher := outpassword.NewHasher()
	oauth := outoauth.New(outoauth.Config{
		BaseURL:            cfg.BaseURL,
		SessionSecret:      cfg.SessionSecret,
		GitHubClientID:     cfg.GitHubClientID,
		GitHubClientSecret: cfg.GitHubClientSecret,
		GoogleClientID:     cfg.GoogleClientID,
		GoogleClientSecret: cfg.GoogleClientSecret,
	})
	svc := service.New(repo, tokens, email, hasher, cfg.BaseURL, oauth)

	r := gin.New()
	if err := r.SetTrustedProxies(cfg.RateLimit.TrustedProxies); err != nil {
		log.Fatal("invalid trusted proxies configuration")
	}
	logger := platform.NewLogger(cfg.LogLevel)
	metrics := platform.NewMetrics(pool)
	audit := platform.NewAuditLogger(logger, metrics)
	rateLimitStore, redisClient, err := newRateLimitStore(cfg.RateLimit, pool)
	if err != nil {
		log.Fatal("rate-limit backend is unavailable")
	}
	defer rateLimitStore.Close()
	if redisClient != nil {
		defer redisClient.Close()
	}

	r.Use(inhttpmw.RequestID())
	if cfg.EnableMetrics {
		r.Use(inhttpmw.Metrics(metrics))
	}
	r.Use(inhttpmw.RequestLogger(logger))

	handler := inhttp.NewHandler(svc, tokens)
	handler.Audit = audit
	handler.RateLimiter = platform.NewRateLimiter(rateLimitStore, cfg.RateLimit.Key, cfg.RateLimit.Policies)
	inhttp.RegisterRoutes(r, handler, logger)
	inhttp.RegisterSwaggerRoutes(r, cfg.EnableSwagger, cfg.Environment)
	inhttp.RegisterHealthRoutes(r, pool)
	inhttp.RegisterMetricsRoute(r, metrics, cfg.EnableMetrics)

	platform.InitServer(r, cfg)
}
