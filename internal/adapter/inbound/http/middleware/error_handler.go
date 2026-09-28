package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		status, message := mapError(err)
		if status == http.StatusTooManyRequests || status == http.StatusServiceUnavailable {
			c.Header("Retry-After", rateLimitRetryAfter(err))
			c.Header("Cache-Control", "no-store")
		}
		if statusIsServerError(status) && logger != nil {
			logger.ErrorContext(c.Request.Context(), "application error",
				slog.String("request_id", RequestIDFromContext(c.Request.Context())),
				slog.Int("status", status),
				slog.String("error", message),
				slog.String("error_type", fmt.Sprintf("%T", err)),
			)
		}
		c.JSON(status, gin.H{"error": message})
		c.Abort()
	}
}
