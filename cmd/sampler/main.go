package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// Command line flags
	count := flag.Int("count", 0, "Number of emails to extract (required)")
	help := flag.Bool("help", false, "Show usage information")

	flag.Parse()

	// Show help
	if *help {
		printUsage()
		os.Exit(0)
	}

	// Validate flags
	if *count <= 0 {
		log.Fatal("--count must be a positive integer")
	}

	fmt.Printf("Random Email Sampler - extracting %d emails\n", *count)
	// TODO: Wire up extraction logic in later phases
}

func printUsage() {
	fmt.Println("Random Email Sampler")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/sampler/main.go --count <number>")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --count  Number of emails to extract (required, must be positive)")
	fmt.Println("  --help   Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/sampler/main.go --count 10")
	fmt.Println("  go run cmd/sampler/main.go --count 1000")
}
