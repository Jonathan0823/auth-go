package bootstrap

import (
	"github.com/gin-gonic/gin"

	inhttp "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http"
	outemail "github.com/Jonathan0823/auth-go/internal/adapter/outbound/email"
	outjwt "github.com/Jonathan0823/auth-go/internal/adapter/outbound/jwt"
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
	svc := service.New(repo, tokens, email)

	r := gin.New()
	r.Use(gin.Logger())

	handler := inhttp.NewHandler(svc)
	inhttp.RegisterRoutes(r, handler)

	platform.InitServer(r, cfg)
}
