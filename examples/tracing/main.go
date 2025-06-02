package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os" // Required for API key check

	"github.com/arjunsriva/promptgen"
	"github.com/arjunsriva/promptgen/provider" // For provider.ErrRateLimit etc.
	"github.com/arjunsriva/promptgen/tracing"  // For the app-side NewTracerProvider

	"go.opentelemetry.io/otel" // For otel.SetTracerProvider
	// sdktrace "go.opentelemetry.io/otel/sdk/trace" // Not directly needed if using tracing.NewTracerProvider
)

func main() {
	// 1. Initialize OpenTelemetry Tracer Provider (Application's responsibility)
	// This uses the utility from promptgen's tracing package, but an application
	// could configure any exporter and provider setup it needs here.
	tp, err := tracing.NewTracerProvider()
	if err != nil {
		log.Fatalf("Failed to initialize tracer provider: %v", err)
	}

	// 2. Register the TracerProvider globally
	otel.SetTracerProvider(tp)

	// 3. Defer shutdown of the TracerProvider
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown tracer provider: %v", err)
		}
	}()

	// Check for OPENAI_API_KEY (example specific, not for library use)
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Println("OPENAI_API_KEY not set. Traces will be generated for promptgen operations,")
		log.Println("but OpenAI provider calls will fail if used without a key.")
		// Depending on how promptgen.Create handles default provider errors without Run,
		// we might not see OpenAI spans if it errors out early.
		// Forcing a specific provider or handling this more gracefully might be needed for a full demo.
	}

	ctx := context.Background()

	// Example: String type - Joke generator
	// The "jokeGenerator" name will be used for the parent span.
	stringGen, err := promptgen.Create[string, string]("Tell me a {{.}} joke, be creative, unusual", "jokeGenerator")
	if err != nil {
		log.Fatalf("Failed to create string generator: %v", err)
	}

	// Example of using a specific provider (optional, good for testing specific instrumentation)
	// If OPENAI_API_KEY is not set, DefaultOpenAI() in Create will error if not handled.
	// For this example, let's assume it might be set, or we want to see the trace up to that point.
	// openAIProvider, err := provider.DefaultOpenAI()
	// if err == nil {
	// 	stringGen.WithProvider(openAIProvider)
	// } else {
	// 	log.Printf("Could not initialize default OpenAI provider (API key likely missing): %v", err)
	//  log.Println("Proceeding without an explicit provider set for stringGen, OpenAI calls will fail if made.")
	// }


	theme := "developer"
	log.Printf("Requesting a %s joke...", theme)
	// When stringGen.Run is called, it will use the globally set tracer.
	// Spans "jokeGenerator" and then "OpenAI.Complete" (if called) should be printed to stdout.
	joke, err := stringGen.Run(ctx, theme)
	if err != nil {
		handleError(err) // Use a common error handler
	} else {
		fmt.Printf(`
Joke for %s:
%s
`, theme, joke)
	}

	// Add another example to see multiple distinct traces
	// Example: Integer type - Age guesser
	intGen, err := promptgen.Create[string, int]("Given the description '{{.}}', guess the person's age", "ageGuesser")
	if err != nil {
		log.Fatalf("Failed to create int generator: %v", err)
	}
	description := "a seasoned Go developer who loves distributed systems"
	log.Printf(`
Requesting age guess for: %s`, description)
	age, err := intGen.Run(ctx, description)
	if err != nil {
		handleError(err)
	} else {
		fmt.Printf(`
Age Guesser - Description: %s
Estimated age: %d years
`, description, age)
	}

	log.Println(`
Application finished. Traces (if any) should have been printed to stdout.`)
	log.Println("If OPENAI_API_KEY was not set, OpenAI provider calls would have failed,")
	log.Println("but spans for 'jokeGenerator' and 'ageGuesser' should still appear.")
}

// handleError is a utility function to demonstrate error handling with custom errors.
func handleError(err error) {
	// Using errors.Is for specific error types from promptgen or provider
	if errors.Is(err, promptgen.ErrRateLimit) || errors.Is(err, provider.ErrRateLimit) {
		log.Printf("Error: Rate limit exceeded. Please try again later. Details: %v", err)
	} else if errors.Is(err, promptgen.ErrContextLength) || errors.Is(err, provider.ErrContextLength) {
		log.Printf("Error: Input too long. Please reduce the content length. Details: %v", err)
	} else if errors.Is(err, promptgen.ErrTimeout) {
		log.Printf("Error: Request timed out. Details: %v", err)
	} else if errors.Is(err, promptgen.ErrInvalidResponse) {
		var e *promptgen.Error
		if errors.As(err, &e) {
			log.Printf("Error: Invalid response from AI - Code: %s, Message: %s, Details: %v", e.Code, e.Message, e.Err)
		} else {
			log.Printf("Error: Invalid response from AI. Details: %v", err)
		}
	} else if errors.Is(err, promptgen.ErrValidation) {
		var e *promptgen.Error
		if errors.As(err, &e) {
			log.Printf("Error: Response validation failed - Code: %s, Message: %s, Details: %v", e.Code, e.Message, e.Err)
		} else {
			log.Printf("Error: Response validation failed. Details: %v", err)
		}
	} else {
		// General error
		log.Fatalf("Operation failed: %v", err)
	}
}
