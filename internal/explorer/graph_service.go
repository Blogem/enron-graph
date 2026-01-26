package explorer

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	esql "entgo.io/ent/dialect/sql"
	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/ent/email"
	"github.com/Blogem/enron-graph/ent/relationship"
)

type GraphService struct {
	client *ent.Client
	db     *sql.DB
}

func NewGraphService(client *ent.Client, db *sql.DB) *GraphService {
	return &GraphService{
		client: client,
		db:     db,
	}
}

// GetRandomNodes returns exactly `limit` random nodes from all entity types with connecting edges
func (s *GraphService) GetRandomNodes(ctx context.Context, limit int) (*GraphResponse, error) {
	// Get total count of all entities
	totalCount, err := s.getTotalEntityCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total entity count: %w", err)
	}

	// Get random discovered entities
	discoveredEntities, err := s.client.DiscoveredEntity.
		Query().
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query random discovered entities: %w", err)
	}

	// Build nodes from discovered entities
	nodes := make([]GraphNode, 0, len(discoveredEntities))
	nodeIDs := make(map[string]bool)

	for _, entity := range discoveredEntities {
		node := GraphNode{
			ID:         entity.UniqueID,
			Type:       entity.TypeCategory,
			Category:   "discovered",
			Properties: entity.Properties,
			IsGhost:    false,
		}
		nodes = append(nodes, node)
		nodeIDs[entity.UniqueID] = true
	}

	// Get edges connecting these nodes
	edges, err := s.getEdgesBetweenNodes(ctx, discoveredEntities)
	if err != nil {
		return nil, fmt.Errorf("failed to get connecting edges: %w", err)
	}

	return &GraphResponse{
		Nodes:      nodes,
		Edges:      edges,
		TotalNodes: totalCount,
		HasMore:    len(nodes) < totalCount,
	}, nil
}

// GetNodes returns nodes filtered by type, category, and/or search query
func (s *GraphService) GetNodes(ctx context.Context, filter NodeFilter) (*GraphResponse, error) {
	log.Printf("[GetNodes] Starting with filter: Types=%v, Category=%s, SearchQuery=%q, Limit=%d",
		filter.Types, filter.Category, filter.SearchQuery, filter.Limit)

	// Build query based on filters
	query := s.client.DiscoveredEntity.Query()

	// Filter by types if specified
	if len(filter.Types) > 0 {
		query = query.Where(discoveredentity.TypeCategoryIn(filter.Types...))
	}

	// Filter by category if specified
	// Note: Currently all entities in DiscoveredEntity are "discovered" category
	// Promoted types (like Email) would need separate handling
	if filter.Category == "discovered" {
		// Already querying DiscoveredEntity, so this is default behavior
	} else if filter.Category == "promoted" {
		// For now, return empty results as we're only querying DiscoveredEntity
		// Future enhancement: query Email and other promoted types
		return &GraphResponse{
			Nodes:      []GraphNode{},
			Edges:      []GraphEdge{},
			TotalNodes: 0,
			HasMore:    false,
		}, nil
	}

	// Apply search query if specified
	if filter.SearchQuery != "" {
		// Use case-insensitive search across unique_id, name, type_category, and JSONB properties
		searchQuery := fmt.Sprintf("%%%s%%", filter.SearchQuery)
		query = query.Where(func(s *esql.Selector) {
			s.Where(esql.Or(
				esql.P(func(b *esql.Builder) {
					b.Ident(s.C(discoveredentity.FieldUniqueID))
					b.WriteString(" ILIKE ")
					b.Arg(searchQuery)
				}),
				esql.P(func(b *esql.Builder) {
					b.Ident(s.C(discoveredentity.FieldName))
					b.WriteString(" ILIKE ")
					b.Arg(searchQuery)
				}),
				esql.P(func(b *esql.Builder) {
					b.Ident(s.C(discoveredentity.FieldTypeCategory))
					b.WriteString(" ILIKE ")
					b.Arg(searchQuery)
				}),
				esql.P(func(b *esql.Builder) {
					b.WriteString("(")
					b.Ident(s.C(discoveredentity.FieldProperties))
					b.WriteString(" IS NOT NULL AND ")
					b.Ident(s.C(discoveredentity.FieldProperties))
					b.WriteString(" != 'null'::jsonb AND EXISTS (SELECT 1 FROM jsonb_each_text(")
					b.Ident(s.C(discoveredentity.FieldProperties))
					b.WriteString(") WHERE value ILIKE ")
					b.Arg(searchQuery)
					b.WriteString("))")
				}),
			))
		})
		log.Printf("[GetNodes] Applied search query: %q", filter.SearchQuery)
	}

	// Apply limit
	limit := filter.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}
	query = query.Limit(limit)

	// Execute query
	entities, err := query.All(ctx)
	if err != nil {
		log.Printf("[GetNodes] ERROR executing query: %v", err)
		return nil, fmt.Errorf("failed to query filtered nodes: %w", err)
	}
	log.Printf("[GetNodes] Query returned %d entities", len(entities))

	// Debug: log first few entities to see what we got
	for i, e := range entities {
		if i >= 3 {
			break
		}
		log.Printf("[GetNodes] Entity %d: ID=%q, Name=%q, Type=%q", i, e.UniqueID, e.Name, e.TypeCategory)
	}

	// Get total count (without limit)
	countQuery := s.client.DiscoveredEntity.Query()
	if len(filter.Types) > 0 {
		countQuery = countQuery.Where(discoveredentity.TypeCategoryIn(filter.Types...))
	}
	if filter.SearchQuery != "" {
		searchQuery := fmt.Sprintf("%%%s%%", filter.SearchQuery)
		countQuery = countQuery.Where(func(s *esql.Selector) {
			s.Where(esql.Or(
				esql.P(func(b *esql.Builder) {
					b.Ident(s.C(discoveredentity.FieldUniqueID))
					b.WriteString(" ILIKE ")
					b.Arg(searchQuery)
				}),
				esql.P(func(b *esql.Builder) {
					b.Ident(s.C(discoveredentity.FieldName))
					b.WriteString(" ILIKE ")
					b.Arg(searchQuery)
				}),
				esql.P(func(b *esql.Builder) {
					b.Ident(s.C(discoveredentity.FieldTypeCategory))
					b.WriteString(" ILIKE ")
					b.Arg(searchQuery)
				}),
				esql.P(func(b *esql.Builder) {
					b.WriteString("(")
					b.Ident(s.C(discoveredentity.FieldProperties))
					b.WriteString(" IS NOT NULL AND ")
					b.Ident(s.C(discoveredentity.FieldProperties))
					b.WriteString(" != 'null'::jsonb AND EXISTS (SELECT 1 FROM jsonb_each_text(")
					b.Ident(s.C(discoveredentity.FieldProperties))
					b.WriteString(") WHERE value ILIKE ")
					b.Arg(searchQuery)
					b.WriteString("))")
				}),
			))
		})
	}
	totalCount, err := countQuery.Count(ctx)
	if err != nil {
		log.Printf("[GetNodes] ERROR executing count query: %v", err)
		return nil, fmt.Errorf("failed to count filtered nodes: %w", err)
	}
	log.Printf("[GetNodes] Total count: %d", totalCount)

	// Build nodes from entities
	nodes := make([]GraphNode, 0, len(entities))
	nodeIDSet := make(map[string]bool)
	entityIDMap := make(map[int]string) // map from ent ID to unique_id

	for _, entity := range entities {
		node := GraphNode{
			ID:         entity.UniqueID,
			Type:       entity.TypeCategory,
			Category:   "discovered",
			Properties: entity.Properties,
			IsGhost:    false,
		}
		nodes = append(nodes, node)
		nodeIDSet[entity.UniqueID] = true
		entityIDMap[entity.ID] = entity.UniqueID
	}

	// T080a: Get edges from filtered nodes (including edges to nodes outside filter)
	edges, ghostNodes, err := s.getEdgesWithGhostNodes(ctx, entities, nodeIDSet)
	if err != nil {
		return nil, fmt.Errorf("failed to get edges: %w", err)
	}

	// T080b: Add ghost nodes to response
	nodes = append(nodes, ghostNodes...)

	return &GraphResponse{
		Nodes:      nodes,
		Edges:      edges,
		TotalNodes: totalCount,
		HasMore:    len(nodes) < totalCount,
	}, nil
}

// GetRelationships returns paginated relationships for a specific node
func (s *GraphService) GetRelationships(ctx context.Context, nodeID string, offset, limit int) (*RelationshipsResponse, error) {
	// Find the entity by unique_id
	entity, err := s.client.DiscoveredEntity.
		Query().
		Where(discoveredentity.UniqueIDEQ(nodeID)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find entity %s: %w", nodeID, err)
	}

	// Count total relationships for this node (both outgoing and incoming)
	totalOutgoing, err := s.client.Relationship.
		Query().
		Where(
			relationship.FromTypeEQ("discovered_entity"),
			relationship.FromIDEQ(entity.ID),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count outgoing relationships: %w", err)
	}

	totalIncoming, err := s.client.Relationship.
		Query().
		Where(
			relationship.ToTypeEQ("discovered_entity"),
			relationship.ToIDEQ(entity.ID),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count incoming relationships: %w", err)
	}

	totalCount := totalOutgoing + totalIncoming

	// Get paginated outgoing relationships
	outgoingRels, err := s.client.Relationship.
		Query().
		Where(
			relationship.FromTypeEQ("discovered_entity"),
			relationship.FromIDEQ(entity.ID),
		).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query outgoing relationships: %w", err)
	}

	// Get paginated incoming relationships if needed
	incomingRels := []*ent.Relationship{}
	remainingLimit := limit - len(outgoingRels)
	if remainingLimit > 0 {
		adjustedOffset := offset - totalOutgoing
		if adjustedOffset < 0 {
			adjustedOffset = 0
		}

		incomingRels, err = s.client.Relationship.
			Query().
			Where(
				relationship.ToTypeEQ("discovered_entity"),
				relationship.ToIDEQ(entity.ID),
			).
			Offset(adjustedOffset).
			Limit(remainingLimit).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query incoming relationships: %w", err)
		}
	}

	// Collect all connected entity IDs
	connectedEntityIDs := make(map[int]bool)
	for _, rel := range outgoingRels {
		if rel.ToType == "discovered_entity" {
			connectedEntityIDs[rel.ToID] = true
		}
	}
	for _, rel := range incomingRels {
		if rel.FromType == "discovered_entity" {
			connectedEntityIDs[rel.FromID] = true
		}
	}

	// Fetch connected entities
	ids := make([]int, 0, len(connectedEntityIDs))
	for id := range connectedEntityIDs {
		ids = append(ids, id)
	}

	connectedEntities, err := s.client.DiscoveredEntity.
		Query().
		Where(discoveredentity.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query connected entities: %w", err)
	}

	// Build entity map by ID for quick lookup
	entityMap := make(map[int]*ent.DiscoveredEntity)
	entityMap[entity.ID] = entity
	for _, e := range connectedEntities {
		entityMap[e.ID] = e
	}

	// Convert to GraphNodes
	nodes := make([]GraphNode, 0, len(connectedEntities))
	for _, e := range connectedEntities {
		nodes = append(nodes, GraphNode{
			ID:         e.UniqueID,
			Type:       e.TypeCategory,
			Category:   "discovered",
			Properties: e.Properties,
			IsGhost:    false,
		})
	}

	// Convert to GraphEdges
	edges := make([]GraphEdge, 0, len(outgoingRels)+len(incomingRels))
	for _, rel := range outgoingRels {
		fromEntity := entityMap[rel.FromID]
		toEntity := entityMap[rel.ToID]
		if fromEntity != nil && toEntity != nil {
			edges = append(edges, GraphEdge{
				Source: fromEntity.UniqueID,
				Target: toEntity.UniqueID,
				Type:   rel.Type,
			})
		}
	}
	for _, rel := range incomingRels {
		fromEntity := entityMap[rel.FromID]
		toEntity := entityMap[rel.ToID]
		if fromEntity != nil && toEntity != nil {
			edges = append(edges, GraphEdge{
				Source: fromEntity.UniqueID,
				Target: toEntity.UniqueID,
				Type:   rel.Type,
			})
		}
	}

	hasMore := offset+limit < totalCount

	return &RelationshipsResponse{
		Nodes:      nodes,
		Edges:      edges,
		TotalCount: totalCount,
		HasMore:    hasMore,
		Offset:     offset,
	}, nil
}

// GetNodeDetails returns complete information for a specific node
func (s *GraphService) GetNodeDetails(ctx context.Context, nodeID string) (*GraphNode, error) {
	// Try to find as discovered entity
	entity, err := s.client.DiscoveredEntity.
		Query().
		Where(discoveredentity.UniqueIDEQ(nodeID)).
		Only(ctx)
	if err == nil {
		// Count relationships
		degree, _ := s.getNodeDegree(ctx, entity.ID)

		return &GraphNode{
			ID:         entity.UniqueID,
			Type:       entity.TypeCategory,
			Category:   "discovered",
			Properties: entity.Properties,
			IsGhost:    false,
			Degree:     degree,
		}, nil
	}

	// Try to find as email (promoted type)
	emailEntity, err := s.client.Email.
		Query().
		Where(email.MessageIDEQ(nodeID)).
		Only(ctx)
	if err == nil {
		properties := map[string]interface{}{
			"message_id": emailEntity.MessageID,
			"subject":    emailEntity.Subject,
			"from":       emailEntity.From,
			"to":         emailEntity.To,
			"date":       emailEntity.Date,
		}

		return &GraphNode{
			ID:         emailEntity.MessageID,
			Type:       "email",
			Category:   "promoted",
			Properties: properties,
			IsGhost:    false,
		}, nil
	}

	return nil, fmt.Errorf("node %s does not exist", nodeID)
}

// Helper functions

func (s *GraphService) getTotalEntityCount(ctx context.Context) (int, error) {
	count, err := s.client.DiscoveredEntity.Query().Count(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *GraphService) getEdgesBetweenNodes(ctx context.Context, entities []*ent.DiscoveredEntity) ([]GraphEdge, error) {
	if len(entities) == 0 {
		return []GraphEdge{}, nil
	}

	// Build map of entity IDs to unique IDs
	entityIDToUniqueID := make(map[int]string)
	entityIDs := make([]int, 0, len(entities))
	for _, e := range entities {
		entityIDToUniqueID[e.ID] = e.UniqueID
		entityIDs = append(entityIDs, e.ID)
	}

	// Query relationships where both source and target are in the entity set
	rels, err := s.client.Relationship.
		Query().
		Where(
			relationship.FromTypeEQ("discovered_entity"),
			relationship.FromIDIn(entityIDs...),
			relationship.ToTypeEQ("discovered_entity"),
			relationship.ToIDIn(entityIDs...),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to GraphEdges
	edges := make([]GraphEdge, 0, len(rels))
	for _, rel := range rels {
		sourceID, sourceOK := entityIDToUniqueID[rel.FromID]
		targetID, targetOK := entityIDToUniqueID[rel.ToID]

		if sourceOK && targetOK {
			edges = append(edges, GraphEdge{
				Source: sourceID,
				Target: targetID,
				Type:   rel.Type,
			})
		}
	}

	return edges, nil
}

// getEdgesWithGhostNodes returns edges from filtered nodes including edges to nodes outside the filter
// Returns edges and ghost nodes for unmatched targets (FR-007a)
func (s *GraphService) getEdgesWithGhostNodes(ctx context.Context, entities []*ent.DiscoveredEntity, nodeIDSet map[string]bool) ([]GraphEdge, []GraphNode, error) {
	if len(entities) == 0 {
		return []GraphEdge{}, []GraphNode{}, nil
	}

	// Build map of entity IDs
	entityIDToUniqueID := make(map[int]string)
	entityIDs := make([]int, 0, len(entities))
	for _, e := range entities {
		entityIDToUniqueID[e.ID] = e.UniqueID
		entityIDs = append(entityIDs, e.ID)
	}

	// Query ALL relationships from filtered nodes (including to nodes outside filter)
	rels, err := s.client.Relationship.
		Query().
		Where(
			relationship.FromTypeEQ("discovered_entity"),
			relationship.FromIDIn(entityIDs...),
		).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Collect target entity IDs that are NOT in the filtered set
	ghostEntityIDs := make(map[int]bool)
	for _, rel := range rels {
		if rel.ToType == "discovered_entity" {
			targetUniqueID, ok := entityIDToUniqueID[rel.ToID]
			// If target is not in our filtered set, it's a ghost node
			if !ok {
				ghostEntityIDs[rel.ToID] = true
			} else if !nodeIDSet[targetUniqueID] {
				ghostEntityIDs[rel.ToID] = true
			}
		}
	}

	// Fetch ghost entities to get their basic info
	ghostNodes := make([]GraphNode, 0)
	if len(ghostEntityIDs) > 0 {
		ghostIDs := make([]int, 0, len(ghostEntityIDs))
		for id := range ghostEntityIDs {
			ghostIDs = append(ghostIDs, id)
		}

		ghostEntities, err := s.client.DiscoveredEntity.
			Query().
			Where(discoveredentity.IDIn(ghostIDs...)).
			All(ctx)
		if err != nil {
			return nil, nil, err
		}

		// Create ghost nodes
		for _, entity := range ghostEntities {
			ghostNodes = append(ghostNodes, GraphNode{
				ID:         entity.UniqueID,
				Type:       entity.TypeCategory,
				Category:   "discovered",
				Properties: map[string]interface{}{}, // Empty properties for ghost nodes
				IsGhost:    true,
			})
			// Add to maps for edge creation
			entityIDToUniqueID[entity.ID] = entity.UniqueID
		}
	}

	// Convert to GraphEdges
	edges := make([]GraphEdge, 0, len(rels))
	for _, rel := range rels {
		sourceID, sourceOK := entityIDToUniqueID[rel.FromID]
		targetID, targetOK := entityIDToUniqueID[rel.ToID]

		if sourceOK && targetOK {
			edges = append(edges, GraphEdge{
				Source: sourceID,
				Target: targetID,
				Type:   rel.Type,
			})
		}
	}

	return edges, ghostNodes, nil
}

func (s *GraphService) getNodeDegree(ctx context.Context, entityID int) (int, error) {
	outgoing, err := s.client.Relationship.
		Query().
		Where(
			relationship.FromTypeEQ("discovered_entity"),
			relationship.FromIDEQ(entityID),
		).
		Count(ctx)
	if err != nil {
		return 0, err
	}

	incoming, err := s.client.Relationship.
		Query().
		Where(
			relationship.ToTypeEQ("discovered_entity"),
			relationship.ToIDEQ(entityID),
		).
		Count(ctx)
	if err != nil {
		return 0, err
	}

	return outgoing + incoming, nil
}
