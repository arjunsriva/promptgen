package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/arjunsriva/promptgen"
)

// TranslationInput represents the text to be translated
type TranslationInput struct {
	SourceText string `json:"source_text"`
	FromLang   string `json:"from_lang"`
	ToLang     string `json:"to_lang"`
	Context    string `json:"context"`
}

// Translation represents a translation attempt
type Translation struct {
	Text           string   `json:"text" jsonschema:"required"`
	Notes          []string `json:"notes,omitempty"`
	PreservedTerms []string `json:"preserved_terms,omitempty"`
}

// Evaluation represents feedback on a translation
type Evaluation struct {
	Score        float64  `json:"score" jsonschema:"required,minimum=0,maximum=1"`
	Issues       []string `json:"issues" jsonschema:"required"`
	Suggestions  []string `json:"suggestions" jsonschema:"required"`
	IsAcceptable bool     `json:"is_acceptable" jsonschema:"required"`
}

var (
	// Initial translator
	translate, _ = promptgen.Create[TranslationInput, Translation](`
		Translate this text from {{.FromLang}} to {{.ToLang}}.
		Context: {{.Context}}

		Source text: {{.SourceText}}

		Provide:
		1. The translated text
		2. Any notes about specific translation choices
		3. List of terms that should remain unchanged (names, technical terms)
	`)

	// Evaluator that reviews translations
	evaluate, _ = promptgen.Create[struct {
		Original    TranslationInput
		Translation Translation
	}, Evaluation](`
		Critically evaluate this translation from {{.Original.FromLang}} to {{.Original.ToLang}}:

		Original: {{.Original.SourceText}}
		Translation: {{.Translation.Text}}
		Context: {{.Original.Context}}

		Analyze with extreme scrutiny:
		1. Semantic accuracy: Check for subtle meaning shifts
		2. Cultural elements: Evaluate handling of cultural-specific concepts
		3. Domain authenticity: Verify natural usage in target domain
		4. Register accuracy: Compare source/target formality levels
		5. Stylistic elements: Assess rhythm, tone, and rhetorical devices

		For business translations:
		- Check keigo levels
		- Verify humble/honorific form usage
		- Assess business-specific phraseology

		For literary translations:
		- Evaluate poetic devices
		- Check rhythm/meter preservation
		- Assess imagery transfer

		For technical translations:
		- Verify domain-specific terminology
		- Check procedural clarity
		- Assess technical accuracy

		Score strictly - anything above 0.8 must be exceptional.
		Provide detailed issues and specific improvement suggestions.
		Mark as acceptable only if translation would be suitable for professional use.`)

	// Optimizer that improves based on feedback
	optimize, _ = promptgen.Create[struct {
		Original   TranslationInput
		Current    Translation
		Evaluation Evaluation
	}, Translation](`
		Improve this translation based on the evaluation:

		Original ({{.Original.FromLang}}): {{.Original.SourceText}}
		Current ({{.Original.ToLang}}): {{.Current.Text}}

		Issues identified:
		{{range .Evaluation.Issues}}
		- {{.}}
		{{end}}

		Suggestions:
		{{range .Evaluation.Suggestions}}
		- {{.}}
		{{end}}

		Provide an improved translation addressing these points.
	`)
)

func improveTranslation(ctx context.Context, input TranslationInput, maxIterations int) (Translation, error) {
	// Initial translation
	current, err := translate.Run(ctx, input)
	if err != nil {
		return Translation{}, fmt.Errorf("initial translation failed: %w", err)
	}

	fmt.Printf("Initial translation:\n%s\n\n", current.Text)

	for i := 0; i < maxIterations; i++ {
		// Evaluate current translation
		eval, err := evaluate.Run(ctx, struct {
			Original    TranslationInput
			Translation Translation
		}{input, current})
		if err != nil {
			return Translation{}, fmt.Errorf("evaluation failed: %w", err)
		}

		fmt.Printf("Iteration %d Evaluation:\n", i+1)
		fmt.Printf("Score: %.2f\n", eval.Score)
		fmt.Printf("Issues:\n%s\n", strings.Join(eval.Issues, "\n"))

		// Check if translation is acceptable
		if eval.IsAcceptable {
			fmt.Printf("\nAcceptable translation reached after %d iterations\n", i+1)
			return current, nil
		}

		// Optimize based on feedback
		improved, err := optimize.Run(ctx, struct {
			Original   TranslationInput
			Current    Translation
			Evaluation Evaluation
		}{input, current, eval})
		if err != nil {
			return Translation{}, fmt.Errorf("optimization failed: %w", err)
		}

		current = improved
		fmt.Printf("\nImproved translation:\n%s\n\n", current.Text)
	}

	return current, fmt.Errorf("reached maximum iterations without achieving acceptable quality")
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	testCases := []TranslationInput{
		{
			// Haiku Test
			SourceText: `古池や
蛙飛び込む
水の音`,
			FromLang: "Japanese",
			ToLang:   "English",
			Context:  "Basho's haiku emphasizing yugen and kigo",
		},
		{
			// Business Keigo Test
			SourceText: `誠に恐れ入りますが、先日ご依頼いただきました件につきまして、
納期を一週間延長させていただきたくご相談申し上げます。`,
			FromLang: "Japanese",
			ToLang:   "English",
			Context:  "Formal business email requesting deadline extension",
		},
		{
			// Technical Manual
			SourceText: `システムの初期化中にメモリオーバーフローが発生した場合、
即座にログファイルにスタックトレースを出力し、
セーフモードで再起動してください。`,
			FromLang: "Japanese",
			ToLang:   "English",
			Context:  "Software debugging manual, technical terminology preservation required",
		},
		{
			// Cultural Concept
			SourceText: "The tea master arranged the flowers with wabi-sabi in mind, embracing the beauty of imperfection.",
			FromLang:   "English",
			ToLang:     "Japanese",
			Context:    "Tea ceremony description, cultural aesthetic concepts",
		},
		{
			// Mixed Language/Modern
			SourceText: "新しいスマートフォンアプリでAIチャットボットとコミュニケーションが可能です！",
			FromLang:   "Japanese",
			ToLang:     "English",
			Context:    "Tech marketing content with Katakana loanwords",
		},
	}

	for i, testCase := range testCases {
		fmt.Printf("\n=== Test Case %d ===\n", i+1)
		fmt.Printf("Source (%s): %s\n", testCase.FromLang, testCase.SourceText)
		fmt.Printf("Context: %s\n\n", testCase.Context)

		result, err := improveTranslation(ctx, testCase, 3)
		if err != nil {
			log.Printf("Test case %d failed: %v\n", i+1, err)
			continue
		}

		fmt.Printf("\nFinal Translation (%s):\n", testCase.ToLang)
		fmt.Printf("Text: %s\n", result.Text)
		if len(result.Notes) > 0 {
			fmt.Printf("\nTranslation Notes:\n")
			for _, note := range result.Notes {
				fmt.Printf("- %s\n", note)
			}
		}
		if len(result.PreservedTerms) > 0 {
			fmt.Printf("\nPreserved Terms:\n")
			for _, term := range result.PreservedTerms {
				fmt.Printf("- %s\n", term)
			}
		}
		fmt.Println("\n" + strings.Repeat("-", 80))
	}
}
