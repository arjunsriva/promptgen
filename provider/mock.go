package provider

import (
	"context"
	"time"
)

// MockProvider implements Provider interface for testing
type MockProvider struct {
	Response     string    // Fixed response for Complete
	StreamTokens []string  // Tokens to stream
	Errors       []error   // Errors to return
	Requests     []Request // Captured requests for verification
	DelayMs      int       // Optional delay to simulate network latency
}

func (m *MockProvider) Complete(_ context.Context, req Request) (string, error) {
	m.Requests = append(m.Requests, req)

	if len(m.Errors) > 0 {
		return "", m.Errors[0]
	}

	if m.DelayMs > 0 {
		time.Sleep(time.Duration(m.DelayMs) * time.Millisecond)
	}

	return m.Response, nil
}

func (m *MockProvider) Stream(_ context.Context, req Request) (<-chan string, <-chan error, error) {
	m.Requests = append(m.Requests, req)

	contentChan := make(chan string)
	errChan := make(chan error)

	if len(m.Errors) > 0 {
		return nil, nil, m.Errors[0]
	}

	go func() {
		defer close(contentChan)
		defer close(errChan)

		// Send all tokens
		for _, token := range m.StreamTokens {
			if m.DelayMs > 0 {
				time.Sleep(time.Duration(m.DelayMs) * time.Millisecond)
			}
			contentChan <- token
		}
		// Stream is complete after sending all tokens
	}()

	return contentChan, errChan, nil
}
