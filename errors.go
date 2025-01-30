package promptgen

import "errors"

// Common error types for the promptgen package
var (
	ErrNoProvider      = errors.New("no provider configured")
	ErrInvalidResponse = errors.New("invalid response from provider")
	ErrRateLimit       = errors.New("rate limit exceeded")
	ErrContextLength   = errors.New("context length exceeded")
)

// IsRateLimit checks if the error is a rate limit error
func IsRateLimit(err error) bool {
	return errors.Is(err, ErrRateLimit)
}

// IsContextLength checks if the error is a context length error
func IsContextLength(err error) bool {
	return errors.Is(err, ErrContextLength)
}
