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
	Requests     []Request  // Captured requests for verification
	DelayMs      int        // Optional delay to simulate network latency
	mu           sync.Mutex // Protects concurrent access to Requests
}

func (m *MockProvider) Complete(ctx context.Context, req Request) (string, error) {
	m.mu.Lock()
	m.Requests = append(m.Requests, req)
	m.mu.Unlock()

	if len(m.Errors) > 0 {
		return "", m.Errors[0]
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

func (m *MockProvider) Stream(ctx context.Context, req Request) (contentChan <-chan string, errChan <-chan error, err error) {
	m.mu.Lock()
	m.Requests = append(m.Requests, req)
	m.mu.Unlock()

	content := make(chan string)
	errs := make(chan error)
	contentChan = content
	errChan = errs

	if len(m.Errors) > 0 {
		return nil, nil, m.Errors[0]
	}

	go func() {
		defer close(content)
		defer close(errs)

		if m.DelayMs > 0 {
			select {
			case <-time.After(time.Duration(m.DelayMs) * time.Millisecond):
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			}
		}

		for _, token := range m.StreamTokens {
			select {
			case content <- token:
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			}
		}
	}()

	return contentChan, errChan, nil
}
