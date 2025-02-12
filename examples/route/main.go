package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/arjunsriva/promptgen"
)

// SupportQuery represents an incoming support request
type SupportQuery struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

// QueryType represents the classification of a support query
type QueryType struct {
	Category   string   `json:"category" jsonschema:"required,enum=general,enum=technical,enum=billing,enum=urgent"`
	Complexity string   `json:"complexity" jsonschema:"required,enum=simple,enum=medium,enum=complex"`
	Keywords   []string `json:"keywords" jsonschema:"required,minItems=1"`
	Confidence float64  `json:"confidence" jsonschema:"required,minimum=0,maximum=1"`
}

// Response represents the final response to the user
type Response struct {
	Answer     string   `json:"answer" jsonschema:"required,minLength=50"`
	NextSteps  []string `json:"next_steps" jsonschema:"required"`
	References []string `json:"references,omitempty"`
}

// Chain of generators for routing and handling
var (
	// Classifier determines the type and complexity of the query
	classifyQuery, _ = promptgen.Create[SupportQuery, QueryType](`
		Analyze this support query and classify it:
		User ID: {{.UserID}}
		Message: {{.Message}}

		Determine:
		1. The category (general, technical, billing, urgent)
		2. Complexity level (simple, medium, complex)
		3. Key topics/keywords
		4. Your confidence in this classification (0-1)

		Consider:
		- Technical queries involve code, APIs, or system issues
		- Billing queries relate to payments, subscriptions, or pricing
		- Urgent queries indicate system outages or blocking issues
		- General queries are informational or don't fit other categories
	`)

	// General query handler
	handleGeneral, _ = promptgen.Create[SupportQuery, Response](`
		Respond to this general inquiry:
		{{.Message}}

		Provide a helpful, friendly response with clear next steps.
	`)

	// Technical query handler
	handleTechnical, _ = promptgen.Create[SupportQuery, Response](`
		Respond to this technical query:
		{{.Message}}

		Provide:
		1. A technical explanation
		2. Specific steps to resolve the issue
		3. Relevant documentation references
	`)

	// Billing query handler
	handleBilling, _ = promptgen.Create[SupportQuery, Response](`
		Address this billing-related query:
		{{.Message}}

		Provide:
		1. Clear explanation of billing policies
		2. Steps to resolve any payment issues
		3. Links to relevant billing documentation
	`)

	// Urgent query handler
	handleUrgent, _ = promptgen.Create[SupportQuery, Response](`
		Address this urgent support request:
		{{.Message}}

		Provide:
		1. Immediate steps for mitigation
		2. Clear escalation path
		3. Status update requirements
	`)
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Example support queries
	queries := []SupportQuery{
		{
			Message: "How do I reset my password?",
			UserID:  "user123",
		},
		{
			Message: "My API calls are returning 429 errors and my production system is down",
			UserID:  "user456",
		},
		{
			Message: "I was charged twice for my subscription this month",
			UserID:  "user789",
		},
	}

	for i, query := range queries {
		fmt.Printf("\nProcessing Query %d: %s\n", i+1, query.Message)

		// Step 1: Classify the query
		classification, err := classifyQuery.Run(ctx, query)
		if err != nil {
			log.Printf("Failed to classify query: %v", err)
			continue
		}

		fmt.Printf("Classification: %s (Complexity: %s, Confidence: %.2f)\n",
			classification.Category, classification.Complexity, classification.Confidence)

		// Step 2: Route to appropriate handler based on classification
		var response Response

		switch classification.Category {
		case "general":
			response, err = handleGeneral.Run(ctx, query)
		case "technical":
			response, err = handleTechnical.Run(ctx, query)
		case "billing":
			response, err = handleBilling.Run(ctx, query)
		case "urgent":
			response, err = handleUrgent.Run(ctx, query)
		default:
			log.Printf("Unknown category: %s", classification.Category)
			continue
		}

		if err != nil {
			log.Printf("Failed to handle query: %v", err)
			continue
		}

		// Step 3: Output the response
		fmt.Printf("\nResponse:\n%s\n", response.Answer)
		fmt.Printf("\nNext Steps:\n")
		for _, step := range response.NextSteps {
			fmt.Printf("- %s\n", step)
		}
		if len(response.References) > 0 {
			fmt.Printf("\nReferences:\n")
			for _, ref := range response.References {
				fmt.Printf("- %s\n", ref)
			}
		}
		fmt.Println(strings.Repeat("-", 80))
	}
}
