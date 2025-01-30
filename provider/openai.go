package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// OpenAI implements the Provider interface using OpenAI's API
type OpenAI struct {
	client *openai.Client
}

// NewOpenAI creates a new OpenAI provider with the given API key
func NewOpenAI(apiKey string) *OpenAI {
	return &OpenAI{
		client: openai.NewClient(apiKey),
	}
}

// Complete generates a completion for the given request using OpenAI's API
func (o *OpenAI) Complete(ctx context.Context, req Request) (string, error) {
	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: req.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: req.Prompt,
				},
			},
			Temperature: float32(req.Temperature),
			MaxTokens:   req.MaxTokens,
		},
	)

	if err != nil {
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

	return resp.Choices[0].Message.Content, nil
}

// Stream generates a completion and streams the response using OpenAI's API
func (o *OpenAI) Stream(ctx context.Context, req Request) (<-chan string, <-chan error, error) {
	stream, err := o.client.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model: req.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: req.Prompt,
				},
			},
			Temperature: float32(req.Temperature),
			MaxTokens:   req.MaxTokens,
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("openai stream failed: %w", err)
	}

	contentChan := make(chan string)
	errChan := make(chan error)

	go func() {
		defer stream.Close()
		defer close(contentChan)
		defer close(errChan)

		for {
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					// Stream completed normally
					return
				}
				errChan <- fmt.Errorf("stream receive failed: %w", err)
				return
			}
			contentChan <- response.Choices[0].Delta.Content
		}
	}()

	return contentChan, errChan, nil
}
