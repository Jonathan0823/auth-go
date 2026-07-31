package middleware

import (
	"strconv"
	"time"

	"github.com/Jonathan0823/auth-go/internal/platform"
	"github.com/gin-gonic/gin"
)

func Metrics(metrics *platform.Metrics) gin.HandlerFunc {
	if metrics == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		metrics.IncInFlight()
		started := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "/__unmatched__"
		}
		metrics.ObserveHTTPRequest(
			c.Request.Method,
			route,
			strconv.Itoa(c.Writer.Status()),
			time.Since(started),
		)
		metrics.DecInFlight()
	}
}
