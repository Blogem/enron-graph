package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// T059: Unit tests for pagination

// PaginationParams holds pagination configuration
type PaginationParams struct {
	Limit  int
	Offset int
}

// NewPaginationParams creates pagination params with defaults
func NewPaginationParams(limit, offset int) *PaginationParams {
	return &PaginationParams{
		Limit:  limit,
		Offset: offset,
	}
}

// Validate ensures pagination params are within acceptable bounds
func (p *PaginationParams) Validate() error {
	if p.Limit < 0 {
		return ErrInvalidLimit
	}
	if p.Offset < 0 {
		return ErrInvalidOffset
	}
	if p.Limit > MaxPageSize {
		return ErrLimitTooLarge
	}
	return nil
}

// ApplyDefaults sets default values if not specified
func (p *PaginationParams) ApplyDefaults() {
	if p.Limit == 0 {
		p.Limit = DefaultPageSize
	}
	if p.Limit > MaxPageSize {
		p.Limit = MaxPageSize
	}
}

// Constants for pagination
const (
	DefaultPageSize = 50
	MaxPageSize     = 1000
)

// Errors
var (
	ErrInvalidLimit  = &ValidationError{Field: "limit", Message: "limit must be non-negative"}
	ErrInvalidOffset = &ValidationError{Field: "offset", Message: "offset must be non-negative"}
	ErrLimitTooLarge = &ValidationError{Field: "limit", Message: "limit exceeds maximum page size"}
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// TestPaginationLimitOffsetCalculation tests basic pagination math
func TestPaginationLimitOffsetCalculation(t *testing.T) {
	testCases := []struct {
		name           string
		limit          int
		offset         int
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "standard pagination",
			limit:          10,
			offset:         0,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "second page",
			limit:          10,
			offset:         10,
			expectedLimit:  10,
			expectedOffset: 10,
		},
		{
			name:           "large offset",
			limit:          50,
			offset:         1000,
			expectedLimit:  50,
			expectedOffset: 1000,
		},
		{
			name:           "small limit",
			limit:          1,
			offset:         0,
			expectedLimit:  1,
			expectedOffset: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := NewPaginationParams(tc.limit, tc.offset)
			assert.Equal(t, tc.expectedLimit, params.Limit)
			assert.Equal(t, tc.expectedOffset, params.Offset)
		})
	}
}

// TestPaginationBoundaryConditions tests edge cases
func TestPaginationBoundaryConditions(t *testing.T) {
	testCases := []struct {
		name        string
		limit       int
		offset      int
		totalCount  int
		expectError bool
		description string
	}{
		{
			name:        "offset greater than total count",
			limit:       10,
			offset:      100,
			totalCount:  50,
			expectError: false, // Not an error, just returns empty results
			description: "offset beyond total should return empty",
		},
		{
			name:        "offset equals total count",
			limit:       10,
			offset:      50,
			totalCount:  50,
			expectError: false,
			description: "offset at boundary should return empty",
		},
		{
			name:        "negative offset",
			limit:       10,
			offset:      -1,
			totalCount:  100,
			expectError: true,
			description: "negative offset should error",
		},
		{
			name:        "negative limit",
			limit:       -1,
			offset:      0,
			totalCount:  100,
			expectError: true,
			description: "negative limit should error",
		},
		{
			name:        "zero offset with results",
			limit:       10,
			offset:      0,
			totalCount:  100,
			expectError: false,
			description: "zero offset is valid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := NewPaginationParams(tc.limit, tc.offset)
			err := params.Validate()

			if tc.expectError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// TestPaginationDefaults tests default value application
func TestPaginationDefaults(t *testing.T) {
	testCases := []struct {
		name          string
		inputLimit    int
		inputOffset   int
		expectedLimit int
	}{
		{
			name:          "zero limit uses default",
			inputLimit:    0,
			inputOffset:   0,
			expectedLimit: DefaultPageSize,
		},
		{
			name:          "explicit limit preserved",
			inputLimit:    25,
			inputOffset:   0,
			expectedLimit: 25,
		},
		{
			name:          "limit exceeds max capped to max",
			inputLimit:    5000,
			inputOffset:   0,
			expectedLimit: MaxPageSize,
		},
		{
			name:          "limit at max is preserved",
			inputLimit:    MaxPageSize,
			inputOffset:   0,
			expectedLimit: MaxPageSize,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := NewPaginationParams(tc.inputLimit, tc.inputOffset)
			params.ApplyDefaults()
			assert.Equal(t, tc.expectedLimit, params.Limit)
		})
	}
}

// TestCalculatePageNumber tests converting offset to page number
func TestCalculatePageNumber(t *testing.T) {
	testCases := []struct {
		name         string
		offset       int
		limit        int
		expectedPage int
	}{
		{
			name:         "first page",
			offset:       0,
			limit:        10,
			expectedPage: 1,
		},
		{
			name:         "second page",
			offset:       10,
			limit:        10,
			expectedPage: 2,
		},
		{
			name:         "third page",
			offset:       20,
			limit:        10,
			expectedPage: 3,
		},
		{
			name:         "partial page",
			offset:       15,
			limit:        10,
			expectedPage: 2, // offset 15 with limit 10 is still page 2
		},
		{
			name:         "large offset",
			offset:       1000,
			limit:        50,
			expectedPage: 21,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pageNumber := (tc.offset / tc.limit) + 1
			assert.Equal(t, tc.expectedPage, pageNumber)
		})
	}
}

// TestCalculateTotalPages tests total page count calculation
func TestCalculateTotalPages(t *testing.T) {
	testCases := []struct {
		name       string
		totalCount int
		limit      int
		expected   int
	}{
		{
			name:       "exact division",
			totalCount: 100,
			limit:      10,
			expected:   10,
		},
		{
			name:       "with remainder",
			totalCount: 105,
			limit:      10,
			expected:   11,
		},
		{
			name:       "less than one page",
			totalCount: 5,
			limit:      10,
			expected:   1,
		},
		{
			name:       "empty set",
			totalCount: 0,
			limit:      10,
			expected:   0,
		},
		{
			name:       "one item",
			totalCount: 1,
			limit:      10,
			expected:   1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			totalPages := 0
			if tc.totalCount > 0 {
				totalPages = (tc.totalCount + tc.limit - 1) / tc.limit
			}
			assert.Equal(t, tc.expected, totalPages)
		})
	}
}

// TestHasNextPage tests next page detection
func TestHasNextPage(t *testing.T) {
	testCases := []struct {
		name        string
		offset      int
		limit       int
		totalCount  int
		expectNext  bool
		description string
	}{
		{
			name:        "has next page",
			offset:      0,
			limit:       10,
			totalCount:  100,
			expectNext:  true,
			description: "first page of many",
		},
		{
			name:        "last page",
			offset:      90,
			limit:       10,
			totalCount:  100,
			expectNext:  false,
			description: "at last page",
		},
		{
			name:        "beyond last page",
			offset:      100,
			limit:       10,
			totalCount:  100,
			expectNext:  false,
			description: "offset beyond total",
		},
		{
			name:        "partial last page",
			offset:      95,
			limit:       10,
			totalCount:  100,
			expectNext:  false,
			description: "partial results on last page",
		},
		{
			name:        "empty result set",
			offset:      0,
			limit:       10,
			totalCount:  0,
			expectNext:  false,
			description: "no results",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasNext := tc.offset+tc.limit < tc.totalCount
			assert.Equal(t, tc.expectNext, hasNext, tc.description)
		})
	}
}

// TestHasPreviousPage tests previous page detection
func TestHasPreviousPage(t *testing.T) {
	testCases := []struct {
		name        string
		offset      int
		expectPrev  bool
		description string
	}{
		{
			name:        "first page",
			offset:      0,
			expectPrev:  false,
			description: "no previous from first page",
		},
		{
			name:        "second page",
			offset:      10,
			expectPrev:  true,
			description: "has previous from second page",
		},
		{
			name:        "middle page",
			offset:      50,
			expectPrev:  true,
			description: "has previous from middle page",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasPrev := tc.offset > 0
			assert.Equal(t, tc.expectPrev, hasPrev, tc.description)
		})
	}
}

// TestPaginationResponse tests the response structure
func TestPaginationResponse(t *testing.T) {
	type PaginationResponse struct {
		Items      []interface{} `json:"items"`
		Total      int           `json:"total"`
		Limit      int           `json:"limit"`
		Offset     int           `json:"offset"`
		HasNext    bool          `json:"has_next"`
		HasPrev    bool          `json:"has_prev"`
		TotalPages int           `json:"total_pages"`
	}

	testCases := []struct {
		name     string
		offset   int
		limit    int
		total    int
		expected PaginationResponse
	}{
		{
			name:   "first page",
			offset: 0,
			limit:  10,
			total:  100,
			expected: PaginationResponse{
				Total:      100,
				Limit:      10,
				Offset:     0,
				HasNext:    true,
				HasPrev:    false,
				TotalPages: 10,
			},
		},
		{
			name:   "middle page",
			offset: 50,
			limit:  10,
			total:  100,
			expected: PaginationResponse{
				Total:      100,
				Limit:      10,
				Offset:     50,
				HasNext:    true,
				HasPrev:    true,
				TotalPages: 10,
			},
		},
		{
			name:   "last page",
			offset: 90,
			limit:  10,
			total:  100,
			expected: PaginationResponse{
				Total:      100,
				Limit:      10,
				Offset:     90,
				HasNext:    false,
				HasPrev:    true,
				TotalPages: 10,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			totalPages := 0
			if tc.total > 0 {
				totalPages = (tc.total + tc.limit - 1) / tc.limit
			}

			response := PaginationResponse{
				Items:      []interface{}{}, // Would be populated with actual data
				Total:      tc.total,
				Limit:      tc.limit,
				Offset:     tc.offset,
				HasNext:    tc.offset+tc.limit < tc.total,
				HasPrev:    tc.offset > 0,
				TotalPages: totalPages,
			}

			assert.Equal(t, tc.expected.Total, response.Total)
			assert.Equal(t, tc.expected.Limit, response.Limit)
			assert.Equal(t, tc.expected.Offset, response.Offset)
			assert.Equal(t, tc.expected.HasNext, response.HasNext)
			assert.Equal(t, tc.expected.HasPrev, response.HasPrev)
			assert.Equal(t, tc.expected.TotalPages, response.TotalPages)
		})
	}
}
