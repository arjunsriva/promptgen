package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

// Validator handles JSON Schema validation for struct types
type Validator[T any] struct {
	reflector *jsonschema.Reflector
	typeOf    reflect.Type
}

// NewValidator creates a validator from a struct type
func NewValidator[T any]() (*Validator[T], error) {
	var t T
	typ := reflect.TypeOf(t)

	r := &jsonschema.Reflector{
		DoNotReference:             true,  // Prevents $ref usage
		ExpandedStruct:             true,  // Includes all fields
		AllowAdditionalProperties:  false, // Strict object validation
		RequiredFromJSONSchemaTags: true,  // Use jsonschema:"required" tag
	}

	return &Validator[T]{
		reflector: r,
		typeOf:    typ,
	}, nil
}

// SchemaString returns the JSON Schema as a string
func (v *Validator[T]) SchemaString() (string, error) {
	schema := v.reflector.ReflectFromType(v.typeOf)

	// Clean up the schema
	schema.ID = ""           // Remove $id
	schema.Version = ""      // Remove $schema
	schema.Definitions = nil // Remove $defs

	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	return string(jsonBytes), nil
}

// Validate checks if the given JSON data matches the schema
func (v *Validator[T]) Validate(data []byte) error {
	// Get the schema
	schema := v.reflector.ReflectFromType(v.typeOf)

	// Convert schema to JSON
	schemaData, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Create schema loader
	schemaLoader := gojsonschema.NewStringLoader(string(schemaData))
	documentLoader := gojsonschema.NewStringLoader(string(data))

	// Create validator
	validator, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Validate
	result, err := validator.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, err := range result.Errors() {
			errors = append(errors, err.String())
		}
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}
