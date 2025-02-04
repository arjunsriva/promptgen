package json

import (
	"testing"
)

type testOutput struct {
	Response string `json:"response" jsonschema:"required,maxLength=100"`
}

func TestSchemaValidation(t *testing.T) {
	validator, err := NewValidator[testOutput]()
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// Test schema generation
	schema, err := validator.SchemaString()
	if err != nil {
		t.Fatalf("SchemaString failed: %v", err)
	}
	t.Logf("Generated Schema:\n%s", schema)

	// Test response validation
	validJSON := `{"response": "I'm doing well, thank you for asking!"}`
	if err := validator.Validate([]byte(validJSON)); err != nil {
		t.Errorf("Validate failed for valid JSON: %v", err)
	}

	invalidJSON := `{"response": "This response is way too long and exceeds the maximum length of 100 characters that we specified in the JSON Schema validation rules"}`
	if err := validator.Validate([]byte(invalidJSON)); err == nil {
		t.Error("Validate should fail for invalid JSON")
	} else {
		t.Logf("Expected error for invalid JSON: %v", err)
	}

	// Test missing required field
	emptyJSON := `{}`
	if err := validator.Validate([]byte(emptyJSON)); err == nil {
		t.Error("Validate should fail for missing required field")
	} else {
		t.Logf("Expected error for missing field: %v", err)
	}
}

func TestValidatorErrors(t *testing.T) {
	type InvalidType struct {
		Channel chan int // channels can't be converted to JSON Schema
	}

	_, err := NewValidator[InvalidType]()
	if err != nil {
		t.Logf("Expected error for invalid type: %v", err)
	} else {
		t.Error("expected error for invalid type, got nil")
	}
}
