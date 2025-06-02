package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/arjunsriva/promptgen/tracing"
	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// OpenAIConfig holds configuration for the OpenAI provider
type OpenAIConfig struct {
	APIKey      string
	Model       string
	Temperature float64
	MaxTokens   int
}

// OpenAI implements the Provider interface using OpenAI's API
type OpenAI struct {
	client *openai.Client
	config OpenAIConfig
}

// DefaultOpenAI creates a new OpenAI provider with default configuration
func DefaultOpenAI() (*OpenAI, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	return NewOpenAI(OpenAIConfig{
		APIKey:      apiKey,
		Model:       "gpt-4-turbo-preview",
		Temperature: 0.7,
		MaxTokens:   2000,
	})
}

// NewOpenAI creates a new OpenAI provider with the given configuration
func NewOpenAI(config OpenAIConfig) (*OpenAI, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	return &OpenAI{
		client: openai.NewClient(config.APIKey),
		config: config,
	}, nil
}

// Complete generates a completion for the given prompt using OpenAI's API
func (o *OpenAI) Complete(ctx context.Context, prompt string) (string, error) {
	spanCtx, span := tracing.Tracer.Start(ctx, "OpenAI.Complete")
	defer span.End()

	span.SetAttributes(
		attribute.String("llm.provider", "OpenAI"),
		attribute.String("llm.model_name", o.config.Model),
		attribute.Float64("llm.temperature", o.config.Temperature),
		attribute.Int("llm.max_tokens", o.config.MaxTokens),
	)

	resp, err := o.client.CreateChatCompletion(
		spanCtx,
		openai.ChatCompletionRequest{
			Model: o.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: float32(o.config.Temperature),
			MaxTokens:   o.config.MaxTokens,
		},
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		var apiErr *openai.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.HTTPStatusCode {
			case 429:
				return "", fmt.Errorf("%w: %v", ErrRateLimit, err)
			case 400:
				if strings.Contains(apiErr.Message, "maximum context length") {
					return "", fmt.Errorf("%w: %v", ErrContextLength, err)
				}
			}
		}
		return "", fmt.Errorf("openai completion failed: %w", err)
	}

	if resp.Usage.PromptTokens > 0 { // Check if Usage is available
		span.SetAttributes(
			attribute.Int("llm.usage.prompt_tokens", resp.Usage.PromptTokens),
			attribute.Int("llm.usage.completion_tokens", resp.Usage.CompletionTokens),
			attribute.Int("llm.usage.total_tokens", resp.Usage.TotalTokens),
		)
	}

	return resp.Choices[0].Message.Content, nil
}

// Stream generates a completion and streams the response using OpenAI's API
func (o *OpenAI) Stream(ctx context.Context, prompt string) (contentChan <-chan string, errChan <-chan error, err error) {
	spanCtx, span := tracing.Tracer.Start(ctx, "OpenAI.Stream")

	span.SetAttributes(
		attribute.String("llm.provider", "OpenAI"),
		attribute.String("llm.model_name", o.config.Model),
		attribute.Float64("llm.temperature", o.config.Temperature),
		attribute.Int("llm.max_tokens", o.config.MaxTokens),
	)

	stream, err := o.client.CreateChatCompletionStream(
		spanCtx,
		openai.ChatCompletionRequest{
			Model: o.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: float32(o.config.Temperature),
			MaxTokens:   o.config.MaxTokens,
		},
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End() // End span if initial call fails
		return nil, nil, fmt.Errorf("openai stream failed: %w", err)
	}

	content := make(chan string)
	errs := make(chan error)

	go func(routineSpan trace.Span) {
		defer routineSpan.End() // End the span when the goroutine exits
		defer stream.Close()
		defer close(content)
		defer close(errs)

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return // Success, span will be ended by defer
			}
			if err != nil {
				routineSpan.RecordError(err)
				routineSpan.SetStatus(codes.Error, err.Error())
				errs <- fmt.Errorf("stream receive failed: %w", err)
				return // Error, span will be ended by defer
			}

			if len(response.Choices) > 0 {
				content <- response.Choices[0].Delta.Content
			}
		}
	}(span) // Pass the original span to the goroutine

	return content, errs, nil
}
