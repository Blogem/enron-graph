package main

import (
	"fmt"

	"github.com/Blogem/enron-graph/internal/chat"
)

func main() {
	// Use pattern matcher for ambiguous query detection
	matcher := chat.NewPatternMatcher()

	// Test ambiguous queries
	ambiguousQueries := []string{
		"Who is Smith?",        // Common name
		"Tell me about it",     // Vague reference
		"What about that?",     // Unclear query
	}

	fmt.Println("Testing ambiguous query handling...")
	for _, query := range ambiguousQueries {
		result, _ := matcher.Match(query)
		if result.Ambiguous {
			fmt.Printf("✓ Detected ambiguous query: '%s'\n", query)
			fmt.Printf("  Options: %v\n", result.Options)
		} else {
			fmt.Printf("⚠ Query not flagged as ambiguous: '%s'\n", query)
		}
	}

	// Test clear queries
	clearQueries := []string{
		"Who is Jeff Skilling?",
		"Show me emails about energy trading",
	}

	fmt.Println("\nTesting clear query handling...")
	for _, query := range clearQueries {
		result, _ := matcher.Match(query)
		if !result.Ambiguous {
			fmt.Printf("✓ Clear query processed correctly: '%s' (type: %s)\n", query, result.Type)
		} else {
			fmt.Printf("❌ Clear query incorrectly flagged as ambiguous: '%s'\n", query)
		}
	}
}
