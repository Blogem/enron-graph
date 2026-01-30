package promoter

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
)

// T085: Promotion workflow implementation
// Generate ent file, run go generate, migrate database, copy data, validate, audit

// PromotionRequest represents a request to promote a type to a schema
type PromotionRequest struct {
	TypeName         string
	SchemaDefinition SchemaDefinition
	OutputDir        string
	ProjectRoot      string
}

// PromotionResult contains the results of a promotion operation
type PromotionResult struct {
	Success          bool
	TypeName         string
	EntitiesMigrated int
	ValidationErrors int
	SchemaFilePath   string
	Error            error
}

// Promoter handles schema promotion workflow
type Promoter struct {
	client *ent.Client
	db     *sql.DB // Raw SQL connection for data migration
}

// NewPromoter creates a new Promoter instance
func NewPromoter(client *ent.Client) *Promoter {
	return &Promoter{
		client: client,
		db:     nil, // Will be set if needed for data migration
	}
}

// SetDB sets the raw database connection for data migration operations
func (p *Promoter) SetDB(db *sql.DB) {
	p.db = db
}

// GenerateEntSchema generates the ent schema file for a type
func (p *Promoter) GenerateEntSchema(req PromotionRequest) (string, error) {
	// Generate ent schema file
	if err := GenerateEntSchemaFile(req.SchemaDefinition, req.OutputDir); err != nil {
		return "", fmt.Errorf("failed to generate ent schema: %w", err)
	}

	schemaPath := filepath.Join(req.OutputDir, req.TypeName+".go")
	return schemaPath, nil
}

// RunEntGenerate executes go generate ./ent to regenerate ent code
func (p *Promoter) RunEntGenerate(projectRoot string) error {
	cmd := exec.Command("go", "generate", "./ent")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go generate: %w", err)
	}

	return nil
}

// MigrateDatabase runs database migration to create new table
func (p *Promoter) MigrateDatabase(ctx context.Context, projectRoot string) error {
	// We need to run the migrate command externally because the ent client
	// was created before the new schema existed, so it doesn't know about
	// the new table yet. The migrate command will rebuild with the new schema.
	cmd := exec.Command("go", "run", "cmd/migrate/main.go")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run migration: %w (output: %s)", err, string(output))
	}

	return nil
}

// CopyEntities copies data from DiscoveredEntity to the new typed table using raw SQL
func (p *Promoter) CopyEntities(ctx context.Context, typeName string, schema SchemaDefinition) (int, error) {
	// Query all entities of this type
	entities, err := p.client.DiscoveredEntity.
		Query().
		Where(discoveredentity.TypeCategory(typeName)).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to query entities: %w", err)
	}

	if len(entities) == 0 {
		return 0, nil
	}

	// If no raw DB connection, just return count (data migration will be manual)
	if p.db == nil {
		fmt.Printf("Note: No raw SQL connection available. %d entities ready for manual migration\n", len(entities))
		return len(entities), nil
	}

	// Build column list from schema properties
	columns := []string{}
	for propName := range schema.Properties {
		columns = append(columns, propName)
	}

	// Calculate table name (ent pluralizes by adding 's')
	tableName := strings.ToLower(typeName) + "s"

	// Use a transaction for atomic insertion
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Track ID mapping: old discovered_entities.id → new promoted table id
	oldToNewIDMap := make(map[int]int)

	count := 0
	for _, entity := range entities {
		// Build INSERT statement dynamically with RETURNING clause
		columnNames := []string{}
		values := []interface{}{}
		placeholders := []string{}

		idx := 1
		for _, colName := range columns {
			if val, exists := entity.Properties[colName]; exists {
				columnNames = append(columnNames, colName)
				values = append(values, val)
				placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
				idx++
			}
		}

		if len(columnNames) > 0 {
			query := fmt.Sprintf(
				"INSERT INTO %s (%s) VALUES (%s) RETURNING id",
				tableName,
				strings.Join(columnNames, ", "),
				strings.Join(placeholders, ", "),
			)

			var newID int
			if err := tx.QueryRowContext(ctx, query, values...).Scan(&newID); err != nil {
				return 0, fmt.Errorf("failed to insert entity %d: %w", entity.ID, err)
			}

			// Store the ID mapping
			oldToNewIDMap[entity.ID] = newID
			count++
		}
	}

	// Update relationships - both FROM and TO references
	if len(oldToNewIDMap) > 0 {
		// Update FROM references
		fromCount, err := updateRelationshipsFrom(ctx, tx, "discovered_entity", typeName, oldToNewIDMap)
		if err != nil {
			return 0, fmt.Errorf("failed to update FROM relationships: %w", err)
		}
		fmt.Printf("Updated %d FROM relationships\n", fromCount)

		// Update TO references
		toCount, err := updateRelationshipsTo(ctx, tx, "discovered_entity", typeName, oldToNewIDMap)
		if err != nil {
			return 0, fmt.Errorf("failed to update TO relationships: %w", err)
		}
		fmt.Printf("Updated %d TO relationships\n", toCount)

		// Delete old entities from discovered_entities
		deleteCount, err := deleteOldEntities(ctx, tx, oldToNewIDMap)
		if err != nil {
			return 0, fmt.Errorf("failed to delete old entities: %w", err)
		}
		fmt.Printf("Deleted %d old entities from discovered_entities\n", deleteCount)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("Successfully migrated %d entities to table '%s'\n", count, tableName)
	return count, nil
}

// updateRelationshipsFrom updates FROM references in relationships table
func updateRelationshipsFrom(ctx context.Context, tx *sql.Tx, oldType, newType string, idMap map[int]int) (int, error) {
	// Build list of old IDs
	oldIDs := make([]int, 0, len(idMap))
	for oldID := range idMap {
		oldIDs = append(oldIDs, oldID)
	}

	if len(oldIDs) == 0 {
		return 0, nil
	}

	// Update in batches of 1000
	batchSize := 1000
	totalUpdated := 0

	for i := 0; i < len(oldIDs); i += batchSize {
		end := i + batchSize
		if end > len(oldIDs) {
			end = len(oldIDs)
		}
		batch := oldIDs[i:end]

		// Update each relationship in this batch
		for _, oldID := range batch {
			newID := idMap[oldID]
			query := `
				UPDATE relationships 
				SET from_type = $1, from_id = $2 
				WHERE from_type = $3 AND from_id = $4
			`
			result, err := tx.ExecContext(ctx, query, newType, newID, oldType, oldID)
			if err != nil {
				return totalUpdated, err
			}

			rows, _ := result.RowsAffected()
			totalUpdated += int(rows)
		}
	}

	return totalUpdated, nil
}

// updateRelationshipsTo updates TO references in relationships table
func updateRelationshipsTo(ctx context.Context, tx *sql.Tx, oldType, newType string, idMap map[int]int) (int, error) {
	// Build list of old IDs
	oldIDs := make([]int, 0, len(idMap))
	for oldID := range idMap {
		oldIDs = append(oldIDs, oldID)
	}

	if len(oldIDs) == 0 {
		return 0, nil
	}

	// Update in batches of 1000
	batchSize := 1000
	totalUpdated := 0

	for i := 0; i < len(oldIDs); i += batchSize {
		end := i + batchSize
		if end > len(oldIDs) {
			end = len(oldIDs)
		}
		batch := oldIDs[i:end]

		// Update each relationship in this batch
		for _, oldID := range batch {
			newID := idMap[oldID]
			query := `
				UPDATE relationships 
				SET to_type = $1, to_id = $2 
				WHERE to_type = $3 AND to_id = $4
			`
			result, err := tx.ExecContext(ctx, query, newType, newID, oldType, oldID)
			if err != nil {
				return totalUpdated, err
			}

			rows, _ := result.RowsAffected()
			totalUpdated += int(rows)
		}
	}

	return totalUpdated, nil
}

// deleteOldEntities deletes entities from discovered_entities after promotion
func deleteOldEntities(ctx context.Context, tx *sql.Tx, idMap map[int]int) (int, error) {
	// Build list of old IDs to delete
	oldIDs := make([]int, 0, len(idMap))
	for oldID := range idMap {
		oldIDs = append(oldIDs, oldID)
	}

	if len(oldIDs) == 0 {
		return 0, nil
	}

	// Delete in batches of 1000
	batchSize := 1000
	totalDeleted := 0

	for i := 0; i < len(oldIDs); i += batchSize {
		end := i + batchSize
		if end > len(oldIDs) {
			end = len(oldIDs)
		}
		batch := oldIDs[i:end]

		// Build placeholders for IN clause
		placeholders := make([]string, len(batch))
		args := make([]interface{}, len(batch))
		for j, id := range batch {
			placeholders[j] = fmt.Sprintf("$%d", j+1)
			args[j] = id
		}

		query := fmt.Sprintf(
			"DELETE FROM discovered_entities WHERE id IN (%s)",
			strings.Join(placeholders, ", "),
		)

		result, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return totalDeleted, err
		}

		rows, _ := result.RowsAffected()
		totalDeleted += int(rows)
	}

	return totalDeleted, nil
}

// ValidateEntities validates entities against the schema
func (p *Promoter) ValidateEntities(ctx context.Context, typeName string, schema SchemaDefinition) (int, error) {
	// Query entities
	entities, err := p.client.DiscoveredEntity.
		Query().
		Where(discoveredentity.TypeCategory(typeName)).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to query entities: %w", err)
	}

	errorCount := 0

	// Validate each entity
	for _, entity := range entities {
		props := entity.Properties

		// Check required properties
		for propName, propDef := range schema.Properties {
			if propDef.Required {
				if _, exists := props[propName]; !exists {
					errorCount++
					break
				}
			}
		}
	}

	return errorCount, nil
}

// CreateAuditRecord creates a SchemaPromotion audit record
func (p *Promoter) CreateAuditRecord(ctx context.Context, result PromotionResult) error {
	// Create audit record
	_, err := p.client.SchemaPromotion.
		Create().
		SetTypeName(result.TypeName).
		SetEntitiesAffected(result.EntitiesMigrated).
		SetValidationFailures(result.ValidationErrors).
		SetPromotedAt(time.Now()).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to create audit record: %w", err)
	}

	return nil
}

// PromoteType executes the complete promotion workflow
func (p *Promoter) PromoteType(ctx context.Context, req PromotionRequest) (*PromotionResult, error) {
	result := &PromotionResult{
		TypeName: req.TypeName,
	}

	// Step 1: Generate ent schema file
	schemaPath, err := p.GenerateEntSchema(req)
	if err != nil {
		result.Error = fmt.Errorf("schema generation failed: %w", err)
		result.Success = false
		p.CreateAuditRecord(ctx, *result)
		return result, result.Error
	}
	result.SchemaFilePath = schemaPath

	// Step 2: Run go generate ./ent
	if err := p.RunEntGenerate(req.ProjectRoot); err != nil {
		result.Error = fmt.Errorf("code generation failed: %w", err)
		result.Success = false
		p.CreateAuditRecord(ctx, *result)
		return result, result.Error
	}

	// Step 3: Run database migration
	if err := p.MigrateDatabase(ctx, req.ProjectRoot); err != nil {
		result.Error = fmt.Errorf("migration failed: %w", err)
		result.Success = false
		p.CreateAuditRecord(ctx, *result)
		return result, result.Error
	}

	// Step 4: Validate entities
	validationErrors, err := p.ValidateEntities(ctx, req.TypeName, req.SchemaDefinition)
	if err != nil {
		result.Error = fmt.Errorf("validation failed: %w", err)
		result.Success = false
		p.CreateAuditRecord(ctx, *result)
		return result, result.Error
	}
	result.ValidationErrors = validationErrors

	// Step 5: Copy data from discovered_entities to new typed table
	count, err := p.CopyEntities(ctx, req.TypeName, req.SchemaDefinition)
	if err != nil {
		result.Error = fmt.Errorf("data copy failed: %w", err)
		result.Success = false
		p.CreateAuditRecord(ctx, *result)
		return result, result.Error
	}
	result.EntitiesMigrated = count

	// Step 6: Create audit record
	result.Success = true
	if err := p.CreateAuditRecord(ctx, *result); err != nil {
		return result, fmt.Errorf("audit record creation failed: %w", err)
	}

	return result, nil
}
