// @title auth-go API
// @version 1.0
// @description Authentication and user management API with JWT access tokens and rotating refresh-token cookies.
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey CookieAuth
// @in header
// @name Cookie
// @description Send the access_token cookie for authenticated user endpoints.

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/Jonathan0823/auth-go/internal/bootstrap"
	"github.com/Jonathan0823/auth-go/internal/platform"
)

func main() {
	_ = godotenv.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := bootstrap.Run(ctx, platform.LoadConfig()); err != nil {
		log.Fatal(err)
	}
}
