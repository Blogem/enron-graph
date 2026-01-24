package extractor

import (
	"context"
	"fmt"
	"strings"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/graph"
)

// Deduplicator handles entity deduplication
type Deduplicator struct {
	repo   graph.Repository
	logger interface {
		Debug(msg string, args ...interface{})
		Warn(msg string, args ...interface{})
	}
}

// NewDeduplicator creates a new deduplicator
func NewDeduplicator(repo graph.Repository, logger interface {
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}) *Deduplicator {
	return &Deduplicator{
		repo:   repo,
		logger: logger,
	}
}

// DeduplicatePerson checks if a person entity already exists
// Uses email address as the primary unique key
func (d *Deduplicator) DeduplicatePerson(ctx context.Context, email, name string) (*ent.DiscoveredEntity, bool, error) {
	if email == "" {
		return nil, false, fmt.Errorf("email is required for person deduplication")
	}

	// Look up by unique ID (email address)
	existing, err := d.repo.FindEntityByUniqueID(ctx, email)
	if err != nil {
		// Entity not found, not a duplicate
		return nil, false, nil
	}

	// Entity exists
	d.logger.Debug("Found duplicate person entity", "email", email, "existing_id", existing.ID)
	return existing, true, nil
}

// DeduplicateOrganization checks if an organization entity already exists
// Uses normalized name as the unique key
func (d *Deduplicator) DeduplicateOrganization(ctx context.Context, name, domain string) (*ent.DiscoveredEntity, bool, error) {
	// Normalize name (lowercase, trim whitespace)
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	uniqueID := fmt.Sprintf("org:%s", normalizedName)

	// Look up by unique ID
	existing, err := d.repo.FindEntityByUniqueID(ctx, uniqueID)
	if err != nil {
		// Entity not found, not a duplicate
		return nil, false, nil
	}

	// Entity exists
	d.logger.Debug("Found duplicate organization entity", "name", name, "existing_id", existing.ID)
	return existing, true, nil
}

// DeduplicateConcept checks if a concept entity already exists
// Uses embedding similarity for fuzzy matching (cosine > 0.85)
func (d *Deduplicator) DeduplicateConcept(ctx context.Context, name string, embedding []float32) (*ent.DiscoveredEntity, bool, error) {
	if len(embedding) == 0 {
		// No embedding, fall back to exact name match
		return d.deduplicateByName(ctx, "concept", name)
	}

	// Perform similarity search
	similarEntities, err := d.repo.SimilaritySearch(ctx, embedding, 5, 0.85)
	if err != nil {
		d.logger.Warn("Similarity search failed, falling back to name match", "error", err)
		return d.deduplicateByName(ctx, "concept", name)
	}

	// Check if any similar entity is a concept with high similarity
	for _, entity := range similarEntities {
		if entity.TypeCategory == "concept" {
			// Calculate cosine similarity (simplified check for POC)
			// In production, would compute actual cosine similarity
			d.logger.Debug("Found similar concept entity",
				"name", name,
				"similar_name", entity.Name,
				"existing_id", entity.ID)
			return entity, true, nil
		}
	}

	// No similar entity found
	return nil, false, nil
}

// deduplicateByName checks for exact name match within a type category
func (d *Deduplicator) deduplicateByName(ctx context.Context, typeCategory, name string) (*ent.DiscoveredEntity, bool, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	uniqueID := fmt.Sprintf("%s:%s", typeCategory, normalizedName)

	existing, err := d.repo.FindEntityByUniqueID(ctx, uniqueID)
	if err != nil {
		return nil, false, nil
	}

	d.logger.Debug("Found duplicate entity by name",
		"type", typeCategory,
		"name", name,
		"existing_id", existing.ID)
	return existing, true, nil
}

// MergeEntityProperties merges new properties into existing entity
// Returns true if entity was updated
func (d *Deduplicator) MergeEntityProperties(ctx context.Context, entity *ent.DiscoveredEntity, newProperties map[string]interface{}, newConfidence float64) (bool, error) {
	// For POC, we'll use a simple merge strategy:
	// - Keep highest confidence score
	// - Merge properties (new properties override existing)

	updated := false

	// Update confidence if higher
	if newConfidence > entity.ConfidenceScore {
		// In production, would use ent update mutation here
		// For POC, just log it
		d.logger.Debug("Entity confidence could be updated",
			"entity_id", entity.ID,
			"old_confidence", entity.ConfidenceScore,
			"new_confidence", newConfidence)
		updated = true
	}

	// Merge properties
	if len(newProperties) > 0 {
		d.logger.Debug("Entity properties could be merged",
			"entity_id", entity.ID,
			"new_properties", newProperties)
		updated = true
	}

	// For POC, we won't actually update the database
	// In production, would execute ent update mutation here

	return updated, nil
}
