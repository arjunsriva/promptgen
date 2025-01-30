package hooks

import (
	"context"
	"log"
)

// LoggingHook implements logging functionality
type LoggingHook struct {
	Logger *log.Logger
}

func NewLoggingHook(logger *log.Logger) *LoggingHook {
	if logger == nil {
		logger = log.Default()
	}
	return &LoggingHook{Logger: logger}
}

func (h *LoggingHook) BeforeRequest(_ context.Context, prompt string) (string, error) {
	h.Logger.Printf("Sending prompt to provider:\n%s\n", prompt)
	return prompt, nil
}

func (h *LoggingHook) AfterResponse(_ context.Context, response string, err error) (string, error) {
	if err != nil {
		h.Logger.Printf("Provider error: %v\n", err)
		return response, err
	}
	h.Logger.Printf("Received response from provider:\n%s\n", response)
	return response, nil
}
