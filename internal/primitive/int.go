package primitive

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arjunsriva/promptgen/internal/handler"
)

// Int handles integer input/output
type Int[O any] struct{}

// NewInt creates a new integer handler
func NewInt[O any]() (handler.Handler[O], error) {
	return &Int[O]{}, nil
}

func (h *Int[O]) WrapPrompt(basePrompt string) string {
	return fmt.Sprintf(`%s

Provide your response as a single integer number.
Do not include any units, symbols, or additional text.
Examples: 42, -17, 0`, basePrompt)
}

func (h *Int[O]) Parse(response string) (O, error) {
	var output O
	// Clean up the response
	cleaned := strings.TrimSpace(response)

	// Parse the integer
	num, err := strconv.Atoi(cleaned)
	if err != nil {
		return output, fmt.Errorf("failed to parse integer: %v", err)
	}

	// Type assert to output type
	if v, ok := any(num).(O); ok {
		return v, nil
	}
	return output, fmt.Errorf("failed to convert integer to output type")
}

func (h *Int[O]) Validate(output O) error {
	// For integers, we just ensure it's the right type
	if _, ok := any(output).(int); !ok {
		return fmt.Errorf("expected integer output, got %T", output)
	}
	return nil
}
