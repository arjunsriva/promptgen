package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arjunsriva/promptgen"
)

// DocumentInput represents the raw input for document generation
type DocumentInput struct {
	RawText string `json:"raw_text"`
	Style   string `json:"style"`
}

// OutlineOutput represents the structured outline
type OutlineOutput struct {
	Sections []string `json:"sections" jsonschema:"required,minItems=1"`
	Topics   []string `json:"topics" jsonschema:"required,minItems=1"`
}

// ValidationOutput represents the outline validation results
type ValidationOutput struct {
	IsValid  bool     `json:"is_valid" jsonschema:"required"`
	Issues   []string `json:"issues" jsonschema:"required"`
	Coverage float64  `json:"coverage" jsonschema:"required,minimum=0,maximum=1"`
}

// FinalDocument represents the complete generated document
type FinalDocument struct {
	Title    string   `json:"title" jsonschema:"required"`
	Content  string   `json:"content" jsonschema:"required,minLength=100"`
	Keywords []string `json:"keywords" jsonschema:"required,minItems=3"`
}

// Chain of generators for each step
var (
	// Step 1: Generate outline from raw text
	generateOutline, _ = promptgen.Create[DocumentInput, OutlineOutput](`
		Create a structured outline for a document based on this text.
		The document style should be: {{.Style}}

		Text to analyze: {{.RawText}}

		Generate an outline with main sections and key topics to cover.
	`)

	// Step 2: Validate outline
	validateOutline, _ = promptgen.Create[OutlineOutput, ValidationOutput](`
		Validate this outline for completeness and coherence:

		Sections:
		{{range .Sections}}
		- {{.}}
		{{end}}

		Topics:
		{{range .Topics}}
		- {{.}}
		{{end}}

		Check for:
		1. Logical flow between sections
		2. Complete coverage of topics
		3. Appropriate depth of content
	`)

	// Step 3: Generate final document
	generateDocument, _ = promptgen.Create[OutlineOutput, FinalDocument](`
		Create a complete document based on this outline:

		Sections:
		{{range .Sections}}
		- {{.}}
		{{end}}

		Topics:
		{{range .Topics}}
		- {{.}}
		{{end}}

		Generate a coherent document that follows this structure.
	`)
)

func main() {
	// Configure timeout for the entire chain
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	input := DocumentInput{
		RawText: `AI agents are systems where LLMs dynamically direct their own processes
		          and tool usage. They maintain control over how they accomplish tasks.
		          When building applications with LLMs, find the simplest solution possible,
		          and only increase complexity when needed.`,
		Style: "technical blog post",
	}

	// Step 1: Generate outline
	outline, err := generateOutline.Run(ctx, input)
	if err != nil {
		log.Fatalf("Failed to generate outline: %v", err)
	}

	// Step 2: Validate outline
	validation, err := validateOutline.Run(ctx, outline)
	if err != nil {
		log.Fatalf("Failed to validate outline: %v", err)
	}

	// Check validation results
	if !validation.IsValid {
		log.Printf("Outline validation failed:")
		for _, issue := range validation.Issues {
			log.Printf("- %s", issue)
		}
		log.Printf("Coverage: %.2f%%", validation.Coverage*100)
		return
	}

	// Step 3: Generate final document
	document, err := generateDocument.Run(ctx, outline)
	if err != nil {
		log.Fatalf("Failed to generate document: %v", err)
	}

	// Output results
	fmt.Printf("\nGenerated Document:\n")
	fmt.Printf("Title: %s\n\n", document.Title)
	fmt.Printf("Content:\n%s\n\n", document.Content)
	fmt.Printf("Keywords: %v\n", document.Keywords)
}
