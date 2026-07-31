package bootstrap

import (
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
	logger := platform.NewLogger(cfg.LogLevel)
	metrics := platform.NewMetrics(pool)
	audit := platform.NewAuditLogger(logger, metrics)

	r.Use(inhttpmw.RequestID())
	if cfg.EnableMetrics {
		r.Use(inhttpmw.Metrics(metrics))
	}
	r.Use(inhttpmw.RequestLogger(logger))

	handler := inhttp.NewHandler(svc, tokens)
	handler.Audit = audit
	inhttp.RegisterRoutes(r, handler, logger)
	inhttp.RegisterSwaggerRoutes(r, cfg.EnableSwagger, cfg.Environment)
	inhttp.RegisterHealthRoutes(r, pool)
	inhttp.RegisterMetricsRoute(r, metrics, cfg.EnableMetrics)

	platform.InitServer(r, cfg)
}
