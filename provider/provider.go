package provider

import (
	"context"
	"errors"
)

// Request represents a request to an AI provider
type Request struct {
	Prompt      string
	Model       string
	Temperature float64
	MaxTokens   int
}

// Provider defines the interface for AI providers
type Provider interface {
	// Complete generates a completion for the given request
	Complete(ctx context.Context, req Request) (string, error)

	// Stream generates a completion and streams the response
	Stream(ctx context.Context, req Request) (<-chan string, <-chan error, error)
}

// Common provider errors
var (
	ErrRateLimit     = errors.New("rate limit exceeded")
	ErrContextLength = errors.New("context length exceeded")
)
