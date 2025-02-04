package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/arjunsriva/promptgen"
)

func main() {
	ctx := context.Background()

	// Example 1: String type - Joke generator
	stringGen, err := promptgen.Create[string, string]("Tell me a {{.}} joke, be creative, unusual")
	if err != nil {
		log.Fatalf("Failed to create string generator: %v", err)
	}

	theme := "Dad"
	joke, err := stringGen.Run(ctx, theme)
	handleError(err)
	fmt.Printf("\n1. String Example - %s joke:\n%s\n", theme, joke)

	// Example 2: Integer type - Age guesser
	intGen, err := promptgen.Create[string, int]("Given the description '{{.}}', guess the person's age")
	if err != nil {
		log.Fatalf("Failed to create int generator: %v", err)
	}

	description := "a college professor who loves coffee and has grey hair"
	age, err := intGen.Run(ctx, description)
	handleError(err)
	fmt.Printf("\n2. Integer Example - Age guess:\nDescription: %s\nEstimated age: %d years\n", description, age)

	// Example 3: Float type - Temperature converter
	floatGen, err := promptgen.Create[float64, float64]("Convert {{.}} degrees Fahrenheit to Celsius")
	if err != nil {
		log.Fatalf("Failed to create float generator: %v", err)
	}

	fahrenheit := 98.6
	celsius, err := floatGen.Run(ctx, fahrenheit)
	handleError(err)
	fmt.Printf("\n3. Float Example - Temperature conversion:\n%.1f°F = %.1f°C\n", fahrenheit, celsius)

	// Example 4: Boolean type - Decision maker
	boolGen, err := promptgen.Create[string, bool]("Given the scenario '{{.}}', should I do it? Answer with true or false")
	if err != nil {
		log.Fatalf("Failed to create boolean generator: %v", err)
	}

	scenario := "go skydiving on a cloudy day with inexperienced instructors"
	decision, err := boolGen.Run(ctx, scenario)
	handleError(err)
	fmt.Printf("\n4. Boolean Example - Decision:\nScenario: %s\nDecision: %v\n", scenario, decision)
}

func handleError(err error) {
	if err != nil {
		switch {
		case errors.Is(err, promptgen.ErrRateLimit):
			log.Fatal("Rate limit exceeded, please try again later")
		case errors.Is(err, promptgen.ErrContextLength):
			log.Fatal("Input too long, please reduce the content")
		default:
			log.Fatalf("Operation failed: %v", err)
		}
	}
}
