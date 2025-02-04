package promptgen

import (
	"errors"
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	err := &Error{
		Err:     ErrValidation,
		Code:    "test_code",
		Message: "test message",
		Details: map[string]interface{}{
			"key": "value",
		},
	}

	// Test Error() string method
	want := "test_code: test message"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}

	// Test Unwrap() method
	if got := err.Unwrap(); !errors.Is(got, ErrValidation) {
		t.Errorf("Unwrap() = %v, want %v", got, ErrValidation)
	}
}

func TestErrorPredicates(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		check    func(error) bool
		expected bool
	}{
		{
			name:     "IsValidation with validation error",
			err:      &Error{Err: ErrValidation},
			check:    IsValidation,
			expected: true,
		},
		{
			name:     "IsValidation with other error",
			err:      &Error{Err: ErrTimeout},
			check:    IsValidation,
			expected: false,
		},
		{
			name:     "IsTimeout with timeout error",
			err:      &Error{Err: ErrTimeout},
			check:    IsTimeout,
			expected: true,
		},
		{
			name:     "IsTimeout with other error",
			err:      &Error{Err: ErrValidation},
			check:    IsTimeout,
			expected: false,
		},
		{
			name:     "IsRateLimit with rate limit error",
			err:      &Error{Err: ErrRateLimit},
			check:    IsRateLimit,
			expected: true,
		},
		{
			name:     "IsRateLimit with other error",
			err:      &Error{Err: ErrTimeout},
			check:    IsRateLimit,
			expected: false,
		},
		{
			name:     "IsContextLength with context length error",
			err:      &Error{Err: ErrContextLength},
			check:    IsContextLength,
			expected: true,
		},
		{
			name:     "IsContextLength with other error",
			err:      &Error{Err: ErrTimeout},
			check:    IsContextLength,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			check:    IsValidation,
			expected: false,
		},
		{
			name:     "wrapped error",
			err:      fmt.Errorf("wrapped: %w", &Error{Err: ErrValidation}),
			check:    IsValidation,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.check(tt.err); got != tt.expected {
				t.Errorf("%s(%v) = %v, want %v",
					getFunctionName(tt.check), tt.err, got, tt.expected)
			}
		})
	}
}

// Helper function to get the name of the error check function
func getFunctionName(f func(error) bool) string {
	// Create a sample error to test against
	testErr := &Error{Err: ErrValidation}

	switch {
	case f(testErr) == IsValidation(testErr):
		return "IsValidation"
	case f(testErr) == IsTimeout(testErr):
		return "IsTimeout"
	case f(testErr) == IsRateLimit(testErr):
		return "IsRateLimit"
	case f(testErr) == IsContextLength(testErr):
		return "IsContextLength"
	default:
		return "unknown"
	}
}
