package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Jonathan0823/auth-go/internal/platform"
)

type requestIDKey struct{}

const requestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if _, err := uuid.Parse(requestID); err != nil {
			requestID = uuid.NewString()
		}

		ctx := context.WithValue(c.Request.Context(), requestIDKey{}, requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Set(requestIDHeader, requestID)
		c.Header(requestIDHeader, requestID)
		c.Next()
	}
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger.InfoContext(c.Request.Context(), "http request",
			slog.String("request_id", RequestIDFromContext(c.Request.Context())),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(start)),
			platform.Redact("user_agent", c.GetHeader("User-Agent")),
		)
	}
}

func statusIsServerError(status int) bool {
	return status >= http.StatusInternalServerError
}
