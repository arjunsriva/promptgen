package promptgen

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/arjunsriva/promptgen/provider"
)

// streamTestHook is a simple hook implementation for testing
type streamTestHook struct {
	beforePrefix string
	afterSuffix  string
}

func (h *streamTestHook) BeforeRequest(ctx context.Context, prompt string) (string, error) {
	return h.beforePrefix + " " + prompt, nil
}

func (h *streamTestHook) AfterResponse(ctx context.Context, response string, err error) (string, error) {
	return response + " " + h.afterSuffix, err
}

func TestStream(t *testing.T) {
	t.Run("successful stream", func(t *testing.T) {
		mock := &provider.MockProvider{
			Response:     "Hello world!",
			StreamTokens: []string{"Hello", " ", "world", "!"},
			DelayMs:      10,
		}

		gen, _ := Create[TestInput, TestOutput]("test")
		gen.WithProvider(mock)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		stream, err := gen.Stream(ctx, TestInput{Message: "test"})
		if err != nil {
			t.Fatalf("Stream() error = %v", err)
		}

		var tokens []string
		for token := range stream.Content {
			tokens = append(tokens, token)
		}

		want := []string{"Hello", " ", "world", "!"}
		if !reflect.DeepEqual(tokens, want) {
			t.Errorf("Stream() tokens = %v, want %v", tokens, want)
		}

		select {
		case err := <-stream.Err:
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		case <-stream.Done:
			// Success
		case <-time.After(time.Second):
			t.Error("stream did not complete")
		}
	})

	t.Run("missing provider", func(t *testing.T) {
		gen, _ := Create[TestInput, TestOutput]("test")
		os.Unsetenv("OPENAI_API_KEY")

		_, err := gen.Stream(context.Background(), TestInput{Message: "test"})
		if err == nil {
			t.Error("expected error for missing provider")
		}
	})

	t.Run("provider error", func(t *testing.T) {
		mock := &provider.MockProvider{
			Errors: []error{errors.New("stream error")},
		}

		gen, _ := Create[TestInput, TestOutput]("test")
		gen.WithProvider(mock)

		stream, err := gen.Stream(context.Background(), TestInput{Message: "test"})
		if err != nil {
			t.Fatalf("Stream() error = %v", err)
		}

		select {
		case err := <-stream.Err:
			if err == nil {
				t.Error("expected error from stream")
			}
		case <-stream.Done:
			t.Error("stream completed despite error")
		case <-time.After(time.Second):
			t.Error("stream did not complete")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		t.Skip("TODO: Add tests for stream cancellation behavior")
	})

	t.Run("with hooks", func(t *testing.T) {
		mock := &provider.MockProvider{
			StreamTokens: []string{"Hello", "world"},
		}

		hook := &streamTestHook{
			beforePrefix: "modified",
			afterSuffix:  "modified",
		}

		gen, _ := Create[TestInput, TestOutput]("test")
		gen.WithProvider(mock).WithHook(hook)

		stream, err := gen.Stream(context.Background(), TestInput{Message: "test"})
		if err != nil {
			t.Fatalf("Stream() error = %v", err)
		}

		var tokens []string
		for token := range stream.Content {
			tokens = append(tokens, token)
		}

		want := []string{"Hello modified", "world modified"}
		if !reflect.DeepEqual(tokens, want) {
			t.Errorf("Stream() tokens = %v, want %v", tokens, want)
		}

		// Verify the prompt was modified by the before hook
		if len(mock.Prompts) != 1 {
			t.Errorf("expected 1 prompt, got %d", len(mock.Prompts))
		} else if !strings.HasPrefix(mock.Prompts[0], "modified") {
			t.Errorf("before hook was not applied, got prompt: %q", mock.Prompts[0])
		}
	})
}
