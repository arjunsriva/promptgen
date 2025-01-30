package promptgen

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/arjunsriva/promptgen/provider"
)

// Test types
type TestInput struct {
	Name    string
	Message string
}

type TestOutput struct {
	Response string `json:"response" jsonschema:"required,maxLength=100"`
}

// Core functionality tests
func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name:        "valid template",
			template:    "Hello {{.Name}}, {{.Message}}",
			expectError: false,
		},
		{
			name:        "invalid template",
			template:    "Hello {{.Name}, {{.Message}}", // Missing closing brace
			expectError: true,
			errorCheck: func(err error) bool {
				return strings.Contains(err.Error(), "invalid template")
			},
		},
		{
			name:        "empty template",
			template:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := Create[TestInput, TestOutput](tt.template)

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, got nil")
					return
				}
				if tt.errorCheck != nil && !tt.errorCheck(err) {
					t.Errorf("error check failed for error: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if generator == nil {
				t.Error("expected generator to not be nil")
				return
			}
		})
	}
}

func TestGeneratorConfiguration(t *testing.T) {
	generator, err := Create[TestInput, TestOutput]("test template")
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	// Test WithModel
	model := "gpt-4"
	generator.WithModel(model)
	if generator.model != model {
		t.Errorf("expected model %q, got %q", model, generator.model)
	}

	// Test WithTemperature
	temp := 0.5
	generator.WithTemperature(temp)
	if generator.temp != temp {
		t.Errorf("expected temperature %v, got %v", temp, generator.temp)
	}

	// Test WithMaxTokens
	tokens := 500
	generator.WithMaxTokens(tokens)
	if generator.maxTokens != tokens {
		t.Errorf("expected maxTokens %v, got %v", tokens, generator.maxTokens)
	}
}

func TestTemplateExecution(t *testing.T) {
	template := "Hello {{.Name}}, {{.Message}}"
	generator, err := Create[TestInput, TestOutput](template)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	input := TestInput{
		Name:    "Alice",
		Message: "how are you?",
	}

	var buf strings.Builder
	err = generator.prompt.Execute(&buf, input)
	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	expected := "Hello Alice, how are you?"
	if got := buf.String(); got != expected {
		t.Errorf("template execution\nwant: %q\ngot:  %q", expected, got)
	}
}

// Schema validation tests
func TestSchemaValidation(t *testing.T) {
	gen, err := Create[TestInput, TestOutput]("Hello {{.Name}}, {{.Message}}")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test schema generation
	schema, err := gen.SchemaString()
	if err != nil {
		t.Fatalf("SchemaString failed: %v", err)
	}
	t.Logf("Generated Schema:\n%s", schema)

	// Test response validation
	validJSON := `{"response": "I'm doing well, thank you for asking!"}`
	if err := gen.ValidateResponse([]byte(validJSON)); err != nil {
		t.Errorf("ValidateResponse failed for valid JSON: %v", err)
	}

	invalidJSON := `{"response": "This response is way too long and exceeds the maximum length of 100 characters that we specified in the JSON Schema validation rules"}`
	if err := gen.ValidateResponse([]byte(invalidJSON)); err == nil {
		t.Error("ValidateResponse should fail for invalid JSON")
	}
}

// Provider integration tests
func TestGeneratorWithProvider(t *testing.T) {
	type ChatInput struct {
		Message string
	}

	type ChatOutput struct {
		Response string `json:"response" jsonschema:"required,maxLength=100"`
	}

	tests := []struct {
		name        string
		input       ChatInput
		mockResp    string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid response",
			input: ChatInput{
				Message: "Hello!",
			},
			mockResp: `{"response": "Hi there! How can I help you today?"}`,
			wantErr:  false,
		},
		{
			name: "invalid json response",
			input: ChatInput{
				Message: "Hello!",
			},
			mockResp:    `not a json response`,
			wantErr:     true,
			errContains: "invalid response",
		},
		{
			name: "response too long",
			input: ChatInput{
				Message: "Hello!",
			},
			mockResp:    `{"response": "This response is way too long and exceeds the maximum length of 100 characters that we specified in the JSON Schema validation rules"}`,
			wantErr:     true,
			errContains: "validation errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := Create[ChatInput, ChatOutput]("User: {{.Message}}")
			if err != nil {
				t.Fatalf("Create failed: %v", err)
			}

			mock := &provider.MockProvider{
				Response: tt.mockResp,
			}

			gen.WithProvider(mock).
				WithModel("test-model").
				WithTemperature(0.7).
				WithMaxTokens(100)

			output, err := gen.Run(context.Background(), tt.input)

			// Verify provider was called with correct request
			if len(mock.Requests) != 1 {
				t.Errorf("expected 1 request, got %d", len(mock.Requests))
			}

			req := mock.Requests[0]
			if req.Model != "test-model" {
				t.Errorf("expected model %q, got %q", "test-model", req.Model)
			}

			// Check error cases
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify output
			var expected ChatOutput
			if err := json.Unmarshal([]byte(tt.mockResp), &expected); err != nil {
				t.Fatalf("failed to parse expected output: %v", err)
			}

			if output.Response != expected.Response {
				t.Errorf("response\nwant: %q\ngot:  %q", expected.Response, output.Response)
			}
		})
	}
}

func TestStreamingWithProvider(t *testing.T) {
	type ChatInput struct {
		Message string
	}

	type ChatOutput struct {
		Response string `json:"response" jsonschema:"required"`
	}

	mock := &provider.MockProvider{
		StreamTokens: []string{"Hello", " ", "world", "!"},
		DelayMs:      10,
	}

	gen, err := Create[ChatInput, ChatOutput]("User: {{.Message}}")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	gen.WithProvider(mock)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stream, err := gen.Stream(ctx, ChatInput{Message: "Hi"})
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	var received []string
	for token := range stream.Content {
		received = append(received, token)
	}

	want := []string{"Hello", " ", "world", "!"}
	if !reflect.DeepEqual(received, want) {
		t.Errorf("received tokens\nwant: %v\ngot:  %v", want, received)
	}

	// Wait for stream completion
	select {
	case err := <-stream.Err:
		t.Errorf("unexpected error: %v", err)
	case <-stream.Done:
		// Success
	case <-ctx.Done():
		t.Error("stream timeout")
	}
}
