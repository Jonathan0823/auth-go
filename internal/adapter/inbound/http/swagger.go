package http

import (
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Jonathan0823/auth-go/docs"
	"github.com/gin-gonic/gin"
)

func RegisterSwaggerRoutes(r *gin.Engine, enabled bool, environment string) {
	if !enabled || environment == "production" {
		return
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
