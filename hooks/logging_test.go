package hooks

import (
	"bytes"
	"context"
	"errors"
	"log"
	"testing"
)

func TestLoggingHook(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	hook := NewLoggingHook(logger)
	ctx := context.Background()

	// Test BeforeRequest
	t.Run("BeforeRequest", func(t *testing.T) {
		buf.Reset()
		hook.BeforeRequest(ctx, "test prompt")
		got := buf.String()
		want := "Sending prompt to provider:\ntest prompt\n"
		if got != want {
			t.Errorf("BeforeRequest() logged %q, want %q", got, want)
		}
	})

	// Test AfterResponse with error
	t.Run("AfterResponse with error", func(t *testing.T) {
		buf.Reset()
		testErr := errors.New("test error")
		hook.AfterResponse(ctx, "test response", testErr)
		got := buf.String()
		want := "Provider error: test error\n"
		if got != want {
			t.Errorf("AfterResponse() logged %q, want %q", got, want)
		}
	})

	// Test AfterResponse without error
	t.Run("AfterResponse without error", func(t *testing.T) {
		buf.Reset()
		hook.AfterResponse(ctx, "test response", nil)
		got := buf.String()
		want := "Received response from provider:\ntest response\n"
		if got != want {
			t.Errorf("AfterResponse() logged %q, want %q", got, want)
		}
	})

	// Test with nil logger
	t.Run("nil logger", func(t *testing.T) {
		hook := NewLoggingHook(nil)
		// These should not panic
		hook.BeforeRequest(ctx, "test")
		hook.AfterResponse(ctx, "test", nil)
	})
}
