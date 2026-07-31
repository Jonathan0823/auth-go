package http

import (
	"github.com/Jonathan0823/auth-go/internal/platform"
	"github.com/gin-gonic/gin"
)

func RegisterMetricsRoute(r *gin.Engine, metrics *platform.Metrics, enabled bool) {
	if !enabled || metrics == nil {
		return
	}
	r.GET("/metrics", gin.WrapH(metrics.Handler()))
}
