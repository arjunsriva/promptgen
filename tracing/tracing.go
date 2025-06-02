package tracing

import (
	// "context" // No longer needed for Shutdown
	// "log"     // No longer needed for Shutdown or init logging

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

func init() {
	// Initialize the global tracer. It will use the globally registered
	// TracerProvider or a no-op provider if none is set by an application.
	Tracer = otel.Tracer("github.com/arjunsriva/promptgen")
}

// NewTracerProvider creates a new OpenTelemetry TracerProvider configured
// with a stdout exporter. This function is intended to be called by
// an application to obtain a tracer provider. The application is responsible for
// registering this provider globally (e.g., using otel.SetTracerProvider(tp))
// and shutting it down when appropriate (e.g., defer tp.Shutdown(context.Background())).
func NewTracerProvider() (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	// Example: Configure with a batcher and a sampler.
	// Adjust sampler as needed (e.g., sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))) for production.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Using AlwaysSample for dev; consider ParentBased for prod.
	)
	return tp, nil
}

// Shutdown function is removed. The application is responsible for managing the lifecycle
// of the TracerProvider it creates and registers.
/*
func Shutdown(ctx context.Context, provider *sdktrace.TracerProvider) {
	if err := provider.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down tracer provider: %v", err)
	}
}
*/

// The previous init() that called NewTracerProvider and otel.SetTracerProvider
// is removed as the library should not globally register a provider by default.
/*
func init() {
	_, err := NewTracerProvider() // This also called otel.SetTracerProvider previously
	if err != nil {
		log.Fatalf("Failed to create tracer provider: %v", err)
	}
}
*/
