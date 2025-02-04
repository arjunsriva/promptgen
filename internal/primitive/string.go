// Package primitive implements handlers for primitive types like strings
package primitive

import (
	"fmt"
	"strings"

	"github.com/arjunsriva/promptgen/internal/handler"
)

// String handles string input/output
type String[O any] struct{}

// NewString creates a new string handler
func NewString[O any]() (handler.Handler[O], error) {
	return &String[O]{}, nil
}

func (h *String[O]) WrapPrompt(basePrompt string) string {
	return fmt.Sprintf(`%s

Provide your response as plain text without any formatting or markers.
Keep it concise and to the point.`, basePrompt)
}

func (h *String[O]) Parse(response string) (O, error) {
	var output O
	// Remove any potential whitespace or newlines
	str := strings.TrimSpace(response)
	// Type assert the output
	if v, ok := any(str).(O); ok {
		return v, nil
	}
	return output, fmt.Errorf("failed to convert string to output type")
}

func (h *String[O]) Validate(output O) error {
	// For strings, we just ensure it's not empty
	str, ok := any(output).(string)
	if !ok {
		return fmt.Errorf("expected string output, got %T", output)
	}
	if str == "" {
		return fmt.Errorf("output string cannot be empty")
	}
	return nil
}
