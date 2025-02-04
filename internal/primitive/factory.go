package primitive

import (
	"fmt"

	"github.com/arjunsriva/promptgen/internal/handler"
)

// New creates a new handler for primitive types
func New[O any]() (handler.Handler[O], error) {
	var zero O
	switch any(zero).(type) {
	case string:
		return NewString[O]()
	case int:
		return NewInt[O]()
	case float64:
		return NewFloat[O]()
	case bool:
		return NewBool[O]()
	default:
		return nil, fmt.Errorf("type %T is not a supported primitive type", zero)
	}
}
