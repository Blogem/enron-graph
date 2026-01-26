package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/analyst"
	"github.com/Blogem/enron-graph/internal/promoter"
	"github.com/Blogem/enron-graph/pkg/utils"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

// T086: Analyst CLI implementation
// Commands: analyze (detect and rank candidates), promote (interactive promotion)

var rootCmd = &cobra.Command{
	Use:   "analyst",
	Short: "Analyst CLI for schema evolution and type promotion",
	Long:  "Analyzes discovered entities to identify promotion candidates and execute schema promotions",
}

var (
	minOccurrences int
	minConsistency float64
	topN           int
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze entities and rank promotion candidates",
	Long:  "Run pattern detection, clustering, and ranking to identify top candidates for type promotion",
	RunE:  runAnalyze,
}

var promoteCmd = &cobra.Command{
	Use:   "promote [type-name]",
	Short: "Promote a type to a schema",
	Long:  "Interactive promotion of a discovered type to a formal ent schema",
	Args:  cobra.ExactArgs(1),
	RunE:  runPromote,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(promoteCmd)

	// Add flags to analyze command
	analyzeCmd.Flags().IntVar(&minOccurrences, "min-occurrences", 5, "Minimum number of entity occurrences")
	analyzeCmd.Flags().Float64Var(&minConsistency, "min-consistency", 0.4, "Minimum property consistency (0.0-1.0)")
	analyzeCmd.Flags().IntVar(&topN, "top", 10, "Number of top candidates to display")
}

func getDBClient() (*ent.Client, error) {
	cfg, err := utils.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	client, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return client, nil
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Connect to database
	client, err := getDBClient()
	if err != nil {
		return err
	}
	defer client.Close()

	fmt.Println("Running pattern detection and analysis...")
	fmt.Printf("Thresholds: minOccurrences=%d, minConsistency=%.1f%%, topN=%d\n\n", minOccurrences, minConsistency*100, topN)

	// Run pattern detection and ranking
	candidates, err := analyst.AnalyzeAndRankCandidates(ctx, client, minOccurrences, minConsistency, topN)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Display results
	fmt.Printf("\nTop %d Promotion Candidates:\n\n", len(candidates))
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Rank\tType\tFrequency\tDensity\tConsistency\tScore")
	fmt.Fprintln(w, "----\t----\t---------\t-------\t-----------\t-----")
	for i, candidate := range candidates {
		fmt.Fprintf(w, "%d\t%s\t%d\t%.2f\t%.2f%%\t%.3f\n",
			i+1,
			candidate.Type,
			candidate.Frequency,
			candidate.Density,
			candidate.Consistency*100,
			candidate.Score,
		)
	}
	w.Flush()
	fmt.Println()

	return nil
}

func runPromote(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	typeName := args[0]

	// Connect to database
	client, err := getDBClient()
	if err != nil {
		return err
	}
	defer client.Close()

	// Get raw SQL connection for data migration
	cfg, err := utils.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open raw database connection: %w", err)
	}
	defer db.Close()

	fmt.Printf("Promoting type: %s\n\n", typeName)

	// Generate schema definition
	fmt.Println("Step 1: Generating schema definition...")
	schema, err := analyst.GenerateSchemaForType(ctx, client, typeName)
	if err != nil {
		return fmt.Errorf("schema generation failed: %w", err)
	}

	// Display properties
	fmt.Printf("  Found %d properties\n", len(schema.Properties))
	fmt.Println("\nProperties:")
	for propName, propDef := range schema.Properties {
		required := ""
		if propDef.Required {
			required = " (required)"
		}
		fmt.Printf("  - %s: %s%s\n", propName, propDef.Type, required)
	}

	// Confirm promotion
	fmt.Print("\nProceed with promotion? (yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" && confirm != "y" {
		fmt.Println("Promotion cancelled.")
		return nil
	}

	// Execute promotion
	fmt.Println("\nStep 2: Executing promotion workflow...")
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get project root: %w", err)
	}

	p := promoter.NewPromoter(client)
	p.SetDB(db) // Set raw SQL connection for data migration

	// Convert analyst.SchemaDefinition to promoter.SchemaDefinition
	promoterSchema := promoter.SchemaDefinition{
		Type:       schema.Type,
		Properties: make(map[string]promoter.PropertyDefinition),
	}
	for propName, propDef := range schema.Properties {
		promoterSchema.Properties[propName] = promoter.PropertyDefinition{
			Type:            propDef.Type,
			Required:        propDef.Required,
			ValidationRules: convertValidationRules(propDef.ValidationRules),
		}
	}

	req := promoter.PromotionRequest{
		TypeName:         typeName,
		SchemaDefinition: promoterSchema,
		OutputDir:        projectRoot + "/ent/schema",
		ProjectRoot:      projectRoot,
	}
	result, err := p.PromoteType(ctx, req)
	if err != nil {
		fmt.Printf("Promotion failed: %v\n", err)
		return err
	}

	// Display results
	fmt.Println("\nPromotion Results:")
	fmt.Printf("  Status: %s\n", getStatusIcon(result.Success))
	fmt.Printf("  Schema file: %s\n", result.SchemaFilePath)
	fmt.Printf("  Entities migrated: %d\n", result.EntitiesMigrated)
	fmt.Printf("  Validation errors: %d\n", result.ValidationErrors)

	if result.Success {
		fmt.Println("\nPromotion completed successfully!")
	} else {
		fmt.Printf("\nPromotion failed: %v\n", result.Error)
	}

	return nil
}

func getStatusIcon(success bool) string {
	if success {
		return "✓ SUCCESS"
	}
	return "✗ FAILED"
}

func convertValidationRules(rules []analyst.ValidationRule) []promoter.ValidationRule {
	if len(rules) == 0 {
		return nil
	}
	converted := make([]promoter.ValidationRule, len(rules))
	for i, rule := range rules {
		converted[i] = promoter.ValidationRule{
			Type:  rule.Type,
			Value: rule.Value,
		}
	}
	return converted
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
