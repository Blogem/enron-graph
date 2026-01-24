package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// T059: Unit tests for pagination

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
			params := PaginationParams{
				Limit:  tc.limit,
				Offset: tc.offset,
			}
			params.Validate()
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
		description string
	}{
		{
			name:        "offset greater than total count",
			limit:       10,
			offset:      100,
			description: "offset beyond total should be valid",
		},
		{
			name:        "offset equals total count",
			limit:       10,
			offset:      50,
			description: "offset at boundary should be valid",
		},
		{
			name:        "negative offset gets corrected",
			limit:       10,
			offset:      -1,
			description: "negative offset should be corrected to 0",
		},
		{
			name:        "negative limit gets corrected",
			limit:       -1,
			offset:      0,
			description: "negative limit should be corrected to default",
		},
		{
			name:        "zero offset with results",
			limit:       10,
			offset:      0,
			description: "zero offset is valid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := PaginationParams{
				Limit:  tc.limit,
				Offset: tc.offset,
			}
			err := params.Validate()
			assert.NoError(t, err, tc.description)

			// After validation, limits should be within bounds
			assert.GreaterOrEqual(t, params.Limit, 1)
			assert.LessOrEqual(t, params.Limit, 1000)
			assert.GreaterOrEqual(t, params.Offset, 0)
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
			expectedLimit: 100,
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
			expectedLimit: 1000,
		},
		{
			name:          "limit at max is preserved",
			inputLimit:    1000,
			inputOffset:   0,
			expectedLimit: 1000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := PaginationParams{
				Limit:  tc.inputLimit,
				Offset: tc.inputOffset,
			}
			params.Validate()
			assert.Equal(t, tc.expectedLimit, params.Limit)
		})
	}
}

// TestPaginatedResult tests paginated result creation
func TestPaginatedResult(t *testing.T) {
	testCases := []struct {
		name        string
		total       int
		limit       int
		offset      int
		hasMore     bool
		description string
	}{
		{
			name:        "first page with more",
			total:       100,
			limit:       10,
			offset:      0,
			hasMore:     true,
			description: "first page should have more",
		},
		{
			name:        "last page no more",
			total:       100,
			limit:       10,
			offset:      90,
			hasMore:     false,
			description: "last page should not have more",
		},
		{
			name:        "single page",
			total:       5,
			limit:       10,
			offset:      0,
			hasMore:     false,
			description: "single page should not have more",
		},
		{
			name:        "offset beyond total",
			total:       50,
			limit:       10,
			offset:      100,
			hasMore:     false,
			description: "offset beyond total should not have more",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NewPaginatedResult(tc.total, tc.limit, tc.offset)
			assert.Equal(t, tc.total, result.Total)
			assert.Equal(t, tc.limit, result.Limit)
			assert.Equal(t, tc.offset, result.Offset)
			assert.Equal(t, tc.hasMore, result.HasMore, tc.description)
		})
	}
}
