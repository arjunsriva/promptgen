package json

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
		DoNotReference:             true,
		ExpandedStruct:             true,
		RequiredFromJSONSchemaTags: true,
		AllowAdditionalProperties:  false,
	}

	// Try to generate schema to validate the type
	var schemaErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				schemaErr = fmt.Errorf("type %T contains unsupported type: %v", t, r)
			}
		}()
		r.ReflectFromType(typ)
	}()

	if schemaErr != nil {
		return nil, schemaErr
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
	schema.ID = ""
	schema.Version = ""
	schema.Definitions = nil

	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	return string(jsonBytes), nil
}

// Validate checks if the given JSON data matches the schema
func (v *Validator[T]) Validate(data []byte) error {
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
			msg := err.String()
			msg = strings.TrimPrefix(msg, "(root).")
			errors = append(errors, msg)
		}
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}
