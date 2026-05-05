package errors

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, defaultMsg string, msg ...string) *AppError {
	message := defaultMsg
	if len(msg) > 0 && msg[0] != "" {
		message = msg[0]
	}
	return &AppError{Code: code, Message: message}
}

func BadRequest(msg ...string) *AppError {
	return New(http.StatusBadRequest, "Bad request", msg...)
}

func Unauthorized(msg ...string) *AppError {
	return New(http.StatusUnauthorized, "Unauthorized", msg...)
}

func Forbidden(msg ...string) *AppError {
	return New(http.StatusForbidden, "Forbidden", msg...)
}

func NotFound(msg ...string) *AppError {
	return New(http.StatusNotFound, "Resource not found", msg...)
}

func Conflict(msg ...string) *AppError {
	return New(http.StatusConflict, "Conflict occurred", msg...)
}

func Internal(msg ...string) *AppError {
	return New(http.StatusInternalServerError, "Internal server error", msg...)
}

// Sentinel errors for repo-to-service error mapping
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// MapRepoError maps ORM/database errors to sentinel errors.
// Extend with your ORM-specific error checks (GORM, SQLx, etc.).
// func MapRepoError(err error) error {
// 	if err == nil {
// 		return nil
// 	}
// 	if errors.Is(err, gorm.ErrRecordNotFound) {
// 		return ErrNotFound
// 	}
// 	return err
// }
