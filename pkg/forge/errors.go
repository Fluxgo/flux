package flux

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound      = NewAppError("not found", http.StatusNotFound)
	ErrUnauthorized  = NewAppError("unauthorized", http.StatusUnauthorized)
	ErrForbidden     = NewAppError("forbidden", http.StatusForbidden)
	ErrBadRequest    = NewAppError("bad request", http.StatusBadRequest)
	ErrInternalError = NewAppError("internal server error", http.StatusInternalServerError)
	ErrValidation    = NewAppError("validation error", http.StatusBadRequest)
)

type AppError struct {
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Code       string                 `json:"code,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Err        error                  `json:"-"`
}

func NewAppError(message string, statusCode int) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) WithError(err error) *AppError {
	clone := *e
	clone.Err = err
	return &clone
}

func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	clone := *e
	clone.Details = details
	return &clone
}

func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	clone := *e
	if clone.Details == nil {
		clone.Details = make(map[string]interface{})
	}
	clone.Details[key] = value
	return &clone
}

func (e *AppError) WithCode(code string) *AppError {
	clone := *e
	clone.Code = code
	return &clone
}

func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return ErrInternalError.WithError(err)
}

func ValidationError(errors map[string]string) *AppError {
	details := make(map[string]interface{})
	for field, message := range errors {
		details[field] = message
	}
	return ErrValidation.WithDetails(details)
}

func NotFoundError(entity string) *AppError {
	err := ErrNotFound
	if entity != "" {
		err = NewAppError(fmt.Sprintf("%s not found", entity), http.StatusNotFound)
	}
	return err
}

func HandleError(ctx *Context, err error) error {
	appErr := AsAppError(err)
	return ctx.Status(appErr.StatusCode).JSON(appErr)
}
