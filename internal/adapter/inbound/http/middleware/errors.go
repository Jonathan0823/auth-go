package middleware

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/service"
)

func mapError(err error) (int, string) {
	var limitErr *service.RateLimitError
	if errors.As(err, &limitErr) {
		if limitErr.Unavailable {
			return http.StatusServiceUnavailable, "authentication temporarily unavailable"
		}
		return http.StatusTooManyRequests, "too many requests"
	}
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, "invalid input"
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "conflict"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func rateLimitRetryAfter(err error) string {
	var limitErr *service.RateLimitError
	if !errors.As(err, &limitErr) {
		return RetryAfterHeader(0)
	}
	duration := limitErr.RetryAfter
	if limitErr.Unavailable {
		duration = 30 * time.Second
	}
	return RetryAfterHeader(duration)
}

func RetryAfterHeader(duration time.Duration) string {
	seconds := int64(math.Ceil(duration.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	if seconds > 3600 {
		seconds = 3600
	}
	return strconv.FormatInt(seconds, 10)
}
