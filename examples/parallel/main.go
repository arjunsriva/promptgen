package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/arjunsriva/promptgen"
)

// ContentInput represents the content to be moderated
type ContentInput struct {
	Text    string `json:"text"`
	Context string `json:"context"`
}

// ContentCheck represents a specific aspect to check
type ContentCheck struct {
	Category    string  `json:"category" jsonschema:"required"`
	IsViolation bool    `json:"is_violation" jsonschema:"required"`
	Confidence  float64 `json:"confidence" jsonschema:"required,minimum=0,maximum=1"`
	Reason      string  `json:"reason,omitempty"`
}

// ModeratorVote represents a single moderator's decision
type ModeratorVote struct {
	Decision   string   `json:"decision" jsonschema:"required,enum=approve,enum=reject"`
	Confidence float64  `json:"confidence" jsonschema:"required,minimum=0,maximum=1"`
	Reasons    []string `json:"reasons" jsonschema:"required"`
}

var (
	// Check for harmful content
	checkHarmful, _ = promptgen.Create[ContentInput, ContentCheck](`
		Analyze this content for harmful elements (violence, hate speech, etc.):
		Context: {{.Context}}
		Content: {{.Text}}

		Determine if there are any harmful elements present.
		Respond with high confidence only if you're very sure.
		Do not make any other comments, unrelated to the specific ask.
	`)

	// Check for spam/promotional content
	checkSpam, _ = promptgen.Create[ContentInput, ContentCheck](`
		Analyze this content for spam or promotional elements:
		Context: {{.Context}}
		Content: {{.Text}}

		Determine if this is spam or promotional content.
		Consider the context carefully - promotional content might be appropriate in some contexts.
		Do not make any other comments, unrelated to the specific ask.
	`)

	// Check for inappropriate content
	checkInappropriate, _ = promptgen.Create[ContentInput, ContentCheck](`
		Analyze this content for inappropriate elements:
		Context: {{.Context}}
		Content: {{.Text}}

		Consider cultural and contextual nuances when determining if content is inappropriate.
		Flag subtle inappropriate content but with lower confidence scores.
		Do not make any other comments, unrelated to the specific ask.
	`)

	// Independent moderator vote
	getModerationVote, _ = promptgen.Create[ContentInput, ModeratorVote](`
		As an independent content moderator, review this content:
		Context: {{.Context}}
		Content: {{.Text}}

		Consider:
		1. The appropriateness for the given context
		2. Any subtle harmful elements
		3. Whether promotional content is relevant/appropriate
		4. Cultural sensitivities

		Provide detailed reasons and adjust confidence based on certainty.
	`)
)

// RunParallelChecks runs all content checks in parallel
func RunParallelChecks(ctx context.Context, input ContentInput) ([]ContentCheck, error) {
	var wg sync.WaitGroup
	checks := make([]ContentCheck, 3)
	errs := make([]error, 3)

	// Run all checks in parallel
	checkers := []struct {
		index int
		name  string
		fn    func(context.Context, ContentInput) (ContentCheck, error)
	}{
		{0, "harmful", checkHarmful.Run},
		{1, "spam", checkSpam.Run},
		{2, "inappropriate", checkInappropriate.Run},
	}

	for _, c := range checkers {
		wg.Add(1)
		go func(c struct {
			index int
			name  string
			fn    func(context.Context, ContentInput) (ContentCheck, error)
		}) {
			defer wg.Done()
			result, err := c.fn(ctx, input)
			if err != nil {
				errs[c.index] = fmt.Errorf("%s check failed: %w", c.name, err)
				return
			}
			checks[c.index] = result
		}(c)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return checks, nil
}

// GetConsensusVote gets multiple independent votes and determines consensus
func GetConsensusVote(ctx context.Context, input ContentInput, numVoters int) (string, []ModeratorVote, error) {
	var wg sync.WaitGroup
	votes := make([]ModeratorVote, numVoters)
	errs := make([]error, numVoters)

	// Get votes in parallel
	for i := 0; i < numVoters; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			vote, err := getModerationVote.Run(ctx, input)
			if err != nil {
				errs[index] = fmt.Errorf("voter %d failed: %w", index+1, err)
				return
			}
			votes[index] = vote
		}(i)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errs {
		if err != nil {
			return "", nil, err
		}
	}

	return calculateConsensus(votes)
}

func calculateConsensus(votes []ModeratorVote) (string, []ModeratorVote, error) {
	approveCount := 0
	totalConfidence := 0.0

	for _, vote := range votes {
		if vote.Decision == "approve" {
			approveCount++
		}
		totalConfidence += vote.Confidence
	}

	avgConfidence := totalConfidence / float64(len(votes))

	// If confidence is low or votes are evenly split, require manual review
	if avgConfidence < 0.8 || approveCount == len(votes)/2 {
		return "manual_review", votes, nil
	}

	if approveCount > len(votes)/2 {
		return "approve", votes, nil
	}
	return "reject", votes, nil
}

func runExample(ctx context.Context, name string, input ContentInput) {
	fmt.Printf("\n=== Running Example: %s ===\n", name)
	fmt.Printf("Content: %s\n", input.Text)
	fmt.Printf("Context: %s\n\n", input.Context)

	// Run parallel content checks
	fmt.Println("Running parallel content checks...")
	checks, err := RunParallelChecks(ctx, input)
	if err != nil {
		log.Printf("Parallel checks failed: %v\n", err)
		return
	}

	for _, check := range checks {
		fmt.Printf("\nCheck Result (%s):\n", check.Category)
		fmt.Printf("Violation: %v\n", check.IsViolation)
		fmt.Printf("Confidence: %.2f\n", check.Confidence)
		if check.Reason != "" {
			fmt.Printf("Reason: %s\n", check.Reason)
		}
	}

	// Get consensus vote
	fmt.Println("\nGetting consensus vote from multiple moderators...")
	decision, votes, err := GetConsensusVote(ctx, input, 3)
	if err != nil {
		log.Printf("Consensus vote failed: %v\n", err)
		return
	}

	fmt.Printf("\nFinal Decision: %s\n", decision)
	fmt.Println("\nIndividual Votes:")
	for i, vote := range votes {
		fmt.Printf("\nModerator %d:\n", i+1)
		fmt.Printf("Decision: %s\n", vote.Decision)
		fmt.Printf("Confidence: %.2f\n", vote.Confidence)
		fmt.Printf("Reasons:\n")
		for _, reason := range vote.Reasons {
			fmt.Printf("- %s\n", reason)
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	examples := []struct {
		name  string
		input ContentInput
	}{
		{
			name: "Obvious Spam",
			input: ContentInput{
				Text:    "Check out these amazing deals! Limited time offer at http://example.com. Don't miss out on this incredible opportunity to make money fast!",
				Context: "Comment on a technical blog post",
			},
		},
		{
			name: "Subtle Inappropriate Content",
			input: ContentInput{
				Text:    "This solution might work, but it seems a bit naive. Just like some people I know...",
				Context: "Code review comment",
			},
		},
		{
			name: "Borderline Promotional",
			input: ContentInput{
				Text:    "I've written a detailed blog post about solving this exact problem at myblog.dev/solution. It includes benchmarks and performance comparisons.",
				Context: "Answer to a Stack Overflow question",
			},
		},
		{
			name: "Cultural Sensitivity",
			input: ContentInput{
				Text:    "This approach is totally crazy! It's completely mental to do it this way.",
				Context: "Technical discussion in an international forum",
			},
		},
	}

	for _, ex := range examples {
		runExample(ctx, ex.name, ex.input)
	}
}
