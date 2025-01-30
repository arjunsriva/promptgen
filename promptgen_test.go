package promptgen

import (
	"context"
	"errors"
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
		wantResp    string
		wantErr     bool
		errContains string
		errIs       error
	}{
		{
			name: "valid markdown json response",
			input: ChatInput{
				Message: "Hello!",
			},
			mockResp: "Here's my response:\n```json\n{\"response\": \"Hi there!\"}\n```",
			wantResp: "Hi there!",
			wantErr:  false,
		},
		{
			name: "valid markdown response without json tag",
			input: ChatInput{
				Message: "Hello!",
			},
			mockResp: "Let me think...\n```\n{\"response\": \"Hi there!\"}\n```",
			wantResp: "Hi there!",
			wantErr:  false,
		},
		{
			name: "missing markdown blocks",
			input: ChatInput{
				Message: "Hello!",
			},
			mockResp:    "{\"response\": \"Hi there!\"}",
			wantErr:     true,
			errContains: "no JSON found in response",
		},
		{
			name: "invalid json inside markdown",
			input: ChatInput{
				Message: "Hello!",
			},
			mockResp:    "```json\nnot a json response\n```",
			wantErr:     true,
			errIs:       ErrValidation,
			errContains: "validation failed",
		},
		{
			name: "response too long",
			input: ChatInput{
				Message: "Hello!",
			},
			mockResp: "```json\n{\"response\": \"This response is way too long and exceeds the maximum length of 100 characters that we specified in the JSON Schema validation rules\"}\n```",
			wantErr:  true,
			errIs:    ErrValidation,
			errContains: "validation errors: response: String length must be less than or equal to 100",
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
				} else {
					if tt.errIs != nil && !errors.Is(err, tt.errIs) {
						t.Errorf("expected error type %v, got %v", tt.errIs, err)
					}
					if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify output matches expected response
			if output.Response != tt.wantResp {
				t.Errorf("response\nwant: %q\ngot:  %q", tt.wantResp, output.Response)
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

func TestConcurrentUsage(t *testing.T) {
	gen, err := Create[TestInput, TestOutput]("test")
	if err != nil {
		t.Fatal(err)
	}

	mock := &provider.MockProvider{
		Response: "```json\n{\"response\": \"concurrent test\"}\n```",
	}
	gen.WithProvider(mock)

	const goroutines = 10
	errChan := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			_, err := gen.Run(context.Background(), TestInput{})
			errChan <- err
		}()
	}

	for i := 0; i < goroutines; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("concurrent request failed: %v", err)
		}
	}
}

// Add new test for timeout functionality
func TestGeneratorTimeout(t *testing.T) {
	gen, err := Create[TestInput, TestOutput]("test")
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock provider that delays
	mock := &provider.MockProvider{
		Response: "```json\n{\"response\": \"delayed response\"}\n```",
		DelayMs:  500, // Increase delay to 500ms to ensure timeout
	}

	gen.WithProvider(mock).
		WithTimeout(100 * time.Millisecond) // Keep timeout at 100ms

	ctx := context.Background() // Create a fresh context
	_, err = gen.Run(ctx, TestInput{})
	
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

// Add test for regex extraction
func TestJSONExtraction(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     string
		wantErr  bool
	}{
		{
			name:     "simple json block",
			response: "```json\n{\"test\": \"value\"}\n```",
			want:     "\n{\"test\": \"value\"}\n",
			wantErr:  false,
		},
		{
			name:     "json block with thinking",
			response: "Let me think about it...\n```json\n{\"test\": \"value\"}\n```",
			want:     "\n{\"test\": \"value\"}\n",
			wantErr:  false,
		},
		{
			name:     "json block without json tag",
			response: "```\n{\"test\": \"value\"}\n```",
			want:     "\n{\"test\": \"value\"}\n",
			wantErr:  false,
		},
		{
			name:     "missing json block",
			response: "{\"test\": \"value\"}",
			wantErr:  true,
		},
		{
			name:     "multiple json blocks",
			response: "```json\n{\"first\": true}\n```\n```json\n{\"second\": true}\n```",
			want:     "\n{\"first\": true}\n",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := jsonRegex.FindStringSubmatch(tt.response)
			if tt.wantErr {
				if len(matches) >= 2 {
					t.Error("expected no match, got one")
				}
				return
			}
			if len(matches) < 2 {
				t.Fatal("expected match, got none")
			}
			if got := matches[1]; got != tt.want {
				t.Errorf("extracted JSON\nwant: %q\ngot:  %q", tt.want, got)
			}
		})
	}
}
