package domain

const (
	ErrCodeInternal     = 1
	ErrCodeBadRequest   = 2
	ErrCodeNotFound     = 3
	ErrCodeConflict     = 4
	ErrCodeUnauthorized = 5
	ErrCodeForbidden    = 6
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string { return e.Message }

func NewError(code int, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

func Unauthorized(message string, err error) *Error {
	return &Error{Code: ErrCodeUnauthorized, Message: message, Err: err}
}

func NotFound(message string, err error) *Error {
	return &Error{Code: ErrCodeNotFound, Message: message, Err: err}
}

func BadRequest(message string, err error) *Error {
	return &Error{Code: ErrCodeBadRequest, Message: message, Err: err}
}

func InternalServerError(message string, err error) *Error {
	return &Error{Code: ErrCodeInternal, Message: message, Err: err}
}

func Conflict(message string, err error) *Error {
	return &Error{Code: ErrCodeConflict, Message: message, Err: err}
}

func Forbidden(message string, err error) *Error {
	return &Error{Code: ErrCodeForbidden, Message: message, Err: err}
}

func ErrorCode(err error) int {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return ErrCodeInternal
}

func ErrorMessage(err error) string {
	if e, ok := err.(*Error); ok {
		return e.Message
	}
	return "internal server error"
}
