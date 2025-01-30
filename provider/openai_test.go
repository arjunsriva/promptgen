//go:build integration
// +build integration

package provider

import (
	"context"
	"os"
	"strings"
	"testing"
)

// These tests require:
// 1. OPENAI_API_KEY environment variable to be set
// 2. -tags=integration flag when running tests
// Example: go test -tags=integration ./pkg/provider -run TestOpenAI

func TestOpenAICompletion(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenAI integration test: OPENAI_API_KEY not set")
	}

	provider := NewOpenAI(apiKey)

	tests := []struct {
		name        string
		req         Request
		wantErr     bool
		errContains string
	}{
		{
			name: "basic completion",
			req: Request{
				Prompt:      "Say hello",
				Model:       "gpt-3.5-turbo",
				Temperature: 0.7,
			},
			wantErr: false,
		},
		{
			name: "invalid model",
			req: Request{
				Prompt:      "Say hello",
				Model:       "invalid-model",
				Temperature: 0.7,
			},
			wantErr:     true,
			errContains: "The model `invalid-model` does not exist",
		},
		// Add more test cases for rate limits, context length, etc.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := provider.Complete(context.Background(), tt.req)
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
			if resp == "" {
				t.Error("expected non-empty response")
			}
		})
	}
}

func TestOpenAIStreaming(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenAI integration test: OPENAI_API_KEY not set")
	}

	provider := NewOpenAI(apiKey)

	contentChan, errChan, err := provider.Stream(context.Background(), Request{
		Prompt:      "Count from 1 to 5",
		Model:       "gpt-4o-mini",
		Temperature: 0.7,
	})
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	var tokens []string
	for token := range contentChan {
		tokens = append(tokens, token)
	}

	// Now check for errors after stream is complete
	if err := <-errChan; err != nil {
		t.Errorf("stream error: %v", err)
	}

	if len(tokens) == 0 {
		t.Error("expected some tokens")
	}
}
