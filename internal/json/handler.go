// Package json implements handlers for JSON/struct types
package json

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/arjunsriva/promptgen/internal/handler"
)

// Handler handles JSON/struct input/output
type Handler[O any] struct {
	validator *Validator[O]
}

// New creates a new JSON handler
func New[O any]() (handler.Handler[O], error) {
	validator, err := NewValidator[O]()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}
	return &Handler[O]{
		validator: validator,
	}, nil
}

func (h *Handler[O]) WrapPrompt(basePrompt string) string {
	schema, _ := h.validator.SchemaString()
	return fmt.Sprintf(`%s

Format your response according to this JSON schema, pay close attention to the validation rules in the schema:
%s

Provide the result enclosed in triple backticks with 'json' on the first line.
Don't put control characters in the wrong place or the JSON will be invalid.`, basePrompt, schema)
}

func (h *Handler[O]) Parse(response string) (O, error) {
	var output O

	// Extract JSON from markdown code blocks if present
	jsonStr := response
	if matches := jsonRegex.FindStringSubmatch(response); len(matches) >= 2 {
		jsonStr = matches[1]
	}

	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		return output, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return output, nil
}

func (h *Handler[O]) Validate(output O) error {
	// Convert to JSON for validation
	jsonBytes, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	return h.validator.Validate(jsonBytes)
}

// Regex for extracting JSON from markdown code blocks
var jsonRegex = regexp.MustCompile("```(?:json)?([\\s\\S]*?)```")
