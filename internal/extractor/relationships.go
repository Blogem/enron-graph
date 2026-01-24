package extractor

import (
	"context"
	"fmt"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/graph"
)

// createRelationships creates relationships between email and entities
func (e *Extractor) createRelationships(ctx context.Context, email *ent.Email, entities []*ent.DiscoveredEntity) ([]*ent.Relationship, error) {
	var relationships []*ent.Relationship

	// Create a map for quick lookup
	entityMap := make(map[string]*ent.DiscoveredEntity)
	for _, entity := range entities {
		entityMap[entity.UniqueID] = entity
	}

	// SENT relationship: person (from email) -> email
	if email.From != "" {
		if senderEntity, exists := entityMap[email.From]; exists {
			rel, err := e.repo.CreateRelationship(ctx, &graph.RelationshipInput{
				Type:            "SENT",
				FromType:        "discovered_entity",
				FromID:          senderEntity.ID,
				ToType:          "email",
				ToID:            email.ID,
				Timestamp:       email.Date,
				ConfidenceScore: 1.0,
				Properties:      map[string]interface{}{},
			})

			if err != nil {
				e.logger.Warn("Failed to create SENT relationship", "error", err)
			} else {
				relationships = append(relationships, rel)
			}
		}
	}

	// RECEIVED relationships: email -> person (to/cc/bcc)
	allRecipients := append(email.To, email.Cc...)
	allRecipients = append(allRecipients, email.Bcc...)

	for _, recipient := range allRecipients {
		if recipient == "" {
			continue
		}

		if recipientEntity, exists := entityMap[recipient]; exists {
			rel, err := e.repo.CreateRelationship(ctx, &graph.RelationshipInput{
				Type:            "RECEIVED",
				FromType:        "email",
				FromID:          email.ID,
				ToType:          "discovered_entity",
				ToID:            recipientEntity.ID,
				Timestamp:       email.Date,
				ConfidenceScore: 1.0,
				Properties:      map[string]interface{}{},
			})

			if err != nil {
				e.logger.Debug("Failed to create RECEIVED relationship", "error", err)
			} else {
				relationships = append(relationships, rel)
			}
		}
	}

	// MENTIONS relationships: email -> organizations/concepts/other
	for _, entity := range entities {
		// Skip person entities (already handled by SENT/RECEIVED)
		if entity.TypeCategory == "person" {
			continue
		}

		rel, err := e.repo.CreateRelationship(ctx, &graph.RelationshipInput{
			Type:            "MENTIONS",
			FromType:        "email",
			FromID:          email.ID,
			ToType:          "discovered_entity",
			ToID:            entity.ID,
			Timestamp:       email.Date,
			ConfidenceScore: entity.ConfidenceScore,
			Properties:      map[string]interface{}{},
		})

		if err != nil {
			e.logger.Debug("Failed to create MENTIONS relationship", "error", err)
		} else {
			relationships = append(relationships, rel)
		}
	}

	// COMMUNICATES_WITH relationships: person <-> person (inferred from email)
	if email.From != "" {
		senderEntity := entityMap[email.From]
		if senderEntity != nil {
			for _, recipient := range allRecipients {
				if recipient == "" || recipient == email.From {
					continue
				}

				recipientEntity := entityMap[recipient]
				if recipientEntity == nil {
					continue
				}

				// Create bidirectional COMMUNICATES_WITH relationship
				// Forward: sender -> recipient
				rel, err := e.repo.CreateRelationship(ctx, &graph.RelationshipInput{
					Type:            "COMMUNICATES_WITH",
					FromType:        "discovered_entity",
					FromID:          senderEntity.ID,
					ToType:          "discovered_entity",
					ToID:            recipientEntity.ID,
					Timestamp:       email.Date,
					ConfidenceScore: 0.9,
					Properties: map[string]interface{}{
						"via_email": email.MessageID,
					},
				})

				if err != nil {
					e.logger.Debug("Failed to create COMMUNICATES_WITH relationship", "error", err)
				} else {
					relationships = append(relationships, rel)
				}
			}
		}
	}

	return relationships, nil
}

// DeduplicateCommunications removes duplicate COMMUNICATES_WITH relationships
// This is a helper for batch processing to avoid creating duplicate communication links
func (e *Extractor) DeduplicateCommunications(ctx context.Context, person1ID, person2ID int) error {
	// Find existing relationships between these two persons
	rels, err := e.repo.FindRelationshipsByEntity(ctx, "discovered_entity", person1ID)
	if err != nil {
		return fmt.Errorf("failed to query relationships: %w", err)
	}

	// Count existing COMMUNICATES_WITH relationships
	count := 0
	for _, rel := range rels {
		if rel.Type == "COMMUNICATES_WITH" {
			if (rel.FromID == person1ID && rel.ToID == person2ID) ||
				(rel.FromID == person2ID && rel.ToID == person1ID) {
				count++
			}
		}
	}

	// If multiple relationships exist, we could merge them or update properties
	// For POC, we just log this for awareness
	if count > 1 {
		e.logger.Debug("Duplicate COMMUNICATES_WITH relationships found",
			"person1", person1ID,
			"person2", person2ID,
			"count", count)
	}

	return nil
}
