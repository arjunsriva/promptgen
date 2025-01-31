package promptgen

import (
	"errors"
	"fmt"
)

// Common error types for the promptgen package
var (
	ErrNoProvider      = errors.New("no provider configured")
	ErrInvalidResponse = errors.New("invalid response from provider")
	ErrValidation      = errors.New("validation failed")
	ErrConfiguration   = errors.New("invalid configuration")
	ErrTimeout         = errors.New("request timeout")
	ErrRateLimit       = errors.New("rate limit exceeded")
	ErrContextLength   = errors.New("context length exceeded")
)

// Error wraps provider errors with additional context
type Error struct {
	Err     error
	Code    string
	Message string
	Details map[string]interface{}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// IsValidation checks if the error is a validation error
func IsValidation(err error) bool {
	return errors.Is(err, ErrValidation)
}

// IsTimeout checks if the error is a timeout error
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsRateLimit checks if the error is a rate limit error
func IsRateLimit(err error) bool {
	return errors.Is(err, ErrRateLimit)
}

// IsContextLength checks if the error is a context length error
func IsContextLength(err error) bool {
	return errors.Is(err, ErrContextLength)
}
