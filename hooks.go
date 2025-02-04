package promptgen

import "context"

// Hook represents a function that can intercept and modify requests/responses
type Hook interface {
	BeforeRequest(ctx context.Context, prompt string) (string, error)
	AfterResponse(ctx context.Context, response string, err error) (string, error)
}
