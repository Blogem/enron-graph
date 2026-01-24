package graph

import (
	"github.com/Blogem/enron-graph/ent"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Limit  int // Maximum number of results to return
	Offset int // Number of results to skip
}

// DefaultPaginationParams returns sensible defaults
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Limit:  100,
		Offset: 0,
	}
}

// Validate ensures pagination parameters are within acceptable ranges
func (p *PaginationParams) Validate() error {
	if p.Limit < 1 {
		p.Limit = 100
	}
	if p.Limit > 1000 {
		p.Limit = 1000
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return nil
}

// ApplyToDiscoveredEntityQuery applies pagination to a DiscoveredEntity query
func (p *PaginationParams) ApplyToDiscoveredEntityQuery(query *ent.DiscoveredEntityQuery) *ent.DiscoveredEntityQuery {
	return query.Limit(p.Limit).Offset(p.Offset)
}

// ApplyToEmailQuery applies pagination to an Email query
func (p *PaginationParams) ApplyToEmailQuery(query *ent.EmailQuery) *ent.EmailQuery {
	return query.Limit(p.Limit).Offset(p.Offset)
}

// ApplyToRelationshipQuery applies pagination to a Relationship query
func (p *PaginationParams) ApplyToRelationshipQuery(query *ent.RelationshipQuery) *ent.RelationshipQuery {
	return query.Limit(p.Limit).Offset(p.Offset)
}

// PaginatedResult represents a paginated result set
type PaginatedResult struct {
	Total   int  // Total count of results (without pagination)
	Limit   int  // Limit used
	Offset  int  // Offset used
	HasMore bool // Whether there are more results
}

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult(total, limit, offset int) PaginatedResult {
	return PaginatedResult{
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: offset+limit < total,
	}
}
