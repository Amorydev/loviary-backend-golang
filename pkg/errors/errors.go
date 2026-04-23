package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Error types
var (
	ErrNotFound = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden = errors.New("forbidden")
	ErrConflict = errors.New("conflict")
	ErrInternal = errors.New("internal server error")
)

// AppError represents an application error with context
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithCause adds the underlying error
func (e *AppError) WithCause(err error) *AppError {
	e.Err = err
	return e
}

// New creates a new AppError
func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewWith creates a new AppError with underlying error
func NewWith(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// HTTPStatus returns the HTTP status code for the error
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case "NOT_FOUND", "USER_NOT_FOUND", "COUPLE_NOT_FOUND", "MOOD_NOT_FOUND":
		return http.StatusNotFound
	case "INVALID_INPUT", "VALIDATION_ERROR":
		return http.StatusBadRequest
	case "UNAUTHORIZED", "INVALID_TOKEN", "EXPIRED_TOKEN":
		return http.StatusUnauthorized
	case "FORBIDDEN", "NO_ACTIVE_COUPLE":
		return http.StatusForbidden
	case "CONFLICT", "USER_EXISTS", "EMAIL_EXISTS", "INVITATION_EXPIRED":
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// Predefined errors
var (
	UserNotFound = New("USER_NOT_FOUND", "User not found")
	EmailExists = New("EMAIL_EXISTS", "Email already registered")
	UsernameExists = New("USERNAME_EXISTS", "Username already taken")
	InvalidCredentials = New("INVALID_CREDENTIALS", "Invalid email or password")
	ExpiredToken = New("EXPIRED_TOKEN", "Token has expired")
	InvalidToken = New("INVALID_TOKEN", "Invalid token")
	NoActiveCouple = New("NO_ACTIVE_COUPLE", "No active couple relationship")
	InvitationExpired = New("INVITATION_EXPIRED", "Invitation has expired")
	MoodNotFound = New("MOOD_NOT_FOUND", "Mood not found")
	StreakNotFound = New("STREAK_NOT_FOUND", "Streak not found")
	ReminderNotFound = New("REMINDER_NOT_FOUND", "Reminder not found")
	MemoryNotFound = New("MEMORY_NOT_FOUND", "Memory not found")
)
