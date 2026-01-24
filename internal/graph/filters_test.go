package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// T060: Unit tests for filters

// TestConfidenceScoreFiltering tests filtering by confidence score
func TestConfidenceScoreFiltering(t *testing.T) {
	testCases := []struct {
		name          string
		minConfidence *float64
		maxConfidence *float64
		expectValid   bool
		description   string
	}{
		{
			name:          "valid range",
			minConfidence: float64Ptr(0.7),
			maxConfidence: float64Ptr(0.9),
			expectValid:   true,
			description:   "normal confidence range",
		},
		{
			name:          "min only",
			minConfidence: float64Ptr(0.5),
			maxConfidence: nil,
			expectValid:   true,
			description:   "min confidence only",
		},
		{
			name:          "max only",
			minConfidence: nil,
			maxConfidence: float64Ptr(0.8),
			expectValid:   true,
			description:   "max confidence only",
		},
		{
			name:          "no filters",
			minConfidence: nil,
			maxConfidence: nil,
			expectValid:   true,
			description:   "no confidence filters",
		},
		{
			name:          "min below zero gets corrected",
			minConfidence: float64Ptr(-0.5),
			maxConfidence: nil,
			expectValid:   true,
			description:   "negative min should be corrected",
		},
		{
			name:          "max above one gets corrected",
			minConfidence: nil,
			maxConfidence: float64Ptr(1.5),
			expectValid:   true,
			description:   "max above 1 should be corrected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := FilterParams{
				MinConfidence: tc.minConfidence,
				MaxConfidence: tc.maxConfidence,
			}

			err := filters.Validate()
			assert.NoError(t, err, tc.description)

			// After validation, values should be in range
			if filters.MinConfidence != nil {
				assert.GreaterOrEqual(t, *filters.MinConfidence, 0.0)
				assert.LessOrEqual(t, *filters.MinConfidence, 1.0)
			}
			if filters.MaxConfidence != nil {
				assert.GreaterOrEqual(t, *filters.MaxConfidence, 0.0)
				assert.LessOrEqual(t, *filters.MaxConfidence, 1.0)
			}
		})
	}
}

// TestDateRangeFiltering tests filtering by date range
func TestDateRangeFiltering(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	lastWeek := now.AddDate(0, 0, -7)
	nextWeek := now.AddDate(0, 0, 7)

	testCases := []struct {
		name        string
		startDate   *time.Time
		endDate     *time.Time
		description string
	}{
		{
			name:        "valid range",
			startDate:   &lastWeek,
			endDate:     &nextWeek,
			description: "normal date range",
		},
		{
			name:        "start only",
			startDate:   &yesterday,
			endDate:     nil,
			description: "start date only",
		},
		{
			name:        "end only",
			startDate:   nil,
			endDate:     &tomorrow,
			description: "end date only",
		},
		{
			name:        "no dates",
			startDate:   nil,
			endDate:     nil,
			description: "no date filters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := FilterParams{
				StartDate: tc.startDate,
				EndDate:   tc.endDate,
			}

			err := filters.Validate()
			assert.NoError(t, err, tc.description)
		})
	}
}

// TestTypeFilteringTests filtering by entity type
func TestTypeFiltering(t *testing.T) {
	testCases := []struct {
		name         string
		typeCategory *string
		description  string
	}{
		{
			name:         "person type",
			typeCategory: strPtr("person"),
			description:  "filter by person type",
		},
		{
			name:         "organization type",
			typeCategory: strPtr("organization"),
			description:  "filter by organization type",
		},
		{
			name:         "concept type",
			typeCategory: strPtr("concept"),
			description:  "filter by concept type",
		},
		{
			name:         "no type filter",
			typeCategory: nil,
			description:  "no type filtering",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := FilterParams{
				TypeCategory: tc.typeCategory,
			}

			err := filters.Validate()
			assert.NoError(t, err, tc.description)
		})
	}
}

// TestNameFiltering tests filtering by name
func TestNameFiltering(t *testing.T) {
	testCases := []struct {
		name        string
		nameFilter  *string
		description string
	}{
		{
			name:        "partial name",
			nameFilter:  strPtr("john"),
			description: "filter by partial name",
		},
		{
			name:        "full name",
			nameFilter:  strPtr("John Doe"),
			description: "filter by full name",
		},
		{
			name:        "no name filter",
			nameFilter:  nil,
			description: "no name filtering",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := FilterParams{
				Name: tc.nameFilter,
			}

			err := filters.Validate()
			assert.NoError(t, err, tc.description)
		})
	}
}

// TestCombinedFilters tests multiple filters together
func TestCombinedFilters(t *testing.T) {
	minConf := 0.7
	maxConf := 0.9
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	typeCategory := "person"
	name := "John"

	filters := FilterParams{
		MinConfidence: &minConf,
		MaxConfidence: &maxConf,
		StartDate:     &yesterday,
		EndDate:       &tomorrow,
		TypeCategory:  &typeCategory,
		Name:          &name,
	}

	err := filters.Validate()
	assert.NoError(t, err)

	// All values should be preserved after validation
	assert.Equal(t, minConf, *filters.MinConfidence)
	assert.Equal(t, maxConf, *filters.MaxConfidence)
	assert.Equal(t, typeCategory, *filters.TypeCategory)
	assert.Equal(t, name, *filters.Name)
}

// TestCombinedQueryParams tests pagination and filtering together
func TestCombinedQueryParams(t *testing.T) {
	minConf := 0.8
	typeCategory := "person"

	params := CombinedQueryParams{
		Pagination: PaginationParams{
			Limit:  50,
			Offset: 0,
		},
		Filters: FilterParams{
			MinConfidence: &minConf,
			TypeCategory:  &typeCategory,
		},
	}

	err := params.Validate()
	assert.NoError(t, err)

	assert.Equal(t, 50, params.Pagination.Limit)
	assert.Equal(t, 0, params.Pagination.Offset)
	assert.Equal(t, 0.8, *params.Filters.MinConfidence)
	assert.Equal(t, "person", *params.Filters.TypeCategory)
}

// Helper functions

func float64Ptr(f float64) *float64 {
	return &f
}

func strPtr(s string) *string {
	return &s
}
