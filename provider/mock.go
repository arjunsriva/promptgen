package provider

import (
	"context"
	"sync"
	"time"
)

// MockProvider implements Provider interface for testing
type MockProvider struct {
	Response     string     // Fixed response for Complete
	StreamTokens []string   // Tokens to stream
	Errors       []error    // Errors to return
	Prompts      []string   // Captured prompts for verification
	DelayMs      int        // Optional delay to simulate network latency
	mu           sync.Mutex // Protects concurrent access to Prompts
}

func (m *MockProvider) Complete(ctx context.Context, prompt string) (string, error) {
	m.mu.Lock()
	m.Prompts = append(m.Prompts, prompt)
	m.mu.Unlock()

	if len(m.Errors) > 0 {
		err := m.Errors[0]
		m.Errors = m.Errors[1:]
		return "", err
	}

	if m.DelayMs > 0 {
		select {
		case <-time.After(time.Duration(m.DelayMs) * time.Millisecond):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	return m.Response, nil
}

func (m *MockProvider) Stream(ctx context.Context, prompt string) (contentChan <-chan string, errChan <-chan error, err error) {
	m.mu.Lock()
	m.Prompts = append(m.Prompts, prompt)
	m.mu.Unlock()

	content := make(chan string)
	errs := make(chan error, 1) // Buffer error channel to prevent blocking

	// Return immediate error if set
	if len(m.Errors) > 0 {
		err := m.Errors[0]
		m.Errors = m.Errors[1:]
		close(content)
		errs <- err
		close(errs)
		return content, errs, nil
	}

	go func() {
		defer close(content)
		defer close(errs)

		for _, token := range m.StreamTokens {
			if m.DelayMs > 0 {
				select {
				case <-time.After(time.Duration(m.DelayMs) * time.Millisecond):
				case <-ctx.Done():
					errs <- ctx.Err()
					return
				}
			}

			select {
			case content <- token:
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			}
		}
	}()

	return content, errs, nil
}
