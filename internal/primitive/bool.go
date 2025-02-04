package primitive

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arjunsriva/promptgen/internal/handler"
)

// Bool handles boolean input/output
type Bool[O any] struct{}

// NewBool creates a new boolean handler
func NewBool[O any]() (handler.Handler[O], error) {
	return &Bool[O]{}, nil
}

func (h *Bool[O]) WrapPrompt(basePrompt string) string {
	return fmt.Sprintf(`%s

Provide your response as a single word: true or false.
Do not include any additional text or explanation.
Examples: true, false`, basePrompt)
}

func (h *Bool[O]) Parse(response string) (O, error) {
	var output O
	// Clean up the response
	cleaned := strings.ToLower(strings.TrimSpace(response))

	// Parse the boolean
	switch cleaned {
	case "true", "yes", "1":
		if v, ok := any(true).(O); ok {
			return v, nil
		}
	case "false", "no", "0":
		if v, ok := any(false).(O); ok {
			return v, nil
		}
	default:
		// Try standard parsing as fallback
		if b, err := strconv.ParseBool(cleaned); err == nil {
			if v, ok := any(b).(O); ok {
				return v, nil
			}
		}
	}

	return output, fmt.Errorf("failed to parse boolean from: %s", response)
}

func (h *Bool[O]) Validate(output O) error {
	// For booleans, we just ensure it's the right type
	if _, ok := any(output).(bool); !ok {
		return fmt.Errorf("expected boolean output, got %T", output)
	}
	return nil
}
