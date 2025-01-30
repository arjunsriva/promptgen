// Package main demonstrates using promptgen to generate product copy with OpenAI
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/arjunsriva/promptgen"
	"github.com/arjunsriva/promptgen/provider"
)

// ProductInput represents the input data for product copy generation
type ProductInput struct {
	Name     string   `json:"name"`
	Features []string `json:"features"`
}

// ProductCopy represents the generated product copy with SEO constraints
type ProductCopy struct {
	Title       string `json:"title" jsonschema:"required,maxLength=60,description=SEO-friendly product title"`
	Description string `json:"description" jsonschema:"required,minLength=100,maxLength=160,description=Meta description for search results"`
}

var GenerateProductCopy, _ = promptgen.Create[ProductInput, ProductCopy](`
	Write product copy for {{.Name}}. 
	
	Key features:
	{{range .Features}}- {{.}}
	{{end}}

	Generate a title and description suitable for an e-commerce website.
`)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// You can configure the generator optionally, we default to OpenAI gpt-4o-mini
	GenerateProductCopy.WithProvider(provider.NewOpenAI(apiKey)).
		WithModel("gpt-3.5-turbo").
		WithTemperature(0.7)

	input := ProductInput{
		Name: "Ergonomic Office Chair",
		Features: []string{
			"Adjustable lumbar support",
			"4D armrests",
			"Breathable mesh back",
			"5-year warranty",
		},
	}

	copy, err := GenerateProductCopy.Run(context.Background(), input)
	if err != nil {
		switch {
		case promptgen.IsRateLimit(err):
			log.Fatal("Rate limit exceeded, please try again later")
		case promptgen.IsContextLength(err):
			log.Fatal("Input too long, please reduce the content")
		default:
			log.Fatalf("Failed to generate copy: %v", err)
		}
	}

	fmt.Printf("Title: %s\n", copy.Title)
	fmt.Printf("Description: %s\n", copy.Description)
}
