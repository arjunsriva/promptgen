package promptgen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"text/template"

	"github.com/arjunsriva/promptgen/internal/schema"
	"github.com/arjunsriva/promptgen/provider"
)

// Hook represents a function that can intercept and modify requests/responses
type Hook interface {
	BeforeRequest(ctx context.Context, prompt string) (string, error)
	AfterResponse(ctx context.Context, response string, err error) (string, error)
}

// Generator handles prompt generation and response validation
type Generator[I any, O any] struct {
	prompt    *template.Template
	validator *schema.Validator[O]
	provider  provider.Provider
	model     string
	temp      float64
	maxTokens int
	hooks     []Hook
}

var jsonRegex = regexp.MustCompile("```(?:json)?([\\s\\S]*?)```")

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

	return &Generator[I, O]{
		prompt:    tmpl,
		validator: validator,
	}, nil
}

// Add this private method to handle default configuration
func (g *Generator[I, O]) ensureDefaultConfig() error {
	if g.provider == nil {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("OPENAI_API_KEY environment variable is required")
		}
		g.provider = provider.NewOpenAI(apiKey)
	}

	// Set default values if not configured
	if g.model == "" {
		g.model = "gpt-4o-mini"
	}
	if g.temp == 0 {
		g.temp = 0.7
	}
	if g.maxTokens == 0 {
		g.maxTokens = 2000
	}
	return nil
}

// Run executes the prompt with the given input and returns the validated output
func (g *Generator[I, O]) Run(ctx context.Context, input I) (O, error) {
	var output O

	if err := g.ensureDefaultConfig(); err != nil {
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

	// Create prompt with schema instructions
	prompt := fmt.Sprintf(
		`%s

Format your response according to this JSON schema, pay close attention to the validation rules in the schema:
%s

Provide the result enclosed in triple backticks with 'json' on the first line.
Don't put control characters in the wrong place or the JSON will be invalid.`,
		buf.String(), schema)

	// Run before hooks
	for _, hook := range g.hooks {
		if prompt, err = hook.BeforeRequest(ctx, prompt); err != nil {
			return output, fmt.Errorf("hook error: %w", err)
		}
	}

	// Call provider
	if g.provider == nil {
		return output, fmt.Errorf("no provider configured")
	}

	response, err := g.provider.Complete(ctx, provider.Request{
		Prompt:      prompt,
		Model:       g.model,
		Temperature: g.temp,
		MaxTokens:   g.maxTokens,
	})

	// Run after hooks
	for _, hook := range g.hooks {
		if response, err = hook.AfterResponse(ctx, response, err); err != nil {
			return output, fmt.Errorf("hook error: %w", err)
		}
	}

	// Extract JSON from markdown code blocks
	matches := jsonRegex.FindStringSubmatch(response)
	if len(matches) < 2 {
		return output, fmt.Errorf("no JSON found in response: %s", response)
	}

	jsonStr := matches[1]

	// Validate response
	if err := g.validator.Validate([]byte(jsonStr)); err != nil {
		return output, fmt.Errorf("invalid response: %w", err)
	}

	// Parse response into output type
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
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

// WithHook adds a hook to the generator
func (g *Generator[I, O]) WithHook(hook Hook) *Generator[I, O] {
	g.hooks = append(g.hooks, hook)
	return g
}

// Stream provides real-time streaming of the generated content
func (g *Generator[I, O]) Stream(ctx context.Context, input I) (*Stream, error) {
	if err := g.ensureDefaultConfig(); err != nil {
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
