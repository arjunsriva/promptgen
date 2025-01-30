package promptgen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/arjunsriva/promptgen/internal/schema"
	"github.com/arjunsriva/promptgen/provider"
)

// Generator handles prompt generation and response validation
type Generator[I any, O any] struct {
	prompt    *template.Template
	validator *schema.Validator[O]
	provider  provider.Provider
	model     string
	temp      float64
	maxTokens int
}

// Create initializes a new Generator with the given prompt template
func Create[I any, O any](promptTemplate string) (*Generator[I, O], error) {
	// Parse the template
	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	// Create validator for output type
	validator, err := schema.NewValidator[O]()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	// Initialize with default values
	return &Generator[I, O]{
		prompt:    tmpl,
		validator: validator,
		model:     "gpt-4o-mini", // Default model
		temp:      0.7,           // Default temperature
		maxTokens: 2000,          // Default max tokens
	}, nil
}

// Run executes the prompt with the given input and returns the validated output
func (g *Generator[I, O]) Run(ctx context.Context, input I) (O, error) {
	var output O

	// Ensure provider is configured
	if err := g.ensureProvider(); err != nil {
		return output, err
	}

	// Execute template
	var buf bytes.Buffer
	if err := g.prompt.Execute(&buf, input); err != nil {
		return output, fmt.Errorf("failed to execute template: %w", err)
	}

	// Get schema for output type
	schema, err := g.validator.SchemaString()
	if err != nil {
		return output, fmt.Errorf("failed to get schema: %w", err)
	}

	// Create prompt with schema
	prompt := fmt.Sprintf(`Given this JSON Schema:
%s

Generate a valid JSON response for this prompt:
%s`, schema, buf.String())

	// Call provider
	response, err := g.provider.Complete(ctx, provider.Request{
		Prompt:      prompt,
		Model:       g.model,
		Temperature: g.temp,
		MaxTokens:   g.maxTokens,
	})
	if err != nil {
		return output, fmt.Errorf("provider completion failed: %w", err)
	}

	// Validate response
	if err := g.validator.Validate([]byte(response)); err != nil {
		return output, fmt.Errorf("invalid response: %w", err)
	}

	// Parse response into output type
	if err := json.Unmarshal([]byte(response), &output); err != nil {
		return output, fmt.Errorf("failed to parse response: %w", err)
	}

	return output, nil
}

// ValidateResponse validates a JSON response against the output schema
func (g *Generator[I, O]) ValidateResponse(response []byte) error {
	return g.validator.Validate(response)
}

// SchemaString returns the JSON Schema for the output type
func (g *Generator[I, O]) SchemaString() (string, error) {
	return g.validator.SchemaString()
}

// Stream represents a real-time stream of content from the AI provider
type Stream struct {
	Content chan string
	Err     chan error
	Done    chan struct{}
}

// WithProvider sets the AI provider to use
func (g *Generator[I, O]) WithProvider(p provider.Provider) *Generator[I, O] {
	g.provider = p
	return g
}

// WithModel sets the AI model to use
func (g *Generator[I, O]) WithModel(model string) *Generator[I, O] {
	g.model = model
	return g
}

// WithTemperature sets the sampling temperature
func (g *Generator[I, O]) WithTemperature(temp float64) *Generator[I, O] {
	g.temp = temp
	return g
}

// WithMaxTokens sets the maximum number of tokens to generate
func (g *Generator[I, O]) WithMaxTokens(tokens int) *Generator[I, O] {
	g.maxTokens = tokens
	return g
}

// Stream provides real-time streaming of the generated content
func (g *Generator[I, O]) Stream(ctx context.Context, input I) (*Stream, error) {
	// Ensure provider is configured
	if err := g.ensureProvider(); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := g.prompt.Execute(&buf, input); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	contentChan, errChan, err := g.provider.Stream(ctx, provider.Request{
		Prompt:      buf.String(),
		Model:       g.model,
		Temperature: g.temp,
		MaxTokens:   g.maxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("provider stream failed: %w", err)
	}

	stream := &Stream{
		Content: make(chan string),
		Err:     make(chan error),
		Done:    make(chan struct{}),
	}

	go func() {
		defer func() {
			close(stream.Content)
			stream.Done <- struct{}{} // Signal completion before closing
			close(stream.Done)
			close(stream.Err)
		}()

		for {
			select {
			case content, ok := <-contentChan:
				if !ok {
					return
				}
				stream.Content <- content
			case err, ok := <-errChan:
				if !ok {
					continue
				}
				stream.Err <- err
				return
			case <-ctx.Done():
				stream.Err <- ctx.Err()
				return
			}
		}
	}()

	return stream, nil
}

// Add a new method to set default provider if none is configured
func (g *Generator[I, O]) ensureProvider() error {
	if g.provider == nil {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("OPENAI_API_KEY environment variable is required")
		}
		g.provider = provider.NewOpenAI(apiKey)
	}
	return nil
}
