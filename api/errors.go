package api

import (
	"errors"
	"net/http"
)

type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Details    any
	Err        error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	if e.Err != nil {
		return e.Err.Error()
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func (e *AppError) WithCause(err error) *AppError {
	e.Err = err
	return e
}

func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

func NewError(statusCode int, code, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

func BadRequest(code, message string) *AppError {
	return NewError(http.StatusBadRequest, code, message)
}

func Unauthorized(code, message string) *AppError {
	return NewError(http.StatusUnauthorized, code, message)
}

func Forbidden(code, message string) *AppError {
	return NewError(http.StatusForbidden, code, message)
}

func Conflict(code, message string) *AppError {
	return NewError(http.StatusConflict, code, message)
}

func TooManyRequests(code, message string) *AppError {
	return NewError(http.StatusTooManyRequests, code, message)
}

func Internal(code, message string) *AppError {
	return NewError(http.StatusInternalServerError, code, message)
}

func NotFound(code, message string) *AppError {
	return NewError(http.StatusNotFound, code, message)
}

func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}

	return nil, false
}

func NormalizeError(err error) *AppError {
	if err == nil {
		return nil
	}

	if appErr, ok := AsAppError(err); ok {
		return appErr
	}

	return Internal("INTERNAL_SERVER_ERROR", "Something went wrong").WithCause(err)
}
