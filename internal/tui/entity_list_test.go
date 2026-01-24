package tui

import (
	"testing"
)

// TestTableRendering tests that entity table renders correctly with columns
func TestTableRendering(t *testing.T) {
	entities := []Entity{
		{ID: 1, Type: "person", Name: "Jeff Skilling", Confidence: 0.95},
		{ID: 2, Type: "organization", Name: "Enron Corp", Confidence: 0.88},
		{ID: 3, Type: "concept", Name: "energy trading", Confidence: 0.72},
	}

	table := renderEntityTable(entities)

	// Check that headers are present
	if !contains(table, "ID") || !contains(table, "Type") || !contains(table, "Name") || !contains(table, "Confidence") {
		t.Error("Table missing required column headers")
	}

	// Check that data rows are present
	if !contains(table, "Jeff Skilling") || !contains(table, "Enron Corp") || !contains(table, "energy trading") {
		t.Error("Table missing entity data")
	}

	// Check confidence scores are formatted
	if !contains(table, "95%") || !contains(table, "88%") || !contains(table, "72%") {
		t.Error("Table confidence scores not formatted correctly")
	}
}

// TestPaginationCalculation tests pagination calculations
func TestPaginationCalculation(t *testing.T) {
	tests := []struct {
		name        string
		totalItems  int
		pageSize    int
		currentPage int
		wantStart   int
		wantEnd     int
	}{
		{
			name:        "first page",
			totalItems:  100,
			pageSize:    10,
			currentPage: 0,
			wantStart:   0,
			wantEnd:     10,
		},
		{
			name:        "middle page",
			totalItems:  100,
			pageSize:    10,
			currentPage: 5,
			wantStart:   50,
			wantEnd:     60,
		},
		{
			name:        "last page with partial results",
			totalItems:  95,
			pageSize:    10,
			currentPage: 9,
			wantStart:   90,
			wantEnd:     95,
		},
		{
			name:        "page beyond total",
			totalItems:  50,
			pageSize:    10,
			currentPage: 10,
			wantStart:   50,
			wantEnd:     50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := calculatePagination(tt.totalItems, tt.pageSize, tt.currentPage)
			if start != tt.wantStart || end != tt.wantEnd {
				t.Errorf("calculatePagination(%d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.totalItems, tt.pageSize, tt.currentPage, start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

// TestFilterByType tests filtering entities by type
func TestFilterByType(t *testing.T) {
	entities := []Entity{
		{ID: 1, Type: "person", Name: "Jeff Skilling"},
		{ID: 2, Type: "organization", Name: "Enron Corp"},
		{ID: 3, Type: "person", Name: "Kenneth Lay"},
		{ID: 4, Type: "concept", Name: "energy trading"},
		{ID: 5, Type: "person", Name: "Andrew Fastow"},
	}

	tests := []struct {
		name       string
		filterType string
		wantCount  int
	}{
		{
			name:       "filter person",
			filterType: "person",
			wantCount:  3,
		},
		{
			name:       "filter organization",
			filterType: "organization",
			wantCount:  1,
		},
		{
			name:       "filter concept",
			filterType: "concept",
			wantCount:  1,
		},
		{
			name:       "no filter (all)",
			filterType: "",
			wantCount:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterEntitiesByType(entities, tt.filterType)
			if len(filtered) != tt.wantCount {
				t.Errorf("filterEntitiesByType() returned %d entities, want %d",
					len(filtered), tt.wantCount)
			}
		})
	}
}

// TestSearchByName tests searching entities by name
func TestSearchByName(t *testing.T) {
	entities := []Entity{
		{ID: 1, Type: "person", Name: "Jeff Skilling"},
		{ID: 2, Type: "organization", Name: "Enron Corp"},
		{ID: 3, Type: "person", Name: "Kenneth Lay"},
		{ID: 4, Type: "concept", Name: "energy trading"},
		{ID: 5, Type: "person", Name: "Jeffrey McMahon"},
	}

	tests := []struct {
		name        string
		searchQuery string
		wantCount   int
		wantIDs     []int
	}{
		{
			name:        "search jeff",
			searchQuery: "jeff",
			wantCount:   2,
			wantIDs:     []int{1, 5},
		},
		{
			name:        "search enron",
			searchQuery: "enron",
			wantCount:   1,
			wantIDs:     []int{2},
		},
		{
			name:        "search energy",
			searchQuery: "energy",
			wantCount:   1,
			wantIDs:     []int{4},
		},
		{
			name:        "case insensitive search",
			searchQuery: "KENNETH",
			wantCount:   1,
			wantIDs:     []int{3},
		},
		{
			name:        "no matches",
			searchQuery: "xyz",
			wantCount:   0,
			wantIDs:     []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := searchEntitiesByName(entities, tt.searchQuery)
			if len(results) != tt.wantCount {
				t.Errorf("searchEntitiesByName() returned %d entities, want %d",
					len(results), tt.wantCount)
			}
			for i, result := range results {
				if i < len(tt.wantIDs) && result.ID != tt.wantIDs[i] {
					t.Errorf("searchEntitiesByName() result[%d].ID = %d, want %d",
						i, result.ID, tt.wantIDs[i])
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
			findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
