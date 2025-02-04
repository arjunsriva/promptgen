//go:build integration
// +build integration

package provider

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
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

	provider, err := NewOpenAI(OpenAIConfig{
		APIKey:      apiKey,
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   100,
	})
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	tests := []struct {
		name        string
		prompt      string
		config      OpenAIConfig
		wantErr     bool
		errContains string
	}{
		{
			name:   "basic completion",
			prompt: "Say hello",
			config: OpenAIConfig{
				APIKey:      apiKey,
				Model:       "gpt-3.5-turbo",
				Temperature: 0.7,
				MaxTokens:   100,
			},
		},
		{
			name:   "invalid model",
			prompt: "Say hello",
			config: OpenAIConfig{
				APIKey:      apiKey,
				Model:       "invalid-model",
				Temperature: 0.7,
				MaxTokens:   100,
			},
			wantErr:     true,
			errContains: "does not exist",
		},
		{
			name:   "empty prompt",
			prompt: "",
			config: OpenAIConfig{
				APIKey:      apiKey,
				Model:       "gpt-3.5-turbo",
				Temperature: 0.7,
				MaxTokens:   100,
			},
			wantErr:     true,
			errContains: "empty prompt",
		},
		{
			name:   "temperature too high",
			prompt: "Say hello",
			config: OpenAIConfig{
				APIKey:      apiKey,
				Model:       "gpt-3.5-turbo",
				Temperature: 2.0,
				MaxTokens:   100,
			},
			wantErr:     true,
			errContains: "temperature",
		},
		{
			name:   "very long prompt",
			prompt: strings.Repeat("test ", 5000),
			config: OpenAIConfig{
				APIKey:      apiKey,
				Model:       "gpt-3.5-turbo",
				Temperature: 0.7,
				MaxTokens:   100,
			},
			wantErr:     true,
			errContains: "maximum context length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOpenAI(tt.config)
			if err != nil {
				t.Fatalf("failed to create provider: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := provider.Complete(ctx, tt.prompt)
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
	tests := []struct {
		name        string
		prompt      string
		model       string
		temperature float32
		maxTokens   int
		wantErr     bool
		errContains string
	}{
		{
			name:        "basic streaming",
			prompt:      "Count from 1 to 5",
			model:       "gpt-3.5-turbo",
			temperature: 0.7,
			maxTokens:   100,
			wantErr:     false,
		},
		{
			name:        "invalid model",
			prompt:      "Say hello",
			model:       "invalid-model",
			temperature: 0.7,
			maxTokens:   100,
			wantErr:     true,
			errContains: "does not exist",
		},
		{
			name:        "very long prompt",
			prompt:      strings.Repeat("test ", 5000),
			model:       "gpt-3.5-turbo",
			temperature: 0.7,
			maxTokens:   100,
			wantErr:     true,
			errContains: "maximum context length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			ch := make(chan string)
			errCh := make(chan error)

			go provider.Stream(ctx, tt.prompt, tt.model, tt.temperature, tt.maxTokens, ch, errCh)

			var tokens []string
			var streamErr error

			// Use a select to handle both success and error cases
			for {
				select {
				case token, ok := <-ch:
					if !ok {
						// Channel closed, exit loop
						goto done
					}
					tokens = append(tokens, token)
				case err := <-errCh:
					streamErr = err
					goto done
				case <-ctx.Done():
					t.Fatal("test timed out")
				}
			}

		done:
			if tt.wantErr {
				if streamErr == nil {
					t.Error("expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(streamErr.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", streamErr.Error(), tt.errContains)
				}
				return
			}
			if streamErr != nil {
				t.Errorf("unexpected error: %v", streamErr)
				return
			}
			if len(tokens) == 0 {
				t.Error("expected some tokens")
			}
		})
	}
}

func TestDefaultOpenAI(t *testing.T) {
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", originalAPIKey)

	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid api key",
			apiKey:  "test-key",
			wantErr: false,
		},
		{
			name:    "empty api key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("OPENAI_API_KEY", tt.apiKey)
			provider, err := DefaultOpenAI()
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultOpenAI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if provider == nil {
					t.Error("expected non-nil provider")
				}
				if provider.config.Model != "gpt-4-turbo-preview" {
					t.Errorf("expected model gpt-4-turbo-preview, got %s", provider.config.Model)
				}
				if provider.config.Temperature != 0.7 {
					t.Errorf("expected temperature 0.7, got %f", provider.config.Temperature)
				}
				if provider.config.MaxTokens != 2000 {
					t.Errorf("expected maxTokens 2000, got %d", provider.config.MaxTokens)
				}
			}
		})
	}
}

func TestNewOpenAI(t *testing.T) {
	tests := []struct {
		name    string
		config  OpenAIConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: OpenAIConfig{
				APIKey:      "test-key",
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   100,
			},
			wantErr: false,
		},
		{
			name: "empty api key",
			config: OpenAIConfig{
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOpenAI(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOpenAI() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && provider == nil {
				t.Error("expected non-nil provider")
			}
		})
	}
}

type mockOpenAIClient struct {
	createChatCompletionFunc       func(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
	createChatCompletionStreamFunc func(context.Context, openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error)
}

func (m *mockOpenAIClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return m.createChatCompletionFunc(ctx, req)
}

func (m *mockOpenAIClient) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error) {
	return m.createChatCompletionStreamFunc(ctx, req)
}

func TestOpenAIComplete(t *testing.T) {
	config := OpenAIConfig{
		APIKey:      "test-key",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   100,
	}

	mockClient := &mockOpenAIClient{
		createChatCompletionFunc: func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
			if req.Model != config.Model {
				t.Errorf("unexpected model: got %v want %v", req.Model, config.Model)
			}
			if len(req.Messages) != 1 || req.Messages[0].Content != "test prompt" {
				t.Errorf("unexpected messages: %v", req.Messages)
			}
			return openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: "test response"}},
				},
			}, nil
		},
	}

	provider := &OpenAI{
		client: mockClient,
		config: config,
	}

	ctx := context.Background()
	got, err := provider.Complete(ctx, "test prompt")
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if want := "test response"; got != want {
		t.Errorf("Complete() = %v, want %v", got, want)
	}

	// Test error case
	mockClient.createChatCompletionFunc = func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
		return openai.ChatCompletionResponse{}, errors.New("api error")
	}
	if _, err := provider.Complete(ctx, "test prompt"); err == nil {
		t.Error("Complete() expected error")
	}
}

type mockChatCompletionStream struct {
	recvFunc  func() (openai.ChatCompletionStreamResponse, error)
	closeFunc func() error
}

func (m *mockChatCompletionStream) Recv() (openai.ChatCompletionStreamResponse, error) {
	return m.recvFunc()
}

func (m *mockChatCompletionStream) Close() error {
	return m.closeFunc()
}

func TestOpenAIStream(t *testing.T) {
	mockStream := &mockChatCompletionStream{
		recvFunc: func() (openai.ChatCompletionStreamResponse, error) {
			return openai.ChatCompletionStreamResponse{
				Choices: []openai.ChatCompletionStreamChoice{
					{Delta: openai.ChatCompletionStreamChoiceDelta{Content: "test"}},
				},
			}, nil
		},
		closeFunc: func() error { return nil },
	}

	mockClient := &mockOpenAIClient{
		createChatCompletionStreamFunc: func(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error) {
			if req.Model != "gpt-3.5-turbo" {
				t.Errorf("unexpected model: got %v want gpt-3.5-turbo", req.Model)
			}
			if len(req.Messages) != 1 || req.Messages[0].Content != "test prompt" {
				t.Errorf("unexpected messages: %v", req.Messages)
			}
			return &openai.ChatCompletionStream{}, nil
		},
	}

	provider := &OpenAI{
		client: mockClient,
		config: openai.DefaultConfig("test-key"),
	}

	ctx := context.Background()
	ch := make(chan string)
	errCh := make(chan error)

	go provider.Stream(ctx, "test prompt", "gpt-3.5-turbo", 0.7, 100, ch, errCh)

	// Test error case
	mockClient.createChatCompletionStreamFunc = func(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error) {
		return nil, errors.New("stream error")
	}
	go provider.Stream(ctx, "test prompt", "gpt-3.5-turbo", 0.7, 100, ch, errCh)
	select {
	case err := <-errCh:
		if err == nil {
			t.Error("Stream() expected error")
		}
	case <-ch:
		t.Error("Stream() unexpected response")
	}
}
