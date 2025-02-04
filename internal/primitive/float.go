package primitive

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arjunsriva/promptgen/internal/handler"
)

// Float handles float64 input/output
type Float[O any] struct{}

// NewFloat creates a new float handler
func NewFloat[O any]() (handler.Handler[O], error) {
	return &Float[O]{}, nil
}

func (h *Float[O]) WrapPrompt(basePrompt string) string {
	return fmt.Sprintf(`%s

Provide your response as a single decimal number.
Use a period (.) as the decimal separator.
Do not include any units, symbols, or additional text.
Examples: 3.14, -2.5, 0.0, 42.0`, basePrompt)
}

func (h *Float[O]) Parse(response string) (O, error) {
	var output O
	// Clean up the response
	cleaned := strings.TrimSpace(response)

	// Parse the float
	num, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return output, fmt.Errorf("failed to parse float: %v", err)
	}

	// Type assert to output type
	if v, ok := any(num).(O); ok {
		return v, nil
	}
	return output, fmt.Errorf("failed to convert float to output type")
}

func (h *Float[O]) Validate(output O) error {
	// For floats, we just ensure it's the right type
	if _, ok := any(output).(float64); !ok {
		return fmt.Errorf("expected float64 output, got %T", output)
	}
	return nil
}
