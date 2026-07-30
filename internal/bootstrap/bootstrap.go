package bootstrap

import (
	"github.com/gin-gonic/gin"

	inhttp "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http"
	inhttpmw "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/middleware"
	outemail "github.com/Jonathan0823/auth-go/internal/adapter/outbound/email"
	outjwt "github.com/Jonathan0823/auth-go/internal/adapter/outbound/jwt"
	outpassword "github.com/Jonathan0823/auth-go/internal/adapter/outbound/password"
	outpostgres "github.com/Jonathan0823/auth-go/internal/adapter/outbound/postgres"
	"github.com/Jonathan0823/auth-go/internal/core/service"
	"github.com/Jonathan0823/auth-go/internal/platform"
)

func Run(cfg platform.Config) {
	platform.InitOAuth()

	pool := platform.NewPool()
	defer pool.Close()

	repo := outpostgres.NewRepository(pool)
	tokens := outjwt.NewTokenService()
	email := outemail.NewSender()
	hasher := outpassword.NewHasher()
	svc := service.New(repo, tokens, email, hasher, cfg.BaseURL)

	r := gin.New()
	logger := platform.NewLogger(cfg.LogLevel)
	r.Use(inhttpmw.RequestID(), inhttpmw.RequestLogger(logger))

	handler := inhttp.NewHandler(svc, tokens)
	inhttp.RegisterRoutes(r, handler, logger)

	platform.InitServer(r, cfg)
}
