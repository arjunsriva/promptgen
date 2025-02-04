package promptgen

import (
	"bytes"
	"context"
	"fmt"
)

// Stream represents a real-time stream of content from the AI provider
type Stream struct {
	Content chan string
	Err     chan error
	Done    chan struct{}
}

// Stream provides real-time streaming of the generated content
func (g *Generator[I, O]) Stream(ctx context.Context, input I) (*Stream, error) {
	if err := g.ensureDefaultConfig(); err != nil {
		return nil, err
	}

	// Apply timeout if set
	if g.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.timeout)
		defer cancel()
	}

	var buf bytes.Buffer
	if err := g.prompt.Execute(&buf, input); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Wrap prompt with type-specific instructions
	wrappedPrompt := g.handler.WrapPrompt(buf.String())

	// Run before hooks
	for _, hook := range g.hooks {
		var err error
		wrappedPrompt, err = hook.BeforeRequest(ctx, wrappedPrompt)
		if err != nil {
			return nil, fmt.Errorf("hook error: %w", err)
		}
	}

	contentChan, errChan, err := g.provider.Stream(ctx, wrappedPrompt)
	if err != nil {
		return nil, fmt.Errorf("provider stream failed: %w", err)
	}

	stream := &Stream{
		Content: make(chan string),
		Err:     make(chan error, 1), // Buffer error channel to prevent blocking
		Done:    make(chan struct{}),
	}

	go func() {
		defer func() {
			close(stream.Content)
			close(stream.Err)
			close(stream.Done)
		}()

		for {
			select {
			case <-ctx.Done():
				// Always prioritize context cancellation
				stream.Err <- ctx.Err()
				return
			case content, ok := <-contentChan:
				if !ok {
					// Channel closed - check for errors
					select {
					case err := <-errChan:
						stream.Err <- err
						return
					default:
						// Only signal done if no errors
						stream.Done <- struct{}{}
						return
					}
				}

				// Process content through hooks
				for _, hook := range g.hooks {
					var err error
					content, err = hook.AfterResponse(ctx, content, nil)
					if err != nil {
						stream.Err <- fmt.Errorf("hook error: %w", err)
						return
					}
				}

				// Send processed content
				select {
				case stream.Content <- content:
				case <-ctx.Done():
					stream.Err <- ctx.Err()
					return
				}
			case err, ok := <-errChan:
				if !ok {
					continue
				}
				stream.Err <- err
				return
			}
		}
	}()

	return stream, nil
}
