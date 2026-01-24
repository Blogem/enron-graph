package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/analyst"
	"github.com/Blogem/enron-graph/internal/promoter"
	"github.com/Blogem/enron-graph/pkg/utils"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

// T085: Promoter CLI implementation
// Standalone command for interactive schema promotion workflow

var rootCmd = &cobra.Command{
	Use:   "promoter",
	Short: "Schema promotion tool",
	Long:  "Promotes discovered entity types to formal ent schemas with data migration",
}

var promoteCmd = &cobra.Command{
	Use:   "promote [type-name]",
	Short: "Promote a type to a formal schema",
	Long:  "Generate ent schema, run migration, copy data, and validate the promotion",
	Args:  cobra.ExactArgs(1),
	RunE:  runPromote,
}

func init() {
	rootCmd.AddCommand(promoteCmd)
}

func getDBClient() (*ent.Client, error) {
	cfg, err := utils.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	client, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return client, nil
}

func runPromote(cmd *cobra.Command, args []string) error {
	typeName := args[0]

	client, err := getDBClient()
	if err != nil {
		return err
	}
	defer client.Close()

	ctx := context.Background()

	// Step 1: Load type definition from analyst results
	fmt.Printf("Loading type definition for: %s\n", typeName)
	schema, err := analyst.GenerateSchemaForType(ctx, client, typeName)
	if err != nil {
		return fmt.Errorf("failed to load type definition: %w", err)
	}

	// Step 2-7: Execute promotion workflow
	fmt.Printf("Starting promotion workflow...\n")
	p := promoter.NewPromoter(client)

	// Open raw SQL connection for data migration
	cfg, err := utils.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	sqlDB, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open raw SQL connection: %w", err)
	}
	defer sqlDB.Close()

	// Set the raw DB connection for data migration
	p.SetDB(sqlDB)

	req := promoter.PromotionRequest{
		TypeName: typeName,
		SchemaDefinition: promoter.SchemaDefinition{
			Type:       schema.Type,
			Properties: convertSchemaProperties(schema.Properties),
		},
		OutputDir:   "ent/schema",
		ProjectRoot: ".",
	}

	result, err := p.PromoteType(ctx, req)
	if err != nil {
		return fmt.Errorf("promotion failed: %w", err)
	}

	if result.Success {
		fmt.Printf("✓ Successfully promoted %s\n", typeName)
		fmt.Printf("  Schema file: %s\n", result.SchemaFilePath)
		fmt.Printf("  Entities migrated: %d\n", result.EntitiesMigrated)
		fmt.Printf("  Validation errors: %d\n", result.ValidationErrors)
	} else {
		fmt.Printf("✗ Promotion failed: %v\n", result.Error)
	}

	return nil
}

// convertSchemaProperties converts analyst.PropertyDefinition to promoter.PropertyDefinition
func convertSchemaProperties(props map[string]analyst.PropertyDefinition) map[string]promoter.PropertyDefinition {
	result := make(map[string]promoter.PropertyDefinition)

	for name, prop := range props {
		// Convert validation rules
		rules := make([]promoter.ValidationRule, 0, len(prop.ValidationRules))
		for _, rule := range prop.ValidationRules {
			rules = append(rules, promoter.ValidationRule{
				Type:  rule.Type,
				Value: rule.Value,
			})
		}

		result[name] = promoter.PropertyDefinition{
			Type:            prop.Type,
			Required:        prop.Required,
			ValidationRules: rules,
		}
	}

	return result
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
