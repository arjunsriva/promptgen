// Package promptgen provides a type-safe framework for building AI-powered applications in Go.
// It combines Go's template system with JSON Schema validation to ensure reliable AI interactions.
//
// Basic usage:
//
//	type Input struct {
//	    Message string
//	}
//
//	type Output struct {
//	    Response string `json:"response" jsonschema:"required,maxLength=100"`
//	}
//
//	generator := promptgen.Create[Input, Output]("Respond to: {{.Message}}")
//	result, err := generator.Run(context.Background(), Input{Message: "Hello"})
//
// See https://pkg.go.dev/github.com/arjunsriva/promptgen for full documentation.
package promptgen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"text/template"
	"time"

	"github.com/arjunsriva/promptgen/internal/handler"
	jsonhandler "github.com/arjunsriva/promptgen/internal/json"
	"github.com/arjunsriva/promptgen/internal/primitive"
	"github.com/arjunsriva/promptgen/provider"
)

// Generator handles prompt generation and response validation
type Generator[I any, O any] struct {
	prompt   *template.Template
	handler  handler.Handler[O]
	provider provider.Provider
	hooks    []Hook
	timeout  time.Duration
}

// Create initializes a new Generator with the given prompt template
func Create[I any, O any](promptTemplate string) (*Generator[I, O], error) {
	// Parse the template
	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	// Validate template variables by executing with zero value
	var zero I
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, zero); err != nil {
		return nil, fmt.Errorf("invalid template variables: %w", err)
	}

	// Get or create handler
	var h handler.Handler[O]
	switch handler.DetermineType[O]() {
	case handler.TypeString, handler.TypePrimitive:
		h, err = primitive.New[O]()
	case handler.TypeJSON:
		h, err = jsonhandler.New[O]()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create handler: %w", err)
	}

	return &Generator[I, O]{
		prompt:  tmpl,
		handler: h,
	}, nil
}

// Add this private method to handle default configuration
func (g *Generator[I, O]) ensureDefaultConfig() error {
	if g.provider == nil {
		defaultProvider, err := provider.DefaultOpenAI()
		if err != nil {
			return fmt.Errorf("failed to create default provider: %w", err)
		}
		g.provider = defaultProvider
	}
	return nil
}

// Run executes the prompt with the given input and returns the validated output
func (g *Generator[I, O]) Run(ctx context.Context, input I) (O, error) {
	var output O

	if err := g.ensureDefaultConfig(); err != nil {
		return output, &Error{
			Err:     ErrConfiguration,
			Message: err.Error(),
			Code:    "config_error",
		}
	}

	// Apply default timeout if set
	if g.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.timeout)
		defer cancel()
	}

	// Execute template
	var buf bytes.Buffer
	if err := g.prompt.Execute(&buf, input); err != nil {
		return output, fmt.Errorf("failed to execute template: %w", err)
	}

	// Wrap prompt with type-specific instructions
	wrappedPrompt := g.handler.WrapPrompt(buf.String())

	// Run before hooks
	for _, hook := range g.hooks {
		var err error
		wrappedPrompt, err = hook.BeforeRequest(ctx, wrappedPrompt)
		if err != nil {
			return output, fmt.Errorf("hook error: %w", err)
		}
	}

	// Call provider
	response, err := g.provider.Complete(ctx, wrappedPrompt)

	// Check for context/timeout errors first
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded) || ctx.Err() == context.DeadlineExceeded:
			return output, ErrTimeout
		case errors.Is(err, context.Canceled):
			return output, fmt.Errorf("request canceled: %w", err)
		case errors.Is(err, provider.ErrRateLimit):
			return output, ErrRateLimit
		case errors.Is(err, provider.ErrContextLength):
			return output, ErrContextLength
		default:
			return output, err
		}
	}

	// Check context after successful response
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			return output, ErrTimeout
		case context.Canceled:
			return output, fmt.Errorf("request canceled: %w", ctx.Err())
		default:
			return output, ctx.Err()
		}
	}

	// after response hooks
	for _, hook := range g.hooks {
		var err error
		response, err = hook.AfterResponse(ctx, response, err)
		if err != nil {
			return output, fmt.Errorf("hook error: %w", err)
		}
	}

	// Parse response
	output, err = g.handler.Parse(response)
	if err != nil {
		return output, &Error{
			Err:     ErrInvalidResponse,
			Message: fmt.Sprintf("failed to parse response: %v", err),
			Code:    "parse_failed",
		}
	}

	// Validate output
	if err := g.handler.Validate(output); err != nil {
		return output, &Error{
			Err:     ErrValidation,
			Message: err.Error(),
			Code:    "validation_failed",
		}
	}

	return output, nil
}

// WithProvider sets the AI provider to use
func (g *Generator[I, O]) WithProvider(p provider.Provider) *Generator[I, O] {
	g.provider = p
	return g
}

// WithHook adds a hook to the generator
func (g *Generator[I, O]) WithHook(hook Hook) *Generator[I, O] {
	g.hooks = append(g.hooks, hook)
	return g
}

// WithTimeout sets a default timeout for requests
func (g *Generator[I, O]) WithTimeout(timeout time.Duration) *Generator[I, O] {
	g.timeout = timeout
	return g
}

// WithHandler sets a custom handler implementation
func (g *Generator[I, O]) WithHandler(h handler.Handler[O]) *Generator[I, O] {
	g.handler = h
	return g
}
