// Package errors provides a set of error types and functions for handling errors in the application.
package errors

import (
	"net/http"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string {
	return e.Message
}

func New(code int, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func Unauthorized(message string, err error) *Error {
	return &Error{
		Code:    http.StatusUnauthorized,
		Message: message,
		Err:     err,
	}
}

func NotFound(message string, err error) *Error {
	return &Error{
		Code:    http.StatusNotFound,
		Message: message,
		Err:     err,
	}
}

func BadRequest(message string, err error) *Error {
	return &Error{
		Code:    http.StatusBadRequest,
		Message: message,
		Err:     err,
	}
}

func InternalServerError(message string, err error) *Error {
	return &Error{
		Code:    http.StatusInternalServerError,
		Message: message,
		Err:     err,
	}
}

func Conflict(message string, err error) *Error {
	return &Error{
		Code:    http.StatusConflict,
		Message: message,
		Err:     err,
	}
}

func Forbidden(message string, err error) *Error {
	return &Error{
		Code:    http.StatusForbidden,
		Message: message,
		Err:     err,
	}
}
