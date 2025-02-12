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

## Installation

```bash
go get github.com/arjunsriva/promptgen
```

## Quick Start

### Simple Types

Work directly with Go's basic types:

```go
// String generation
stringGen, _ := promptgen.Create[string, string]("Tell me a {{.}} joke")
joke, _ := stringGen.Run(ctx, "Dad")

// Integer estimation
intGen, _ := promptgen.Create[string, int]("Guess the age: {{.}}")
age, _ := intGen.Run(ctx, "college professor with grey hair")

// Float conversion
floatGen, _ := promptgen.Create[float64, float64]("Convert {{.}} Fahrenheit to Celsius")
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
`)

result, err := generator.Run(ctx, ProductInput{
    Name: "Ergonomic Chair",
    Features: []string{"Adjustable height", "Lumbar support"},
})
```

### Real-Time Streaming

Process responses in real-time using Go channels:

```go
stream, _ := generator.Stream(ctx, input)

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
        "Classify this query: {{.Text}}")

    generateResponse, _ = promptgen.Create[Classification, Response](
        "Generate response for {{.Category}} query")
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

generator.WithHook(&LoggingHook{logger: log.Default()})
```

### Provider Interface

Switch between providers or implement your own:

```go
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

generator.WithProvider(mockProvider)
result, err := generator.Run(ctx, input)
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup and guidelines.

## License

Apache 2.0 - See [LICENSE](./LICENSE) for details.

## Acknowledgments

This project was inspired by [promptic](https://github.com/knowsuchagency/promptic), which showed how productive AI development could be in Python. I've built on that vision to create an idiomatic, type-safe Go experience.
