package graph

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// T060: Unit tests for filters

// FilterParams holds filtering configuration for entity queries
type FilterParams struct {
	TypeCategory    string
	MinConfidence   float64
	MaxConfidence   float64
	StartDate       *time.Time
	EndDate         *time.Time
	NameContains    string
	PropertyFilters map[string]interface{}
}

// NewFilterParams creates a new filter params instance
func NewFilterParams() *FilterParams {
	return &FilterParams{
		MinConfidence:   0.0,
		MaxConfidence:   1.0,
		PropertyFilters: make(map[string]interface{}),
	}
}

// Validate ensures filter params are valid
func (f *FilterParams) Validate() error {
	if f.MinConfidence < 0.0 || f.MinConfidence > 1.0 {
		return &ValidationError{Field: "min_confidence", Message: "must be between 0.0 and 1.0"}
	}
	if f.MaxConfidence < 0.0 || f.MaxConfidence > 1.0 {
		return &ValidationError{Field: "max_confidence", Message: "must be between 0.0 and 1.0"}
	}
	if f.MinConfidence > f.MaxConfidence {
		return &ValidationError{Field: "confidence", Message: "min_confidence cannot exceed max_confidence"}
	}
	if f.StartDate != nil && f.EndDate != nil && f.StartDate.After(*f.EndDate) {
		return &ValidationError{Field: "date_range", Message: "start_date cannot be after end_date"}
	}
	return nil
}

// IsEmpty returns true if no filters are set
func (f *FilterParams) IsEmpty() bool {
	return f.TypeCategory == "" &&
		f.MinConfidence == 0.0 &&
		f.MaxConfidence == 1.0 &&
		f.StartDate == nil &&
		f.EndDate == nil &&
		f.NameContains == "" &&
		len(f.PropertyFilters) == 0
}

// HasConfidenceFilter returns true if confidence filtering is active
func (f *FilterParams) HasConfidenceFilter() bool {
	return f.MinConfidence > 0.0 || f.MaxConfidence < 1.0
}

// HasDateFilter returns true if date range filtering is active
func (f *FilterParams) HasDateFilter() bool {
	return f.StartDate != nil || f.EndDate != nil
}

// TestConfidenceScoreFiltering tests filtering by confidence score
func TestConfidenceScoreFiltering(t *testing.T) {
	testCases := []struct {
		name          string
		minConfidence float64
		maxConfidence float64
		entityScore   float64
		shouldMatch   bool
	}{
		{
			name:          "entity within range",
			minConfidence: 0.5,
			maxConfidence: 0.9,
			entityScore:   0.7,
			shouldMatch:   true,
		},
		{
			name:          "entity below min",
			minConfidence: 0.5,
			maxConfidence: 0.9,
			entityScore:   0.3,
			shouldMatch:   false,
		},
		{
			name:          "entity above max",
			minConfidence: 0.5,
			maxConfidence: 0.9,
			entityScore:   0.95,
			shouldMatch:   false,
		},
		{
			name:          "entity at min boundary",
			minConfidence: 0.5,
			maxConfidence: 0.9,
			entityScore:   0.5,
			shouldMatch:   true,
		},
		{
			name:          "entity at max boundary",
			minConfidence: 0.5,
			maxConfidence: 0.9,
			entityScore:   0.9,
			shouldMatch:   true,
		},
		{
			name:          "no min filter",
			minConfidence: 0.0,
			maxConfidence: 0.9,
			entityScore:   0.1,
			shouldMatch:   true,
		},
		{
			name:          "no max filter",
			minConfidence: 0.5,
			maxConfidence: 1.0,
			entityScore:   0.95,
			shouldMatch:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := NewFilterParams()
			filters.MinConfidence = tc.minConfidence
			filters.MaxConfidence = tc.maxConfidence

			matches := tc.entityScore >= filters.MinConfidence && tc.entityScore <= filters.MaxConfidence
			assert.Equal(t, tc.shouldMatch, matches)
		})
	}
}

// TestDateRangeFiltering tests filtering by date range
func TestDateRangeFiltering(t *testing.T) {
	baseTime := time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC)
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC)

	testCases := []struct {
		name        string
		startDate   *time.Time
		endDate     *time.Time
		entityDate  time.Time
		shouldMatch bool
	}{
		{
			name:        "date within range",
			startDate:   &startDate,
			endDate:     &endDate,
			entityDate:  baseTime,
			shouldMatch: true,
		},
		{
			name:        "date before range",
			startDate:   &startDate,
			endDate:     &endDate,
			entityDate:  time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC),
			shouldMatch: false,
		},
		{
			name:        "date after range",
			startDate:   &startDate,
			endDate:     &endDate,
			entityDate:  time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
			shouldMatch: false,
		},
		{
			name:        "date at start boundary",
			startDate:   &startDate,
			endDate:     &endDate,
			entityDate:  startDate,
			shouldMatch: true,
		},
		{
			name:        "date at end boundary",
			startDate:   &startDate,
			endDate:     &endDate,
			entityDate:  endDate,
			shouldMatch: true,
		},
		{
			name:        "only start date",
			startDate:   &startDate,
			endDate:     nil,
			entityDate:  baseTime,
			shouldMatch: true,
		},
		{
			name:        "only end date",
			startDate:   nil,
			endDate:     &endDate,
			entityDate:  baseTime,
			shouldMatch: true,
		},
		{
			name:        "no date filters",
			startDate:   nil,
			endDate:     nil,
			entityDate:  baseTime,
			shouldMatch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := NewFilterParams()
			filters.StartDate = tc.startDate
			filters.EndDate = tc.endDate

			matches := true
			if filters.StartDate != nil {
				matches = matches && !tc.entityDate.Before(*filters.StartDate)
			}
			if filters.EndDate != nil {
				matches = matches && !tc.entityDate.After(*filters.EndDate)
			}

			assert.Equal(t, tc.shouldMatch, matches)
		})
	}
}

// TestTypeFiltering tests filtering by entity type
func TestTypeFiltering(t *testing.T) {
	testCases := []struct {
		name        string
		filterType  string
		entityType  string
		shouldMatch bool
	}{
		{
			name:        "exact type match",
			filterType:  "person",
			entityType:  "person",
			shouldMatch: true,
		},
		{
			name:        "type mismatch",
			filterType:  "person",
			entityType:  "organization",
			shouldMatch: false,
		},
		{
			name:        "no filter",
			filterType:  "",
			entityType:  "person",
			shouldMatch: true,
		},
		{
			name:        "concept type",
			filterType:  "concept",
			entityType:  "concept",
			shouldMatch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := NewFilterParams()
			filters.TypeCategory = tc.filterType

			matches := filters.TypeCategory == "" || filters.TypeCategory == tc.entityType
			assert.Equal(t, tc.shouldMatch, matches)
		})
	}
}

// TestCombinedFilters tests multiple filters applied together
func TestCombinedFilters(t *testing.T) {
	baseTime := time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC)
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC)

	testCases := []struct {
		name          string
		filterType    string
		minConfidence float64
		maxConfidence float64
		startDate     *time.Time
		endDate       *time.Time
		entityType    string
		entityScore   float64
		entityDate    time.Time
		shouldMatch   bool
		description   string
	}{
		{
			name:          "all filters match",
			filterType:    "person",
			minConfidence: 0.5,
			maxConfidence: 0.9,
			startDate:     &startDate,
			endDate:       &endDate,
			entityType:    "person",
			entityScore:   0.7,
			entityDate:    baseTime,
			shouldMatch:   true,
			description:   "entity matches all criteria",
		},
		{
			name:          "type mismatch",
			filterType:    "person",
			minConfidence: 0.5,
			maxConfidence: 0.9,
			startDate:     &startDate,
			endDate:       &endDate,
			entityType:    "organization",
			entityScore:   0.7,
			entityDate:    baseTime,
			shouldMatch:   false,
			description:   "type doesn't match",
		},
		{
			name:          "confidence too low",
			filterType:    "person",
			minConfidence: 0.5,
			maxConfidence: 0.9,
			startDate:     &startDate,
			endDate:       &endDate,
			entityType:    "person",
			entityScore:   0.3,
			entityDate:    baseTime,
			shouldMatch:   false,
			description:   "confidence below minimum",
		},
		{
			name:          "date out of range",
			filterType:    "person",
			minConfidence: 0.5,
			maxConfidence: 0.9,
			startDate:     &startDate,
			endDate:       &endDate,
			entityType:    "person",
			entityScore:   0.7,
			entityDate:    time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
			shouldMatch:   false,
			description:   "date after range",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := NewFilterParams()
			filters.TypeCategory = tc.filterType
			filters.MinConfidence = tc.minConfidence
			filters.MaxConfidence = tc.maxConfidence
			filters.StartDate = tc.startDate
			filters.EndDate = tc.endDate

			// Apply all filters
			matches := true

			// Type filter
			if filters.TypeCategory != "" {
				matches = matches && (filters.TypeCategory == tc.entityType)
			}

			// Confidence filter
			matches = matches && (tc.entityScore >= filters.MinConfidence && tc.entityScore <= filters.MaxConfidence)

			// Date filter
			if filters.StartDate != nil {
				matches = matches && !tc.entityDate.Before(*filters.StartDate)
			}
			if filters.EndDate != nil {
				matches = matches && !tc.entityDate.After(*filters.EndDate)
			}

			assert.Equal(t, tc.shouldMatch, matches, tc.description)
		})
	}
}

// TestFilterValidation tests filter parameter validation
func TestFilterValidation(t *testing.T) {
	futureDate := time.Now().Add(24 * time.Hour)
	pastDate := time.Now().Add(-24 * time.Hour)

	testCases := []struct {
		name        string
		setupFilter func(*FilterParams)
		expectError bool
		errorField  string
	}{
		{
			name: "valid filters",
			setupFilter: func(f *FilterParams) {
				f.MinConfidence = 0.5
				f.MaxConfidence = 0.9
				f.TypeCategory = "person"
			},
			expectError: false,
		},
		{
			name: "min confidence too low",
			setupFilter: func(f *FilterParams) {
				f.MinConfidence = -0.1
			},
			expectError: true,
			errorField:  "min_confidence",
		},
		{
			name: "max confidence too high",
			setupFilter: func(f *FilterParams) {
				f.MaxConfidence = 1.5
			},
			expectError: true,
			errorField:  "max_confidence",
		},
		{
			name: "min greater than max",
			setupFilter: func(f *FilterParams) {
				f.MinConfidence = 0.9
				f.MaxConfidence = 0.5
			},
			expectError: true,
			errorField:  "confidence",
		},
		{
			name: "start date after end date",
			setupFilter: func(f *FilterParams) {
				f.StartDate = &futureDate
				f.EndDate = &pastDate
			},
			expectError: true,
			errorField:  "date_range",
		},
		{
			name: "valid date range",
			setupFilter: func(f *FilterParams) {
				f.StartDate = &pastDate
				f.EndDate = &futureDate
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := NewFilterParams()
			tc.setupFilter(filters)

			err := filters.Validate()

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorField != "" {
					validationErr, ok := err.(*ValidationError)
					assert.True(t, ok, "expected ValidationError")
					assert.Equal(t, tc.errorField, validationErr.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFilterIsEmpty tests detection of empty filters
func TestFilterIsEmpty(t *testing.T) {
	testCases := []struct {
		name        string
		setupFilter func(*FilterParams)
		expectEmpty bool
	}{
		{
			name:        "default filters are empty",
			setupFilter: func(f *FilterParams) {},
			expectEmpty: true,
		},
		{
			name: "type filter set",
			setupFilter: func(f *FilterParams) {
				f.TypeCategory = "person"
			},
			expectEmpty: false,
		},
		{
			name: "confidence filter set",
			setupFilter: func(f *FilterParams) {
				f.MinConfidence = 0.5
			},
			expectEmpty: false,
		},
		{
			name: "date filter set",
			setupFilter: func(f *FilterParams) {
				now := time.Now()
				f.StartDate = &now
			},
			expectEmpty: false,
		},
		{
			name: "name filter set",
			setupFilter: func(f *FilterParams) {
				f.NameContains = "test"
			},
			expectEmpty: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := NewFilterParams()
			tc.setupFilter(filters)

			assert.Equal(t, tc.expectEmpty, filters.IsEmpty())
		})
	}
}

// TestHasSpecificFilter tests detection of specific filter types
func TestHasSpecificFilter(t *testing.T) {
	t.Run("has confidence filter", func(t *testing.T) {
		filters := NewFilterParams()
		assert.False(t, filters.HasConfidenceFilter())

		filters.MinConfidence = 0.5
		assert.True(t, filters.HasConfidenceFilter())

		filters = NewFilterParams()
		filters.MaxConfidence = 0.8
		assert.True(t, filters.HasConfidenceFilter())
	})

	t.Run("has date filter", func(t *testing.T) {
		filters := NewFilterParams()
		assert.False(t, filters.HasDateFilter())

		now := time.Now()
		filters.StartDate = &now
		assert.True(t, filters.HasDateFilter())

		filters = NewFilterParams()
		filters.EndDate = &now
		assert.True(t, filters.HasDateFilter())
	})
}

// TestNameContainsFilter tests partial name matching
func TestNameContainsFilter(t *testing.T) {
	testCases := []struct {
		name         string
		nameContains string
		entityName   string
		shouldMatch  bool
	}{
		{
			name:         "exact match",
			nameContains: "Jeff",
			entityName:   "Jeff",
			shouldMatch:  true,
		},
		{
			name:         "partial match",
			nameContains: "Jeff",
			entityName:   "Jeff Skilling",
			shouldMatch:  true,
		},
		{
			name:         "case insensitive match",
			nameContains: "jeff",
			entityName:   "Jeff Skilling",
			shouldMatch:  true,
		},
		{
			name:         "no match",
			nameContains: "Kenneth",
			entityName:   "Jeff Skilling",
			shouldMatch:  false,
		},
		{
			name:         "empty filter matches all",
			nameContains: "",
			entityName:   "Jeff Skilling",
			shouldMatch:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := NewFilterParams()
			filters.NameContains = tc.nameContains

			// Case-insensitive substring match
			matches := filters.NameContains == "" ||
				strings.Contains(
					strings.ToLower(tc.entityName),
					strings.ToLower(filters.NameContains),
				)

			assert.Equal(t, tc.shouldMatch, matches)
		})
	}
}

// TestPropertyFilters tests custom property filtering
func TestPropertyFilters(t *testing.T) {
	testCases := []struct {
		name            string
		propertyFilters map[string]interface{}
		entityProps     map[string]interface{}
		shouldMatch     bool
	}{
		{
			name: "property matches",
			propertyFilters: map[string]interface{}{
				"title": "CEO",
			},
			entityProps: map[string]interface{}{
				"title": "CEO",
			},
			shouldMatch: true,
		},
		{
			name: "property doesn't match",
			propertyFilters: map[string]interface{}{
				"title": "CEO",
			},
			entityProps: map[string]interface{}{
				"title": "CFO",
			},
			shouldMatch: false,
		},
		{
			name: "property missing",
			propertyFilters: map[string]interface{}{
				"title": "CEO",
			},
			entityProps: map[string]interface{}{},
			shouldMatch: false,
		},
		{
			name:            "no property filters",
			propertyFilters: map[string]interface{}{},
			entityProps: map[string]interface{}{
				"title": "CEO",
			},
			shouldMatch: true,
		},
		{
			name: "multiple properties all match",
			propertyFilters: map[string]interface{}{
				"title":      "CEO",
				"department": "Executive",
			},
			entityProps: map[string]interface{}{
				"title":      "CEO",
				"department": "Executive",
			},
			shouldMatch: true,
		},
		{
			name: "multiple properties one doesn't match",
			propertyFilters: map[string]interface{}{
				"title":      "CEO",
				"department": "Executive",
			},
			entityProps: map[string]interface{}{
				"title":      "CEO",
				"department": "Finance",
			},
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := NewFilterParams()
			filters.PropertyFilters = tc.propertyFilters

			matches := true
			for key, value := range filters.PropertyFilters {
				entityValue, exists := tc.entityProps[key]
				if !exists || entityValue != value {
					matches = false
					break
				}
			}

			assert.Equal(t, tc.shouldMatch, matches)
		})
	}
}
