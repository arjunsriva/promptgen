// Package handler defines the core interface for processing different output types
package handler

import "fmt"

// Handler defines how different output types are processed
type Handler[O any] interface {
	// WrapPrompt adds type-specific instructions to the base prompt
	WrapPrompt(basePrompt string) string

	// Parse converts the AI response into the target type
	Parse(response string) (O, error)

	// Validate checks if the output meets requirements
	Validate(O) error
}

// Type represents the kind of handler needed
type Type int

const (
	TypeUnknown Type = iota
	TypeString
	TypeJSON
	TypePrimitive
)

// String returns a string representation of the Type
func (t Type) String() string {
	switch t {
	case TypeUnknown:
		return "unknown"
	case TypeString:
		return "string"
	case TypeJSON:
		return "json"
	case TypePrimitive:
		return "primitive"
	default:
		return fmt.Sprintf("unknown type %d", t)
	}
}

// DetermineType returns the appropriate handler type for a given type O
func DetermineType[O any]() Type {
	var zero O
	switch any(zero).(type) {
	case string:
		return TypeString
	case int, float64, bool:
		return TypePrimitive
	default:
		return TypeJSON
	}
}
