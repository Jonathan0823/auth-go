// Package errors provides a set of error types and functions for handling errors in the application.
package errors

import (
	"net/http"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

func New(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func Unauthorized(message string) *Error {
	return &Error{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

func NotFound(message string) *Error {
	return &Error{
		Code:    http.StatusNotFound,
		Message: message,
	}
}

func BadRequest(message string) *Error {
	return &Error{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

func InternalServerError(message string) *Error {
	return &Error{
		Code:    http.StatusInternalServerError,
		Message: message,
	}
}

func Conflict(message string) *Error {
	return &Error{
		Code:    http.StatusConflict,
		Message: message,
	}
}
