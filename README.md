# promptgen

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/arjunsriva/promptgen)
[![Go Version](https://img.shields.io/github/go-mod/go-version/arjunsriva/promptgen)](https://github.com/arjunsriva/promptgen)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)


### Build production-ready AI applications in Go with type safety and ease.

Promptgen makes building AI-powered applications in Go productive and enjoyable, while maintaining type safety and performance. It combines Go's strengths with an elegant API that just works.

### At a glance

- 🎯 Type-safe inputs and outputs with Go templates and JSON Schema
- 🔄 Context-aware streaming with Go channels
- 💾 Thread-safe state management
- 🔌 Provider agnostic - switch AI models with a single line
- ⚡️ Production-ready with proper error handling

## Installation

```bash
go get github.com/arjunsriva/promptgen
```

## Quick Examples

### Product Copy Generation

Generate SEO-optimized product titles and descriptions with proper constraints:

```go
type ProductInput struct {
    Name         string   `json:"name"`
    Category     string   `json:"category"`
    MainFeatures []string `json:"main_features"`
}

type ProductCopy struct {
    Title       string `json:"title" jsonschema:"required,maxLength=60,description=SEO title length"`
    Description string `json:"description" jsonschema:"required,minLength=50,maxLength=160,description=Meta description length"`
}

copy := promptgen.Create[ProductInput, ProductCopy](
    `Create a product title and description for {{.Name}} in the {{.Category}} category.
     Key features:
     {{range .MainFeatures}}- {{.}}
     {{end}}`)

result, err := copy.Run(context.Background(), ProductInput{
    Name:     "Ultra Comfort Ergonomic Chair",
    Category: "Office Furniture",
    MainFeatures: []string{
        "Adjustable lumbar support",
        "4D armrests",
        "Breathable mesh back",
    },
})
```

### Review Summary

Demonstrate conditional templates and enumerated outputs:

```go
type ReviewInput struct {
    Content    string `json:"content"`
    MaxLength  int    `json:"max_length"`
}

type ReviewSummary struct {
    Summary    string `json:"summary" jsonschema:"required,maxLength=150"`
    Sentiment  string `json:"sentiment" jsonschema:"required,enum=positive,enum=negative,enum=neutral"`
}

summarize := promptgen.Create[ReviewInput, ReviewSummary](
    `Summarize this review:
     {{.Content}}
     {{if .MaxLength}}Keep the summary under {{.MaxLength}} characters.{{end}}`)

result, err := summarize.Run(context.Background(), ReviewInput{
    Content:   "Long customer review text...",
    MaxLength: 100,
})
```

## Core Features

### Type-Safe Templates

Promptgen uses Go's built-in text/template for prompt creation:

- Full access to template features:
  - Conditionals: `{{if .Condition}}...{{end}}`
  - Loops: `{{range .Items}}...{{end}}`
  - Built-in functions: `{{.Text | lower}}`
- Compile-time checking of template syntax
- Type-safe access to input fields

### JSON Schema Validation

Promptgen uses struct tags to define validation rules for AI outputs:

```go
type Output struct {
    // String validations
    Title       string `json:"title" jsonschema:"required,minLength=3,maxLength=60,pattern=^[A-Za-z].*"`
    
    // Numeric validations
    Score       float64 `json:"score" jsonschema:"minimum=0,maximum=100,multipleOf=0.5"`
    
    // Array validations
    Tags        []string `json:"tags" jsonschema:"minItems=1,maxItems=10,uniqueItems=true"`
    
    // Enum validation
    Status      string   `json:"status" jsonschema:"enum=pending,enum=active,enum=completed"`
    
    // Optional fields
    Metadata    map[string]string `json:"metadata,omitempty" jsonschema:"additionalProperties=true"`
}
```

Available validations:
- Strings: `minLength`, `maxLength`, `pattern`
- Numbers: `minimum`, `maximum`, `multipleOf`
- Arrays: `minItems`, `maxItems`, `uniqueItems`
- Objects: `required`, `additionalProperties`
- Enums: multiple `enum=value` tags
- Common: `description` for documentation

### Real-Time Streaming

Use Go's channels for efficient stream processing:

```go
stream, err := generator.Stream(ctx, input)
if err != nil {
    panic(err)
}

for {
    select {
    case chunk := <-stream.Content:
        fmt.Print(chunk)
    case err := <-stream.Err:
        panic(err)
    case <-stream.Done:
        return
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### Thread-Safe State Management

Built-in conversation memory with customizable storage:

```go
type RedisState struct {
    client *redis.Client
}

func (r *RedisState) Store(ctx context.Context) promptgen.StateStore[[]byte] {
    return &RedisStore{client: r.client}
}

chat := promptgen.Create[ChatInput, string]("{message}").
    WithState(&RedisState{client: redisClient})
```

### Provider Agnostic

Switch between AI providers with a single line:

```go
// OpenAI
generator.WithModel("gpt-4")

// Anthropic
generator.WithModel("claude-3-opus-20240229")

// Local
generator.WithModel("local/mistral-7b")
```

## Best Practices

### Error Handling

```go
result, err := generator.Run(ctx, input)
if err != nil {
    switch {
    case promptgen.IsRateLimit(err):
        // Handle rate limiting
    case promptgen.IsContextTooLong(err):
        // Handle context length
    case promptgen.IsInvalidRequest(err):
        // Handle invalid inputs
    default:
        // Handle other errors
    }
}
```

### Resource Management

```go
// Always use contexts for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Clean up resources
defer stream.Close()
```

### Configuration

```go
generator := promptgen.Create[Input, Output](prompt).
    WithModel("gpt-4").
    WithTemperature(0.7).
    WithMaxTokens(1000).
    WithRetry(retry.Exponential(3))
```

## Authentication

Multiple ways to handle authentication:

```go
// 1. Environment variables (recommended)
export OPENAI_API_KEY=sk-...
export ANTHROPIC_API_KEY=sk-ant-...

// 2. Direct configuration
generator.WithAuth(promptgen.Auth{
    Provider: promptgen.OpenAI,
    APIKey:   "sk-...",
})

// 3. Configuration file
generator.WithConfigFile("config.yaml")
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md)

## License

Apache 2.0 - See [LICENSE](./LICENSE) for more details.

## Acknowledgments

This project was inspired by [promptic](https://github.com/knowsuchagency/promptic), which showed how productive AI development could be in Python. I've built on that vision to create an idiomatic, type-safe Go experience.