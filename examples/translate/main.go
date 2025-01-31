package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/arjunsriva/promptgen"
	"github.com/samber/lo"
)

type TranslationRequest struct {
	SourceText     string `json:"source_text"`
	TargetLanguage string `json:"target_language"`
	Context        string `json:"context"`
}

type TranslationResponse struct {
	TranslatedText string `json:"translated_text" jsonschema:"required,description=The translated text"`
}

var translateText, _ = promptgen.Create[TranslationRequest, TranslationResponse](`
Please translate the text to {{.TargetLanguage}}. based on the context provided.
You should only return the {{.TargetLanguage}} text.

Context and Request: {{.Context}}

Text to translate: {{.SourceText}}
`)

func attemptTranslation(request TranslationRequest, model string) (TranslationResponse, error) {
	log.Printf("Attempting translation with model: %s", model)
	log.Printf("Request: source='%s', target='%s', context='%s'",
		request.SourceText, request.TargetLanguage, request.Context)

	response, err := translateText.WithModel(model).Run(context.Background(), request)
	if err != nil {
		log.Printf("Translation error with %s: %v", model, err)
		return response, err
	}

	log.Printf("Translation result: '%s'", response.TranslatedText)
	return response, nil
}

func containsNonEnglishCharacters(text string) bool {
	// Matches any character that is not a-z, A-Z, numbers, spaces, or common punctuation
	re := regexp.MustCompile(`[^a-zA-Z0-9\s.,!?'"()-]`)
	hasNonEnglish := re.MatchString(text)
	if hasNonEnglish {
		log.Printf("Non-English characters detected in: '%s'", text)
		if matches := re.FindAllString(text, -1); matches != nil {
			log.Printf("Non-English characters found: %v", matches)
		}
	}
	return hasNonEnglish
}

func containsNonJapaneseCharacters(text string) bool {
	// Matches characters in the Japanese writing system ranges:
	// Hiragana (3040-309F), Katakana (30A0-30FF),
	// Kanji/CJK Unified Ideographs (4E00-9FFF),
	// Full-width punctuation and symbols (3000-303F),
	// Half-width Katakana (FF65-FF9F)
	japanesePattern := regexp.MustCompile(`^[\p{Han}\p{Hiragana}\p{Katakana}\s\p{P}0-9]+$`)
	hasNonJapanese := !japanesePattern.MatchString(text)
	if hasNonJapanese {
		log.Printf("Non-Japanese characters detected in: '%s'", text)
	}
	return hasNonJapanese
}

func translateWithFallback(targetLanguage string, context string, sourceText string) (string, error) {
	log.Printf("Starting translation: target='%s', context='%s', source='%s'",
		targetLanguage, context, sourceText)

	request := TranslationRequest{
		SourceText:     sourceText,
		TargetLanguage: targetLanguage,
		Context:        context,
	}

	// Try with the newer model first
	log.Printf("Attempting translation with primary model")
	translation, err := attemptTranslation(request, "gpt-4o-2024-08-06")
	if err != nil {
		log.Printf("Primary model failed, falling back to GPT-4")
		// Fall back to GPT-4
		translation, err = attemptTranslation(request, "gpt-4")
		if err != nil {
			log.Printf("Fallback model also failed: %v", err)
			return "", err
		}
	}

	// Validate translation based on target language
	requiresRetry := (targetLanguage == "English" && containsNonEnglishCharacters(translation.TranslatedText)) ||
		(targetLanguage == "Japanese" && containsNonJapaneseCharacters(translation.TranslatedText))

	if requiresRetry {
		log.Printf("Translation validation failed, retrying with GPT-4")
		translation, err = attemptTranslation(request, "gpt-4")
		if err != nil {
			log.Printf("Retry translation failed: %v", err)
			return "", err
		}
	}

	log.Printf("Translation completed successfully")
	return translation.TranslatedText, nil
}

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Printf("Starting translation example")

	testCases := []struct {
		sourceText string
		targetLang string
		context    string
	}{
		{
			sourceText: "Hello World",
			targetLang: "Japanese",
			context:    "Casual greeting",
		},
		{
			sourceText: "Please proceed with caution",
			targetLang: "Japanese",
			context:    "Warning sign in a construction site",
		},
		{
			sourceText: "こんにちは世界",
			targetLang: "English",
			context:    "Casual greeting",
		},
		{
			sourceText: "ご注意ください",
			targetLang: "English",
			context:    "Warning sign in a public space",
		},
		{
			sourceText: "The meeting will be held at 3 PM tomorrow",
			targetLang: "Japanese",
			context:    "Business email",
		},
		{
			sourceText: "明日の会議は午後3時から開催されます",
			targetLang: "English",
			context:    "Business email",
		},
		{
			sourceText: "Artificial Intelligence",
			targetLang: "Japanese",
			context:    "Name of technology being discussed in the news",
		},
	}

	// Create a channel to collect results
	type translationResult struct {
		index       int
		sourceText  string
		translation string
		targetLang  string
		err         error
	}

	// Process translations in parallel using lo
	results := lo.Map(testCases, func(tc struct {
		sourceText string
		targetLang string
		context    string
	}, i int) translationResult {
		log.Printf("\n=== Starting Test Case %d ===", i+1)
		log.Printf("Source: '%s'", tc.sourceText)
		log.Printf("Target Language: %s", tc.targetLang)
		log.Printf("Context: %s", tc.context)

		translatedText, err := translateWithFallback(tc.targetLang, tc.context, tc.sourceText)
		return translationResult{
			index:       i,
			sourceText:  tc.sourceText,
			translation: translatedText,
			targetLang:  tc.targetLang,
			err:         err,
		}
	})

	// Print results in order
	for _, result := range results {
		fmt.Printf("\nTest Case %d:\n", result.index+1)
		if result.err != nil {
			log.Printf("❌ Translation failed with error: %v", result.err)
			fmt.Printf("Error on test case %d: %v\n", result.index+1, result.err)
			continue
		}

		log.Printf("✅ Success: '%s' -> '%s'", result.sourceText, result.translation)
		fmt.Printf("Source (%s): %s\n", result.targetLang, result.sourceText)
		fmt.Printf("Translation: %s\n", result.translation)
		fmt.Println(strings.Repeat("-", 40))
	}
}
