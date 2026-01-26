package explorer

import (
	"context"
	"sync"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/ent/schemapromotion"
)

type SchemaService struct {
	client *ent.Client
	mu     sync.RWMutex
	cache  *SchemaResponse
}

func NewSchemaService(client *ent.Client) *SchemaService {
	return &SchemaService{
		client: client,
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
	promotions, err := s.client.SchemaPromotion.
		Query().
		All(ctx)
	if err != nil {
		return nil, err
	}

	types := make([]SchemaType, 0, len(promotions))
	for _, p := range promotions {
		count, err := s.client.DiscoveredEntity.
			Query().
			Where(discoveredentity.TypeCategoryEQ(p.TypeName)).
			Count(ctx)
		if err != nil {
			continue
		}

		properties, err := s.extractProperties(ctx, p.TypeName)
		if err != nil {
			properties = []PropertyDefinition{}
		}

		types = append(types, SchemaType{
			Name:       p.TypeName,
			Count:      int64(count),
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
		isPromoted, _ := s.client.SchemaPromotion.
			Query().
			Where(schemapromotion.TypeNameEQ(r.TypeCategory)).
			Exist(ctx)

		// Skip promoted types - they're returned separately
		if isPromoted {
			continue
		}

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
	count, err := s.client.DiscoveredEntity.
		Query().
		Where(discoveredentity.TypeCategoryEQ(typeName)).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	isPromoted, _ := s.client.SchemaPromotion.
		Query().
		Where(schemapromotion.TypeNameEQ(typeName)).
		Exist(ctx)

	properties, err := s.extractProperties(ctx, typeName)
	if err != nil {
		properties = []PropertyDefinition{}
	}

	return &SchemaType{
		Name:       typeName,
		Count:      int64(count),
		IsPromoted: isPromoted,
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
