package promptgen

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arjunsriva/promptgen/provider"
)

// Test types
type TestInput struct {
	Message string
}

type TestOutput struct {
	Response string `json:"response" jsonschema:"required"`
}

// Core functionality tests
func TestCreate(t *testing.T) {
	t.Run("valid template", func(t *testing.T) {
		_, err := Create[TestInput, TestOutput]("Hello {{.Message}}")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid template", func(t *testing.T) {
		_, err := Create[TestInput, TestOutput]("Hello {{.Invalid}}")
		if err == nil {
			t.Error("expected error for invalid template")
		}
	})

	t.Run("string type", func(t *testing.T) {
		_, err := Create[string, string]("Tell me a {{.}} joke")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestTemplateExecution(t *testing.T) {
	template := "Hello {{.Message}}"
	generator, err := Create[TestInput, TestOutput](template)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	input := TestInput{
		Message: "how are you?",
	}

	var buf strings.Builder
	err = generator.prompt.Execute(&buf, input)
	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	expected := "Hello how are you?"
	if got := buf.String(); got != expected {
		t.Errorf("template execution\nwant: %q\ngot:  %q", expected, got)
	}
}

// Provider integration tests
func TestGeneratorWithProvider(t *testing.T) {
	mockProvider := &MockProvider{
		Response: `{"response": "Hello world"}`,
	}

	t.Run("successful response", func(t *testing.T) {
		gen, _ := Create[TestInput, TestOutput]("Hello {{.Message}}")
		gen.WithProvider(mockProvider)

		result, err := gen.Run(context.Background(), TestInput{Message: "test"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.Response != "Hello world" {
			t.Errorf("expected 'Hello world', got %q", result.Response)
		}
	})

	t.Run("string response", func(t *testing.T) {
		mockProvider.Response = "Hello world"
		gen, _ := Create[string, string]("Say {{.}}")
		gen.WithProvider(mockProvider)

		result, err := gen.Run(context.Background(), "hello")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "Hello world" {
			t.Errorf("expected 'Hello world', got %q", result)
		}
	})

	t.Run("provider error", func(t *testing.T) {
		mockProvider.Errors = []error{provider.ErrRateLimit}
		gen, _ := Create[TestInput, TestOutput]("Hello {{.Message}}")
		gen.WithProvider(mockProvider)

		_, err := gen.Run(context.Background(), TestInput{Message: "test"})
		if !errors.Is(err, ErrRateLimit) {
			t.Errorf("expected rate limit error, got %v", err)
		}
	})
}

func TestConcurrentUsage(t *testing.T) {
	gen, err := Create[TestInput, TestOutput]("test")
	if err != nil {
		t.Fatal(err)
	}

	mock := &MockProvider{
		Response: `{"response": "concurrent test"}`,
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

func TestTimeout(t *testing.T) {
	slowProvider := &MockProvider{
		Response: `{"response": "Hello"}`,
		DelayMs:  100,
	}

	gen, _ := Create[TestInput, TestOutput]("Hello {{.Message}}")
	gen.WithProvider(slowProvider).WithTimeout(50 * time.Millisecond)

	_, err := gen.Run(context.Background(), TestInput{Message: "test"})
	if err == nil || !errors.Is(err, ErrTimeout) {
		t.Errorf("expected timeout error, got %v", err)
	}
}

func TestEnsureDefaultConfig(t *testing.T) {
	t.Run("missing api key", func(t *testing.T) {
		gen, _ := Create[TestInput, TestOutput]("test")
		os.Unsetenv("OPENAI_API_KEY")

		err := gen.ensureDefaultConfig()
		if err == nil {
			t.Error("expected error for missing API key")
		}
	})

	t.Run("with api key", func(t *testing.T) {
		gen, _ := Create[TestInput, TestOutput]("test")
		os.Setenv("OPENAI_API_KEY", "test-key")
		defer os.Unsetenv("OPENAI_API_KEY")

		err := gen.ensureDefaultConfig()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if gen.provider == nil {
			t.Error("expected provider to be set")
		}
	})

	t.Run("custom provider preserved", func(t *testing.T) {
		gen, _ := Create[TestInput, TestOutput]("test")
		mockProvider := &MockProvider{}
		gen.WithProvider(mockProvider)

		err := gen.ensureDefaultConfig()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if gen.provider != mockProvider {
			t.Error("custom provider was not preserved")
		}
	})
}

// MockProvider for testing
type MockProvider struct {
	Response string
	Errors   []error
	DelayMs  int
}

func (m *MockProvider) Complete(ctx context.Context, prompt string) (string, error) {
	if m.DelayMs > 0 {
		time.Sleep(time.Duration(m.DelayMs) * time.Millisecond)
	}

	if len(m.Errors) > 0 {
		err := m.Errors[0]
		m.Errors = m.Errors[1:]
		return "", err
	}

	return m.Response, nil
}

func (m *MockProvider) Stream(ctx context.Context, prompt string) (contentChan <-chan string, errChan <-chan error, err error) {
	content := make(chan string)
	errs := make(chan error)

	go func() {
		defer close(content)
		defer close(errs)

		if m.DelayMs > 0 {
			time.Sleep(time.Duration(m.DelayMs) * time.Millisecond)
		}

		if len(m.Errors) > 0 {
			err := m.Errors[0]
			m.Errors = m.Errors[1:]
			errs <- err
			return
		}

		content <- m.Response
	}()

	return content, errs, nil
}
