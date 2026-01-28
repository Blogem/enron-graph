package explorer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	esql "entgo.io/ent/dialect/sql"
	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/ent/email"
	"github.com/Blogem/enron-graph/ent/relationship"
	"github.com/Blogem/enron-graph/pkg/llm"
)

type GraphService struct {
	client    *ent.Client
	db        *sql.DB
	llmClient llm.Client
}

func NewGraphService(client *ent.Client, db *sql.DB, llmClient llm.Client) *GraphService {
	return &GraphService{
		client:    client,
		db:        db,
		llmClient: llmClient,
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
		// Ensure properties map exists and include the name field
		props := entity.Properties
		if props == nil {
			props = make(map[string]interface{})
		}
		// Add name to properties if it exists
		if entity.Name != "" {
			props["name"] = entity.Name
		}

		node := GraphNode{
			ID:         entity.UniqueID,
			Type:       entity.TypeCategory,
			Category:   "discovered",
			Properties: props,
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
	if filter.Category == "discovered" {
		// Already querying DiscoveredEntity, so this is default behavior
	} else if filter.Category == "promoted" {
		// Query promoted types (e.g., Email table)
		return s.getPromotedNodes(ctx, filter)
	}

	// Apply limit
	limit := filter.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}

	var entities []*ent.DiscoveredEntity
	var err error

	// Apply search query if specified - try text search first, fallback to semantic if no results
	if filter.SearchQuery != "" {
		log.Printf("[GetNodes] Using text search for query: %q", filter.SearchQuery)

		// Execute text search first
		textQuery := s.client.DiscoveredEntity.Query()
		if len(filter.Types) > 0 {
			textQuery = textQuery.Where(discoveredentity.TypeCategoryIn(filter.Types...))
		}
		textQuery = s.applyTextSearch(textQuery, filter.SearchQuery)
		textQuery = textQuery.Limit(limit)
		entities, err = textQuery.All(ctx)

		// If text search returned no results and we have LLM client, try semantic search
		if err == nil && len(entities) == 0 && s.llmClient != nil {
			log.Printf("[GetNodes] Text search returned no results, falling back to semantic search")

			// Generate embedding for semantic search
			queryEmbedding, embErr := s.llmClient.GenerateEmbedding(ctx, filter.SearchQuery)
			if embErr != nil {
				log.Printf("[GetNodes] Failed to generate embedding: %v", embErr)
			} else {
				embeddingJSON, marshalErr := json.Marshal(queryEmbedding)
				if marshalErr != nil {
					log.Printf("[GetNodes] Failed to marshal embedding: %v", marshalErr)
				} else {
					// Execute semantic search query
					semanticQuery := s.client.DiscoveredEntity.Query()
					if len(filter.Types) > 0 {
						semanticQuery = semanticQuery.Where(discoveredentity.TypeCategoryIn(filter.Types...))
					}
					semanticQuery = semanticQuery.Where(func(s *esql.Selector) {
						s.Where(esql.P(func(b *esql.Builder) {
							b.Ident(s.C(discoveredentity.FieldEmbedding))
							b.WriteString(" IS NOT NULL AND ")
							b.Ident(s.C(discoveredentity.FieldEmbedding))
							b.WriteString(" != 'null'::jsonb AND ")
							b.WriteString("(1 - (")
							b.Ident(s.C(discoveredentity.FieldEmbedding))
							b.WriteString("::text::vector <=> '")
							b.WriteString(string(embeddingJSON))
							b.WriteString("'::vector)) > 0.3")
						}))
					})
					semanticQuery = semanticQuery.Order(func(s *esql.Selector) {
						s.OrderExpr(esql.Expr(fmt.Sprintf(
							"%s::text::vector <=> '%s'::vector",
							s.C(discoveredentity.FieldEmbedding),
							string(embeddingJSON),
						)))
					})
					semanticQuery = semanticQuery.Limit(limit)
					entities, err = semanticQuery.All(ctx)

					if err == nil {
						log.Printf("[GetNodes] Semantic search returned %d results", len(entities))
					}
				}
			}
		}
	} else {
		// No search query - just apply limit
		query = query.Limit(limit)
		entities, err = query.All(ctx)
	}
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
		// Ensure properties map exists and include the name field
		props := entity.Properties
		if props == nil {
			props = make(map[string]interface{})
		}
		// Add name to properties if it exists
		if entity.Name != "" {
			props["name"] = entity.Name
		}

		node := GraphNode{
			ID:         entity.UniqueID,
			Type:       entity.TypeCategory,
			Category:   "discovered",
			Properties: props,
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
			relationship.FromIDEQ(entity.ID),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count outgoing relationships: %w", err)
	}

	totalIncoming, err := s.client.Relationship.
		Query().
		Where(
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
		connectedEntityIDs[rel.ToID] = true
	}
	for _, rel := range incomingRels {
		connectedEntityIDs[rel.FromID] = true
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
		// Ensure properties map exists and include the name field
		props := e.Properties
		if props == nil {
			props = make(map[string]interface{})
		}
		if e.Name != "" {
			props["name"] = e.Name
		}

		nodes = append(nodes, GraphNode{
			ID:         e.UniqueID,
			Type:       e.TypeCategory,
			Category:   "discovered",
			Properties: props,
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

		// Ensure properties map exists and include the name field
		props := entity.Properties
		if props == nil {
			props = make(map[string]interface{})
		}
		if entity.Name != "" {
			props["name"] = entity.Name
		}

		return &GraphNode{
			ID:         entity.UniqueID,
			Type:       entity.TypeCategory,
			Category:   "discovered",
			Properties: props,
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
			relationship.FromIDIn(entityIDs...),
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
			relationship.FromIDIn(entityIDs...),
		).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Collect target entity IDs that are NOT in the filtered set
	ghostEntityIDs := make(map[int]bool)
	for _, rel := range rels {
		targetUniqueID, ok := entityIDToUniqueID[rel.ToID]
		// If target is not in our filtered set, it's a ghost node
		if !ok {
			ghostEntityIDs[rel.ToID] = true
		} else if !nodeIDSet[targetUniqueID] {
			ghostEntityIDs[rel.ToID] = true
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
			// For ghost nodes, still include name if available
			props := map[string]interface{}{}
			if entity.Name != "" {
				props["name"] = entity.Name
			}
			props["is_ghost"] = true // Mark as ghost in properties too

			ghostNodes = append(ghostNodes, GraphNode{
				ID:         entity.UniqueID,
				Type:       entity.TypeCategory,
				Category:   "discovered",
				Properties: props,
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
			relationship.FromIDEQ(entityID),
		).
		Count(ctx)
	if err != nil {
		return 0, err
	}

	incoming, err := s.client.Relationship.
		Query().
		Where(
			relationship.ToIDEQ(entityID),
		).
		Count(ctx)
	if err != nil {
		return 0, err
	}

	return outgoing + incoming, nil
}

// getPromotedNodes returns nodes from promoted types (dynamically discovered from database tables)
func (s *GraphService) getPromotedNodes(ctx context.Context, filter NodeFilter) (*GraphResponse, error) {
	nodes := []GraphNode{}
	edges := []GraphEdge{}
	totalCount := 0

	limit := filter.Limit
	if limit == 0 {
		limit = 100
	}

	// Dynamically discover promoted type tables (same logic as SchemaService)
	tableQuery := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
		AND table_name NOT IN ('relationships', 'discovered_entities', 'schema_promotions')
		ORDER BY table_name
	`

	rows, err := s.db.QueryContext(ctx, tableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query promoted tables: %w", err)
	}
	defer rows.Close()

	tableNames := []string{}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tableNames = append(tableNames, tableName)
	}

	// Query each promoted table
	remainingLimit := limit
	for _, tableName := range tableNames {
		if remainingLimit <= 0 {
			break
		}

		// Skip if type filter is specified and this table is not in the list
		if len(filter.Types) > 0 {
			found := false
			for _, t := range filter.Types {
				// Match table name (case-insensitive)
				if tableName == t || tableName+"s" == t || tableName == t+"s" {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Query the table for nodes
		tableNodes, err := s.queryPromotedTable(ctx, tableName, remainingLimit, filter.SearchQuery)
		if err != nil {
			log.Printf("[getPromotedNodes] Error querying table %s: %v", tableName, err)
			continue
		}

		nodes = append(nodes, tableNodes...)
		remainingLimit -= len(tableNodes)

		// Get count for this table
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		var count int
		if err := s.db.QueryRowContext(ctx, countQuery).Scan(&count); err == nil {
			totalCount += count
		}
	}

	return &GraphResponse{
		Nodes:      nodes,
		Edges:      edges,
		TotalNodes: totalCount,
		HasMore:    totalCount > len(nodes),
	}, nil
}

// queryPromotedTable queries a specific promoted table and returns GraphNodes
func (s *GraphService) queryPromotedTable(ctx context.Context, tableName string, limit int, searchQuery string) ([]GraphNode, error) {
	// Get column names for this table
	columnQuery := `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = $1
		ORDER BY ordinal_position
	`

	columnRows, err := s.db.QueryContext(ctx, columnQuery, tableName)
	if err != nil {
		return nil, err
	}
	defer columnRows.Close()

	columns := []string{}
	var idColumn string
	for columnRows.Next() {
		var colName string
		if err := columnRows.Scan(&colName); err != nil {
			continue
		}
		columns = append(columns, colName)
		// Try to find ID column
		if idColumn == "" && (colName == "id" || colName == "message_id" || colName == "unique_id") {
			idColumn = colName
		}
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("no columns found for table %s", tableName)
	}

	// Default to first column if no ID column found
	if idColumn == "" {
		idColumn = columns[0]
	}

	// Build column list (exclude ID since we're selecting it separately)
	otherColumns := []string{}
	for _, col := range columns {
		if col != idColumn {
			otherColumns = append(otherColumns, col)
		}
	}

	// Build SELECT query
	selectCols := idColumn
	if len(otherColumns) > 0 {
		selectCols += ", " + strings.Join(otherColumns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", selectCols, tableName)

	// Add search filter if specified
	if searchQuery != "" {
		// Simple text search across all columns
		whereClauses := []string{}
		for _, col := range columns {
			whereClauses = append(whereClauses, fmt.Sprintf("CAST(%s AS TEXT) ILIKE $1", col))
		}
		query += " WHERE (" + strings.Join(whereClauses, " OR ") + ")"
	}

	query += fmt.Sprintf(" LIMIT %d", limit)

	var queryRows *sql.Rows
	if searchQuery != "" {
		searchPattern := "%" + searchQuery + "%"
		queryRows, err = s.db.QueryContext(ctx, query, searchPattern)
	} else {
		queryRows, err = s.db.QueryContext(ctx, query)
	}

	if err != nil {
		return nil, err
	}
	defer queryRows.Close()

	nodes := []GraphNode{}
	for queryRows.Next() {
		// Prepare scan destinations
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(values))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := queryRows.Scan(valuePtrs...); err != nil {
			continue
		}

		// Extract ID (first value)
		var nodeID string
		if values[0] != nil {
			nodeID = fmt.Sprintf("%v", values[0])
		}

		// Build properties map
		properties := make(map[string]interface{})
		for i, col := range columns {
			if values[i] != nil {
				properties[col] = values[i]
			}
		}

		nodes = append(nodes, GraphNode{
			ID:         nodeID,
			Type:       tableName,
			Category:   "promoted",
			Properties: properties,
			IsGhost:    false,
		})
	}

	return nodes, nil
}

// applyTextSearch applies case-insensitive text search across multiple fields
func (s *GraphService) applyTextSearch(query *ent.DiscoveredEntityQuery, searchText string) *ent.DiscoveredEntityQuery {
	searchQuery := fmt.Sprintf("%%%s%%", searchText)
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
	log.Printf("[GetNodes] Applied text search query: %q", searchText)
	return query
}

// mergeAndDeduplicate combines semantic and text search results, removing duplicates
// Prioritizes semantic search results (they appear first)
func (s *GraphService) mergeAndDeduplicate(semanticResults, textResults []*ent.DiscoveredEntity, limit int) []*ent.DiscoveredEntity {
	seen := make(map[string]bool)
	merged := make([]*ent.DiscoveredEntity, 0, limit)

	// Add semantic results first (higher priority)
	for _, entity := range semanticResults {
		if !seen[entity.UniqueID] {
			seen[entity.UniqueID] = true
			merged = append(merged, entity)
			if len(merged) >= limit {
				return merged
			}
		}
	}

	// Add text results (fill remaining slots)
	for _, entity := range textResults {
		if !seen[entity.UniqueID] {
			seen[entity.UniqueID] = true
			merged = append(merged, entity)
			if len(merged) >= limit {
				return merged
			}
		}
	}

	return merged
}
