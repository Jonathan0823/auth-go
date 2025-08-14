// Package errors provides a set of error types and functions for handling errors in the application.
package errors

import "errors"

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)
