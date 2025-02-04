package promptgen

import (
	"context"
	"testing"

	"github.com/arjunsriva/promptgen/provider"
)

func TestHooks(t *testing.T) {
	mockProvider := &provider.MockProvider{
		Response: `{"response": "Hello"}`,
	}

	hook := &TestHook{
		beforeResponse: "Modified prompt",
	}

	gen, _ := Create[TestInput, TestOutput]("Hello {{.Message}}")
	gen.WithProvider(mockProvider).WithHook(hook)

	_, err := gen.Run(context.Background(), TestInput{Message: "test"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !hook.beforeCalled {
		t.Error("before hook not called")
	}
}

type TestHook struct {
	beforeCalled   bool
	beforeResponse string
	afterCalled    bool
	afterResponse  string
	afterError     error
}

func (h *TestHook) BeforeRequest(_ context.Context, prompt string) (string, error) {
	h.beforeCalled = true
	if h.beforeResponse != "" {
		return h.beforeResponse, nil
	}
	return prompt, nil
}

func (h *TestHook) AfterResponse(_ context.Context, response string, err error) (string, error) {
	h.afterCalled = true
	h.afterResponse = response
	return response, h.afterError
}
