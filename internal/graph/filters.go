package graph

import (
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/ent/email"
	"github.com/Blogem/enron-graph/ent/relationship"
)

// FilterParams holds filtering parameters
type FilterParams struct {
	MinConfidence *float64   // Minimum confidence score (0.0-1.0)
	MaxConfidence *float64   // Maximum confidence score (0.0-1.0)
	StartDate     *time.Time // Start of date range
	EndDate       *time.Time // End of date range
	TypeCategory  *string    // Entity type filter
	Name          *string    // Name filter (partial match)
}

// Validate ensures filter parameters are valid
func (f *FilterParams) Validate() error {
	if f.MinConfidence != nil {
		if *f.MinConfidence < 0.0 {
			val := 0.0
			f.MinConfidence = &val
		}
		if *f.MinConfidence > 1.0 {
			val := 1.0
			f.MinConfidence = &val
		}
	}
	if f.MaxConfidence != nil {
		if *f.MaxConfidence < 0.0 {
			val := 0.0
			f.MaxConfidence = &val
		}
		if *f.MaxConfidence > 1.0 {
			val := 1.0
			f.MaxConfidence = &val
		}
	}
	return nil
}

// ApplyToDiscoveredEntityQuery applies filters to a DiscoveredEntity query
func (f *FilterParams) ApplyToDiscoveredEntityQuery(query *ent.DiscoveredEntityQuery) *ent.DiscoveredEntityQuery {
	if f.MinConfidence != nil {
		query = query.Where(discoveredentity.ConfidenceScoreGTE(*f.MinConfidence))
	}
	if f.MaxConfidence != nil {
		query = query.Where(discoveredentity.ConfidenceScoreLTE(*f.MaxConfidence))
	}
	if f.StartDate != nil {
		query = query.Where(discoveredentity.CreatedAtGTE(*f.StartDate))
	}
	if f.EndDate != nil {
		query = query.Where(discoveredentity.CreatedAtLTE(*f.EndDate))
	}
	if f.TypeCategory != nil {
		query = query.Where(discoveredentity.TypeCategoryEQ(*f.TypeCategory))
	}
	if f.Name != nil {
		query = query.Where(discoveredentity.NameContains(*f.Name))
	}
	return query
}

// ApplyToEmailQuery applies filters to an Email query
func (f *FilterParams) ApplyToEmailQuery(query *ent.EmailQuery) *ent.EmailQuery {
	if f.StartDate != nil {
		query = query.Where(email.DateGTE(*f.StartDate))
	}
	if f.EndDate != nil {
		query = query.Where(email.DateLTE(*f.EndDate))
	}
	return query
}

// ApplyToRelationshipQuery applies filters to a Relationship query
func (f *FilterParams) ApplyToRelationshipQuery(query *ent.RelationshipQuery) *ent.RelationshipQuery {
	if f.MinConfidence != nil {
		query = query.Where(relationship.ConfidenceScoreGTE(*f.MinConfidence))
	}
	if f.MaxConfidence != nil {
		query = query.Where(relationship.ConfidenceScoreLTE(*f.MaxConfidence))
	}
	if f.StartDate != nil {
		query = query.Where(relationship.TimestampGTE(*f.StartDate))
	}
	if f.EndDate != nil {
		query = query.Where(relationship.TimestampLTE(*f.EndDate))
	}
	return query
}

// CombinedQueryParams combines pagination and filtering
type CombinedQueryParams struct {
	Pagination PaginationParams
	Filters    FilterParams
}

// NewCombinedQueryParams creates default combined query parameters
func NewCombinedQueryParams() CombinedQueryParams {
	return CombinedQueryParams{
		Pagination: DefaultPaginationParams(),
		Filters:    FilterParams{},
	}
}

// Validate validates both pagination and filter parameters
func (c *CombinedQueryParams) Validate() error {
	if err := c.Pagination.Validate(); err != nil {
		return err
	}
	if err := c.Filters.Validate(); err != nil {
		return err
	}
	return nil
}

// ApplyToDiscoveredEntityQuery applies both pagination and filters to a query
func (c *CombinedQueryParams) ApplyToDiscoveredEntityQuery(query *ent.DiscoveredEntityQuery) *ent.DiscoveredEntityQuery {
	query = c.Filters.ApplyToDiscoveredEntityQuery(query)
	query = c.Pagination.ApplyToDiscoveredEntityQuery(query)
	return query
}

// ApplyToEmailQuery applies both pagination and filters to a query
func (c *CombinedQueryParams) ApplyToEmailQuery(query *ent.EmailQuery) *ent.EmailQuery {
	query = c.Filters.ApplyToEmailQuery(query)
	query = c.Pagination.ApplyToEmailQuery(query)
	return query
}

// ApplyToRelationshipQuery applies both pagination and filters to a query
func (c *CombinedQueryParams) ApplyToRelationshipQuery(query *ent.RelationshipQuery) *ent.RelationshipQuery {
	query = c.Filters.ApplyToRelationshipQuery(query)
	query = c.Pagination.ApplyToRelationshipQuery(query)
	return query
}
