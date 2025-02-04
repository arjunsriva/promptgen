package provider

import (
	"context"
	"errors"
)

// Provider defines the interface for AI providers
type Provider interface {
	// Complete generates a completion for the given prompt
	Complete(ctx context.Context, prompt string) (string, error)

	// Stream generates a completion and streams the response
	Stream(ctx context.Context, prompt string) (<-chan string, <-chan error, error)
}

// Common provider errors
var (
	ErrRateLimit     = errors.New("rate limit exceeded")
	ErrContextLength = errors.New("context length exceeded")
)
