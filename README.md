# promptgen

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/arjunsriva/promptgen)
[![Go Version](https://img.shields.io/github/go-mod/go-version/arjunsriva/promptgen)](https://github.com/arjunsriva/promptgen)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Build production-ready AI applications in Go with type safety and ease.

## Features

- 🎯 Type-safe inputs and outputs with Go generics and JSON Schema validation
- 🔄 Real-time streaming with Go channels
- 🪝 Extensible hook system for pre/post processing
- ⛓️ Support for chaining operations
- 🔌 Provider agnostic with built-in OpenAI support
- 🧪 Comprehensive testing utilities with mock provider
- 📡 OpenTelemetry support for tracing

## Installation

```bash
go get github.com/arjunsriva/promptgen
```

## Quick Start

### Simple Types

Work directly with Go's basic types:

```go
// String generation
stringGen, _ := promptgen.Create[string, string]("Tell me a {{.}} joke", "jokeGenerator")
joke, _ := stringGen.Run(ctx, "Dad")

// Integer estimation
intGen, _ := promptgen.Create[string, int]("Guess the age: {{.}}", "ageGuesser")
age, _ := intGen.Run(ctx, "college professor with grey hair")

// Float conversion
floatGen, _ := promptgen.Create[float64, float64]("Convert {{.}} Fahrenheit to Celsius", "tempConverter")
celsius, _ := floatGen.Run(ctx, 98.6)
```

### Structured Data

Define type-safe inputs and outputs with JSON Schema validation:

```go
type ProductInput struct {
    Name     string   `json:"name"`
    Features []string `json:"features"`
}

type ProductCopy struct {
    Title       string `json:"title" jsonschema:"required,maxLength=60"`
    Description string `json:"description" jsonschema:"required,maxLength=160"`
}

generator, _ := promptgen.Create[ProductInput, ProductCopy](`
    Write product copy for {{.Name}}.
    Features:
    {{range .Features}}- {{.}}
    {{end}}
`, "productCopyGenerator")

result, err := generator.Run(ctx, ProductInput{
    Name: "Ergonomic Chair",
    Features: []string{"Adjustable height", "Lumbar support"},
})
```

### Real-Time Streaming

Process responses in real-time using Go channels:

```go
// Assuming generator is created with an operationName, e.g.:
// generator, _ := promptgen.Create[InputType, OutputType]("Streaming prompt {{.}}", "myStreamer")
stream, _ := generator.Stream(ctx, input) // Stream method will also be instrumented

for {
    select {
    case chunk := <-stream.Content:
        fmt.Print(chunk)
    case err := <-stream.Err:
        handleError(err)
    case <-stream.Done:
        return
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### Operation Chaining

Build complex workflows by chaining operations:

```go
// Define chain of operations
var (
    classifyQuery, _ = promptgen.Create[Query, Classification](
        "Classify this query: {{.Text}}", "queryClassifier")

    generateResponse, _ = promptgen.Create[Classification, Response](
        "Generate response for {{.Category}} query", "responseGenerator")
)

// Execute chain
classification, _ := classifyQuery.Run(ctx, query)
response, _ := generateResponse.Run(ctx, classification)
```

### Hook System

Add pre/post processing hooks for logging, metrics, or transformations:

```go
type LoggingHook struct {
    logger *log.Logger
}

func (h *LoggingHook) BeforeRequest(ctx context.Context, prompt string) (string, error) {
    h.logger.Printf("Sending prompt: %s", prompt)
    return prompt, nil
}

func (h *LoggingHook) AfterResponse(ctx context.Context, response string, err error) (string, error) {
    h.logger.Printf("Got response: %s", response)
    return response, err
}

// Assuming generator is created with an operationName
generator.WithHook(&LoggingHook{logger: log.Default()})
```

### Provider Interface

Switch between providers or implement your own:

```go
// Assuming generator is created with an operationName
// Use OpenAI
generator.WithProvider(provider.NewOpenAI(provider.OpenAIConfig{
    Model: "gpt-4",
    Temperature: 0.7,
}))

// Use mock provider for testing
generator.WithProvider(&provider.MockProvider{
    Response: "mocked response",
})
```

## Advanced Examples

Check out the [examples](./examples) directory for more complex use cases:

- [Chain Operations](./examples/chain/main.go) - Sequential processing
- [Support Routing](./examples/route/main.go) - Query classification and routing
- [Parallel Processing](./examples/parallel/main.go) - Concurrent operations
- [Content Evaluation](./examples/eval/main.go) - Content moderation
- [Translation](./examples/translate/main.go) - Language translation
- [**Tracing Setup**](./examples/tracing/main.go) - How to initialize OpenTelemetry in your application.

## Error Handling

```go
result, err := generator.Run(ctx, input)
if err != nil {
    switch {
    case errors.Is(err, promptgen.ErrRateLimit):
        // Handle rate limiting
    case errors.Is(err, promptgen.ErrContextLength):
        // Handle context length
    case errors.Is(err, promptgen.ErrValidation):
        // Handle validation errors
    default:
        // Handle other errors
    }
}
```

## Testing

Use the mock provider for reliable testing:

```go
mockProvider := &provider.MockProvider{
    Response: `{"title": "Test Title", "description": "Test Description"}`,
}
// Assuming generator is created with an operationName
generator.WithProvider(mockProvider)
result, err := generator.Run(ctx, input)
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup and guidelines.

## License

Apache 2.0 - See [LICENSE](./LICENSE) for details.

## OpenTelemetry Tracing

The `promptgen` library is instrumented with OpenTelemetry. This means its core operations, such as `Generator.Run()`, and calls to underlying AI providers (like `OpenAI.Complete` if the OpenAI provider is used), will create trace spans. These spans provide insights into the execution flow and performance of your AI interactions.

However, `promptgen` **does not initialize or configure any specific OpenTelemetry exporter or the global TracerProvider itself.** It is the application's responsibility to set up the desired OpenTelemetry environment. If an application using `promptgen` does not initialize and register a `TracerProvider`, the tracing calls within `promptgen` will delegate to the default no-op tracer provider from the OpenTelemetry SDK, meaning no traces will be collected or exported.

### Enabling and Configuring Tracing (Application Setup)

To collect and export traces generated by `promptgen` (and your own application code), your application needs to:

1.  **Import necessary OpenTelemetry packages:**
    *   Your chosen exporter package(s) (e.g., `go.opentelemetry.io/otel/exporters/stdout/stdouttrace` for console output, or exporters for Jaeger, Zipkin, OTLP).
    *   OpenTelemetry SDK: `go.opentelemetry.io/otel` (for global operations like `SetTracerProvider`) and `go.opentelemetry.io/otel/sdk/trace` (usually aliased as `sdktrace` for creating a `TracerProvider`).
2.  **Create an Exporter Instance:** Configure and instantiate your chosen trace exporter.
3.  **Create a `TracerProvider`:** Instantiate `sdktrace.NewTracerProvider()` and configure it with your exporter. You can also set up other aspects like a `Sampler` (e.g., `sdktrace.AlwaysSample()`, `sdktrace.TraceIDRatioBased()`) and resource attributes.
4.  **Register Globally:** Register your created `TracerProvider` as the global one using `otel.SetTracerProvider(yourTracerProvider)`. `promptgen` will use this global provider.
5.  **Shutdown Cleanly:** Ensure your `TracerProvider` is shut down when your application exits to flush any buffered traces. This is typically done using `defer yourTracerProvider.Shutdown(context.Background())`.

**For a concrete example** of how to set up OpenTelemetry in your application to receive traces from `promptgen`, please see [`examples/tracing/main.go`](./examples/tracing/main.go). This example demonstrates a basic setup using a stdout exporter.

The `promptgen/tracing` package also provides a utility function `tracing.NewTracerProvider()`. This function creates a basic `*sdktrace.TracerProvider` pre-configured with a stdout exporter and an "always sample" sampler. Applications *can* use this as a quick starting point for development or simple console-based tracing, but for production scenarios, a more comprehensive setup with appropriate exporters and sampling is recommended. If you use this utility, your application is still responsible for globally registering the returned provider and managing its shutdown.

### Span Creation and Attributes

When you create a `Generator` instance, you can provide an `operationName` string:
```go
generator, _ := promptgen.Create[InputType, OutputType]("Your prompt {{.}}", "yourCustomOperationName")
```
This `operationName` is used by `promptgen` to create a parent span for the entire `Generator.Run()` operation (and eventually `Generator.Stream()` once instrumented). This helps in contextualizing the specific business logic being executed. If no `operationName` is provided, a default name like "promptgen.Run" will be used.

This parent operation span (e.g., "yourCustomOperationName" or "promptgen.Run") will include the following attribute if a custom operation name was provided:
- `promptgen.operation_name`: The custom name given to the generator operation (e.g., "yourCustomOperationName").

Provider-specific spans (e.g., `OpenAI.Complete` or `OpenAI.Stream`) are created by the respective provider methods. If the context is passed correctly through `promptgen` to these provider methods, these provider spans will appear as children of the main operation span. These provider-level child spans typically include attributes like:
    - `llm.provider`: The name of the provider (e.g., "OpenAI").
    - `llm.model_name`: The model being used by the provider.
    - `llm.temperature`: The temperature setting for the LLM call.
    - `llm.max_tokens`: The max tokens setting for the LLM call.
    - For `OpenAI.Complete` calls, token usage is also included:
        - `llm.usage.prompt_tokens`
        - `llm.usage.completion_tokens`
        - `llm.usage.total_tokens`

### Viewing Traces

How you view traces depends on the exporter you configure in your application:
- With the `stdouttrace` exporter (as shown in `examples/tracing/main.go`), traces will be printed to your application's standard output.
- If you configure an exporter for a system like Jaeger or Zipkin, you will use their respective UIs to view traces.

## Acknowledgments

This project was inspired by [promptic](https://github.com/knowsuchagency/promptic), which showed how productive AI development could be in Python. I've built on that vision to create an idiomatic, type-safe Go experience.
