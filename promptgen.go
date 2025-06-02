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
	"github.com/arjunsriva/promptgen/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Generator handles prompt generation and response validation
type Generator[I any, O any] struct {
	prompt        *template.Template
	handler       handler.Handler[O]
	provider      provider.Provider
	hooks         []Hook
	timeout       time.Duration
	operationName string // Added field
}

// Create initializes a new Generator with the given prompt template.
// The operationName is used for tracing to identify the business logic.
func Create[I any, O any](promptTemplate string, operationName string) (*Generator[I, O], error) {
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

	g := &Generator[I, O]{
		prompt:        tmpl,
		handler:       h,
		operationName: operationName,
	}
	// Note: Options processing for provider would go here if Create accepted them.
	// For now, ensureDefaultConfig in Run/Stream handles provider initialization.
	return g, nil
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
	var output O // Default output

	spanName := g.operationName
	if spanName == "" {
		spanName = "promptgen.Run" // Default span name
	}
	spanCtx, span := tracing.Tracer.Start(ctx, spanName)
	defer span.End()

	if g.operationName != "" {
		span.SetAttributes(attribute.String("promptgen.operation_name", g.operationName))
	}

	if err := g.ensureDefaultConfig(); err != nil {
		err = &Error{Err: ErrConfiguration, Message: err.Error(), Code: "config_error"}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return output, err
	}

	// Apply default timeout if set
	// Note: The original context `ctx` is used for timeout, then `spanCtx` is passed down.
	// This means the timeout covers the span's operations.
	runCtx := spanCtx // Use spanCtx for operations within the span
	if g.timeout > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(spanCtx, g.timeout) // Use spanCtx here
		defer cancel()
	}

	// Execute template
	var buf bytes.Buffer
	if err := g.prompt.Execute(&buf, input); err != nil {
		err = fmt.Errorf("failed to execute template: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return output, err
	}

	// Wrap prompt with type-specific instructions
	wrappedPrompt := g.handler.WrapPrompt(buf.String())

	// Run before hooks
	for _, hook := range g.hooks {
		var errHook error
		// Pass runCtx (which might have timeout) to hooks
		wrappedPrompt, errHook = hook.BeforeRequest(runCtx, wrappedPrompt)
		if errHook != nil {
			errHook = fmt.Errorf("hook error (BeforeRequest): %w", errHook)
			span.RecordError(errHook)
			span.SetStatus(codes.Error, errHook.Error())
			return output, errHook
		}
	}

	// Call provider
	// Pass runCtx (which might have timeout and is child of spanCtx) to provider
	response, err := g.provider.Complete(runCtx, wrappedPrompt)

	// Check for context/timeout errors first
	if err != nil {
		span.RecordError(err) // Record provider error
		span.SetStatus(codes.Error, err.Error())
		// Specific error handling based on provider errors or context errors
		if runCtx.Err() == context.DeadlineExceeded || errors.Is(err, context.DeadlineExceeded) {
			return output, ErrTimeout
		}
		if runCtx.Err() == context.Canceled || errors.Is(err, context.Canceled) {
			return output, fmt.Errorf("request canceled: %w", err)
		}
		if errors.Is(err, provider.ErrRateLimit) {
			return output, ErrRateLimit
		}
		if errors.Is(err, provider.ErrContextLength) {
			return output, ErrContextLength
		}
		return output, err // General provider error
	}

	// Check context after successful response (e.g. if timeout happened during provider call but provider didn't error)
	if runCtx.Err() != nil {
		err = runCtx.Err()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if err == context.DeadlineExceeded {
			return output, ErrTimeout
		}
		return output, fmt.Errorf("request canceled post-provider: %w", err)
	}

	// after response hooks
	for _, hook := range g.hooks {
		var errHook error
		// Pass runCtx (which might have timeout) to hooks
		response, errHook = hook.AfterResponse(runCtx, response, nil) // Passing nil for error as provider call was successful
		if errHook != nil {
			errHook = fmt.Errorf("hook error (AfterResponse): %w", errHook)
			span.RecordError(errHook)
			span.SetStatus(codes.Error, errHook.Error())
			return output, errHook
		}
	}

	// Parse response
	parsedOutput, err := g.handler.Parse(response)
	if err != nil {
		err = &Error{Err: ErrInvalidResponse, Message: fmt.Sprintf("failed to parse response: %v", err), Code: "parse_failed"}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return output, err // output is zero value of O
	}
	output = parsedOutput // Assign successfully parsed output

	// Validate output
	if err := g.handler.Validate(output); err != nil {
		err = &Error{Err: ErrValidation, Message: err.Error(), Code: "validation_failed"}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return output, err // output here is the parsed but invalid output
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
