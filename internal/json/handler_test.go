package json

import (
	"testing"
)

type handlerTestOutput struct {
	Response string `json:"response" jsonschema:"required,minLength=1"`
}

func TestJSONHandler(t *testing.T) {
	handler, err := New[handlerTestOutput]()
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	t.Run("wrap prompt", func(t *testing.T) {
		prompt := handler.WrapPrompt("Generate a message")
		if prompt == "Generate a message" {
			t.Error("prompt should be wrapped with schema and instructions")
		}
	})

	t.Run("parse valid JSON with markdown", func(t *testing.T) {
		input := "```json\n{\"response\":\"hello\"}\n```"
		result, err := handler.Parse(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.Response != "hello" {
			t.Errorf("expected response 'hello', got %q", result.Response)
		}
	})

	t.Run("parse invalid JSON", func(t *testing.T) {
		input := "```json\n{\"response\":noll}\n```"
		if _, err := handler.Parse(input); err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("parse missing markdown blocks", func(t *testing.T) {
		input := "{\"response\":\"hello\"}"
		result, err := handler.Parse(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.Response != "hello" {
			t.Errorf("expected response 'hello', got %q", result.Response)
		}
	})

	t.Run("validate valid output", func(t *testing.T) {
		output := handlerTestOutput{Response: "hello"}
		if err := handler.Validate(output); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("validate invalid output", func(t *testing.T) {
		output := handlerTestOutput{Response: ""}
		if err := handler.Validate(output); err == nil {
			t.Error("expected error for empty response")
		} else {
			t.Logf("Got error: %v", err)
		}
	})
}

func TestNewJSONHandler(t *testing.T) {
	t.Run("invalid type", func(t *testing.T) {
		type InvalidType struct {
			Channel chan int // channels can't be converted to JSON Schema
		}
		_, err := New[InvalidType]()
		if err == nil {
			t.Error("expected error for invalid type")
		}
	})

	t.Run("valid type", func(t *testing.T) {
		handler, err := New[handlerTestOutput]()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if handler == nil {
			t.Error("handler should not be nil")
		}
	})
}
