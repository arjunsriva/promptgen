// Example of building a chat application
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arjunsriva/promptgen"
)

type ChatInput struct {
	History []Message `json:"history"`
	Query   string    `json:"query"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Reply    string   `json:"reply" jsonschema:"required,maxLength=1000"`
	Intent   string   `json:"intent" jsonschema:"enum=question,enum=statement,enum=command"`
	Keywords []string `json:"keywords" jsonschema:"minItems=1,maxItems=5"`
}

func main() {
	chat, err := promptgen.Create[ChatInput, ChatResponse](`
        Chat History:
        {{range .History}}
        {{.Role}}: {{.Content}}
        {{end}}
        
        User: {{.Query}}
        
        Respond naturally while detecting the intent and key topics.
    `)
	if err != nil {
		log.Fatal(err)
	}

	chat.WithTimeout(10 * time.Second).
		WithTemperature(0.7)

	resp, err := chat.Run(context.Background(), ChatInput{
		Query: "What's the weather like?",
		History: []Message{
			{Role: "user", Content: "Hello!"},
			{Role: "assistant", Content: "Hi there! How can I help?"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Reply: %s\nIntent: %s\nKeywords: %v\n",
		resp.Reply, resp.Intent, resp.Keywords)
}
