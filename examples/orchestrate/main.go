package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arjunsriva/promptgen"
)

// ResearchQuery represents a complex question needing investigation
type ResearchQuery struct {
	Question string   `json:"question"`
	Context  string   `json:"context"`
	Topics   []string `json:"topics,omitempty"`
}

// ResearchPlan represents how the orchestrator breaks down the query
type ResearchPlan struct {
	SubQuestions []SubQuestion `json:"sub_questions" jsonschema:"required,minItems=1"`
	Approach     string        `json:"approach" jsonschema:"required"`
}

// SubQuestion represents a specific aspect to investigate
type SubQuestion struct {
	Question    string   `json:"question" jsonschema:"required"`
	AspectType  string   `json:"aspect_type" jsonschema:"required,enum=technical,enum=market,enum=user,enum=trend"`
	KeyPoints   []string `json:"key_points" jsonschema:"required"`
	DataSources []string `json:"data_sources" jsonschema:"required"`
}

// Finding represents a worker's research on a specific sub-question
type Finding struct {
	Summary     string   `json:"summary" jsonschema:"required"`
	Evidence    []string `json:"evidence" jsonschema:"required"`
	Confidence  float64  `json:"confidence" jsonschema:"required,minimum=0,maximum=1"`
	Limitations []string `json:"limitations,omitempty"`
}

var (
	// Orchestrator breaks down the research query
	planResearch, _ = promptgen.Create[ResearchQuery, ResearchPlan](`
		Break down this research question into specific aspects to investigate:
		Question: {{.Question}}
		Context: {{.Context}}
		{{if .Topics}}
		Related Topics: {{range .Topics}}
		- {{.}}
		{{end}}
		{{end}}

		For each aspect:
		1. Formulate specific sub-questions
		2. Identify key points to investigate
		3. Suggest relevant data sources
		4. Determine the type of analysis needed
	`)

	// Worker investigates a specific sub-question
	investigate, _ = promptgen.Create[struct {
		Original ResearchQuery
		SubQ     SubQuestion
	}, Finding](`
		Investigate this aspect of the research:

		Original Question: {{.Original.Question}}
		Sub-Question: {{.SubQ.Question}}

		Focus on these key points:
		{{range .SubQ.KeyPoints}}
		- {{.}}
		{{end}}

		Consider these sources:
		{{range .SubQ.DataSources}}
		- {{.}}
		{{end}}

		Provide:
		1. A detailed summary of findings
		2. Supporting evidence
		3. Confidence level in conclusions
		4. Any limitations or caveats
	`)

	// Orchestrator synthesizes findings
	synthesize, _ = promptgen.Create[struct {
		Query    ResearchQuery
		Plan     ResearchPlan
		Findings map[string]Finding
	}, struct {
		Conclusion  string   `json:"conclusion" jsonschema:"required"`
		KeyInsights []string `json:"key_insights" jsonschema:"required"`
		Gaps        []string `json:"gaps,omitempty"`
	}](`
		Synthesize the research findings:

		Original Question: {{.Query.Question}}

		Findings by aspect:
		{{range $idx, $subq := .Plan.SubQuestions}}
		{{$finding := index $.Findings $subq.Question}}
		Aspect: {{$subq.AspectType}}
		Summary: {{$finding.Summary}}
		Confidence: {{$finding.Confidence}}
		{{end}}

		Provide:
		1. A comprehensive conclusion
		2. Key insights across all aspects
		3. Identify any remaining gaps
	`)
)

func conductResearch(ctx context.Context, query ResearchQuery) error {
	// Step 1: Orchestrator plans the research
	plan, err := planResearch.Run(ctx, query)
	if err != nil {
		return fmt.Errorf("planning failed: %w", err)
	}

	fmt.Printf("Research Plan:\n")
	fmt.Printf("Approach: %s\n\n", plan.Approach)
	for i, sq := range plan.SubQuestions {
		fmt.Printf("Sub-Question %d (%s):\n", i+1, sq.AspectType)
		fmt.Printf("Q: %s\n", sq.Question)
		fmt.Printf("Key Points: %v\n", sq.KeyPoints)
		fmt.Printf("Sources: %v\n\n", sq.DataSources)
	}

	// Step 2: Workers investigate each aspect
	findings := make(map[string]Finding)
	for _, subq := range plan.SubQuestions {
		finding, err := investigate.Run(ctx, struct {
			Original ResearchQuery
			SubQ     SubQuestion
		}{query, subq})
		if err != nil {
			return fmt.Errorf("investigation failed for '%s': %w", subq.Question, err)
		}

		findings[subq.Question] = finding
		fmt.Printf("\nFindings for: %s\n", subq.Question)
		fmt.Printf("Summary: %s\n", finding.Summary)
		fmt.Printf("Confidence: %.2f\n", finding.Confidence)
	}

	// Step 3: Orchestrator synthesizes findings
	synthesis, err := synthesize.Run(ctx, struct {
		Query    ResearchQuery
		Plan     ResearchPlan
		Findings map[string]Finding
	}{query, plan, findings})
	if err != nil {
		return fmt.Errorf("synthesis failed: %w", err)
	}

	fmt.Printf("\nFinal Synthesis:\n")
	fmt.Printf("Conclusion: %s\n\n", synthesis.Conclusion)
	fmt.Printf("Key Insights:\n")
	for _, insight := range synthesis.KeyInsights {
		fmt.Printf("- %s\n", insight)
	}
	if len(synthesis.Gaps) > 0 {
		fmt.Printf("\nRemaining Gaps:\n")
		for _, gap := range synthesis.Gaps {
			fmt.Printf("- %s\n", gap)
		}
	}

	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	query := ResearchQuery{
		Question: "What are the current challenges and opportunities in implementing " +
			"large language models for real-time customer support applications?",
		Context: "Enterprise SaaS company considering LLM integration",
		Topics: []string{
			"Technical requirements",
			"Cost implications",
			"User experience",
			"Industry trends",
		},
	}

	if err := conductResearch(ctx, query); err != nil {
		log.Fatalf("Research failed: %v", err)
	}
}
