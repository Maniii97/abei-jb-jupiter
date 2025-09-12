package errors

import (
	"errors"
	"fmt"
)

// Custom error types
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInternalServer     = errors.New("internal server error")
	ErrBadRequest         = errors.New("bad request")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrRecordNotFound     = errors.New("record not found")
)

// AppError represents an application error with additional context
type AppError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Error constructors
func NewBadRequestError(message string, cause error) *AppError {
	return &AppError{
		Type:    "BAD_REQUEST",
		Message: message,
		Cause:   cause,
	}
}

func NewUnauthorizedError(message string, cause error) *AppError {
	return &AppError{
		Type:    "UNAUTHORIZED",
		Message: message,
		Cause:   cause,
	}
}

func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Type:    "INTERNAL_ERROR",
		Message: message,
		Cause:   cause,
	}
}

func NewNotFoundError(message string, cause error) *AppError {
	return &AppError{
		Type:    "NOT_FOUND",
		Message: message,
		Cause:   cause,
	}
}

func NewConflictError(message string, cause error) *AppError {
	return &AppError{
		Type:    "CONFLICT",
		Message: message,
		Cause:   cause,
	}
}
