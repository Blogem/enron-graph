package explorer

import (
	"context"
	"database/sql"
	"sync"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
)

type SchemaService struct {
	client *ent.Client
	db     *sql.DB
	mu     sync.RWMutex
	cache  *SchemaResponse
}

func NewSchemaService(client *ent.Client, db *sql.DB) *SchemaService {
	return &SchemaService{
		client: client,
		db:     db,
	}
}

func (s *SchemaService) GetSchema(ctx context.Context) (*SchemaResponse, error) {
	s.mu.RLock()
	if s.cache != nil {
		defer s.mu.RUnlock()
		return s.cache, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache != nil {
		return s.cache, nil
	}

	promotedTypes, err := s.getPromotedTypes(ctx)
	if err != nil {
		return nil, err
	}

	discoveredTypes, err := s.getDiscoveredTypes(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate total entities
	totalEntities := 0
	for _, pt := range promotedTypes {
		totalEntities += int(pt.Count)
	}
	for _, dt := range discoveredTypes {
		totalEntities += int(dt.Count)
	}

	s.cache = &SchemaResponse{
		PromotedTypes:   promotedTypes,
		DiscoveredTypes: discoveredTypes,
		TotalEntities:   totalEntities,
	}

	return s.cache, nil
}

func (s *SchemaService) getPromotedTypes(ctx context.Context) ([]SchemaType, error) {
	// Query PostgreSQL information_schema for all tables (excluding our meta tables)
	query := `
		SELECT 
			t.table_name,
			COALESCE(pg_stat.n_live_tup, 0) as row_count
		FROM information_schema.tables t
		LEFT JOIN pg_stat_user_tables pg_stat ON pg_stat.relname = t.table_name
		WHERE t.table_schema = 'public'
		AND t.table_type = 'BASE TABLE'
		AND t.table_name NOT IN ('relationships', 'discovered_entities', 'schema_promotions')
		ORDER BY t.table_name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := make([]SchemaType, 0)
	for rows.Next() {
		var tableName string
		var rowCount int64
		if err := rows.Scan(&tableName, &rowCount); err != nil {
			continue
		}

		// Get column information
		columnQuery := `
			SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_schema = 'public'
			AND table_name = $1
			ORDER BY ordinal_position
			LIMIT 10
		`

		columnRows, err := s.db.QueryContext(ctx, columnQuery, tableName)
		if err != nil {
			continue
		}

		properties := []PropertyDefinition{}
		for columnRows.Next() {
			var colName, dataType, nullable string
			if err := columnRows.Scan(&colName, &dataType, &nullable); err != nil {
				continue
			}
			properties = append(properties, PropertyDefinition{
				Name:     colName,
				Type:     dataType,
				Nullable: nullable == "YES",
			})
		}
		columnRows.Close()

		types = append(types, SchemaType{
			Name:       tableName,
			Count:      rowCount,
			IsPromoted: true,
			Properties: properties,
		})
	}

	return types, nil
}

func (s *SchemaService) getDiscoveredTypes(ctx context.Context) ([]SchemaType, error) {
	var results []struct {
		TypeCategory string `sql:"type_category"`
		Count        int    `sql:"count"`
	}

	err := s.client.DiscoveredEntity.
		Query().
		GroupBy(discoveredentity.FieldTypeCategory).
		Aggregate(ent.Count()).
		Scan(ctx, &results)
	if err != nil {
		return nil, err
	}

	types := make([]SchemaType, 0, len(results))
	for _, r := range results {
		properties, err := s.extractProperties(ctx, r.TypeCategory)
		if err != nil {
			properties = []PropertyDefinition{}
		}

		types = append(types, SchemaType{
			Name:       r.TypeCategory,
			Count:      int64(r.Count),
			IsPromoted: false,
			Properties: properties,
		})
	}

	return types, nil
}

func (s *SchemaService) extractProperties(ctx context.Context, typeCategory string) ([]PropertyDefinition, error) {
	entities, err := s.client.DiscoveredEntity.
		Query().
		Where(discoveredentity.TypeCategoryEQ(typeCategory)).
		Limit(10).
		All(ctx)
	if err != nil {
		return nil, err
	}

	propMap := make(map[string]*PropertyDefinition)

	for _, entity := range entities {
		propMap["name"] = &PropertyDefinition{
			Name:         "name",
			Type:         "string",
			SampleValues: []string{entity.Name},
		}

		propMap["confidence_score"] = &PropertyDefinition{
			Name:         "confidence_score",
			Type:         "float",
			SampleValues: []string{},
		}
	}

	properties := make([]PropertyDefinition, 0, len(propMap))
	for _, prop := range propMap {
		properties = append(properties, *prop)
	}

	return properties, nil
}

func (s *SchemaService) GetTypeDetails(ctx context.Context, typeName string) (*SchemaType, error) {
	// Check if this is a promoted type by checking if a table exists
	var tableExists bool
	tableQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
			AND table_type = 'BASE TABLE'
			AND table_name NOT IN ('relationships', 'discovered_entities', 'schema_promotions')
		)
	`
	err := s.db.QueryRowContext(ctx, tableQuery, typeName).Scan(&tableExists)
	if err != nil {
		return nil, err
	}

	var count int64
	var properties []PropertyDefinition

	if tableExists {
		// Get count from the actual table
		countQuery := `SELECT COUNT(*) FROM ` + typeName
		err = s.db.QueryRowContext(ctx, countQuery).Scan(&count)
		if err != nil {
			return nil, err
		}

		// Get column information
		columnQuery := `
			SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_schema = 'public'
			AND table_name = $1
			ORDER BY ordinal_position
		`
		rows, err := s.db.QueryContext(ctx, columnQuery, typeName)
		if err != nil {
			properties = []PropertyDefinition{}
		} else {
			defer rows.Close()
			properties = []PropertyDefinition{}
			for rows.Next() {
				var colName, dataType, nullable string
				if err := rows.Scan(&colName, &dataType, &nullable); err != nil {
					continue
				}
				properties = append(properties, PropertyDefinition{
					Name:     colName,
					Type:     dataType,
					Nullable: nullable == "YES",
				})
			}
		}
	} else {
		// Get count from discovered_entities
		intCount, err := s.client.DiscoveredEntity.
			Query().
			Where(discoveredentity.TypeCategoryEQ(typeName)).
			Count(ctx)
		if err != nil {
			return nil, err
		}
		count = int64(intCount)

		properties, err = s.extractProperties(ctx, typeName)
		if err != nil {
			properties = []PropertyDefinition{}
		}
	}

	return &SchemaType{
		Name:       typeName,
		Count:      count,
		IsPromoted: tableExists,
		Properties: properties,
	}, nil
}

func (s *SchemaService) getDiscoveredTypeDetails(ctx context.Context, typeName string) (*SchemaType, error) {
	count, err := s.client.DiscoveredEntity.
		Query().
		Where(discoveredentity.TypeCategoryEQ(typeName)).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	properties, err := s.extractProperties(ctx, typeName)
	if err != nil {
		properties = []PropertyDefinition{}
	}

	return &SchemaType{
		Name:       typeName,
		Count:      int64(count),
		IsPromoted: false,
		Properties: properties,
	}, nil
}

func (s *SchemaService) RefreshSchema(ctx context.Context) error {
	s.mu.Lock()
	s.cache = nil
	s.mu.Unlock()

	_, err := s.GetSchema(ctx)
	return err
}
