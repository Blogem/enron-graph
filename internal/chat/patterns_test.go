package chat

import (
	"strings"
	"testing"
)

// TestEntityLookupPattern tests pattern matching for "Who is X?" queries
func TestEntityLookupPattern(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantType QueryType
		wantArgs map[string]string
	}{
		{
			name:     "simple who is query",
			query:    "Who is Jeff Skilling?",
			wantType: QueryTypeEntityLookup,
			wantArgs: map[string]string{"name": "Jeff Skilling"},
		},
		{
			name:     "what is query",
			query:    "What is Enron?",
			wantType: QueryTypeEntityLookup,
			wantArgs: map[string]string{"name": "Enron"},
		},
		{
			name:     "tell me about query",
			query:    "Tell me about Kenneth Lay",
			wantType: QueryTypeEntityLookup,
			wantArgs: map[string]string{"name": "Kenneth Lay"},
		},
		{
			name:     "lowercase query",
			query:    "who is john doe",
			wantType: QueryTypeEntityLookup,
			wantArgs: map[string]string{"name": "john doe"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewPatternMatcher()
			result, err := matcher.Match(tt.query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if result.Type != tt.wantType {
				t.Errorf("Match() type = %v, want %v", result.Type, tt.wantType)
			}
			for key, want := range tt.wantArgs {
				if got := result.Args[key]; got != want {
					t.Errorf("Match() args[%s] = %v, want %v", key, got, want)
				}
			}
		})
	}
}

// TestRelationshipPattern tests pattern matching for "Who did X email?" queries
func TestRelationshipPattern(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantType QueryType
		wantArgs map[string]string
	}{
		{
			name:     "who did X email",
			query:    "Who did Jeff Skilling email?",
			wantType: QueryTypeRelationship,
			wantArgs: map[string]string{
				"entity":    "Jeff Skilling",
				"relType":   "SENT",
				"direction": "outgoing",
			},
		},
		{
			name:     "who emailed X",
			query:    "Who emailed Jeff Skilling?",
			wantType: QueryTypeRelationship,
			wantArgs: map[string]string{
				"entity":    "Jeff Skilling",
				"relType":   "RECEIVED",
				"direction": "incoming",
			},
		},
		{
			name:     "who did X communicate with",
			query:    "Who did Kenneth Lay communicate with?",
			wantType: QueryTypeRelationship,
			wantArgs: map[string]string{
				"entity":  "Kenneth Lay",
				"relType": "COMMUNICATES_WITH",
			},
		},
		{
			name:     "what did X mention",
			query:    "What organizations did Jeff Skilling mention?",
			wantType: QueryTypeRelationship,
			wantArgs: map[string]string{
				"entity":     "Jeff Skilling",
				"relType":    "MENTIONS",
				"entityType": "organization",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewPatternMatcher()
			result, err := matcher.Match(tt.query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if result.Type != tt.wantType {
				t.Errorf("Match() type = %v, want %v", result.Type, tt.wantType)
			}
			for key, want := range tt.wantArgs {
				if got := result.Args[key]; got != want {
					t.Errorf("Match() args[%s] = %v, want %v", key, got, want)
				}
			}
		})
	}
}

// TestPathFindingPattern tests pattern matching for "How are X and Y connected?" queries
func TestPathFindingPattern(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantType QueryType
		wantArgs map[string]string
	}{
		{
			name:     "how are X and Y connected",
			query:    "How are Jeff Skilling and Kenneth Lay connected?",
			wantType: QueryTypePathFinding,
			wantArgs: map[string]string{
				"source": "Jeff Skilling",
				"target": "Kenneth Lay",
			},
		},
		{
			name:     "what is the relationship between X and Y",
			query:    "What is the relationship between Jeff Skilling and Kenneth Lay?",
			wantType: QueryTypePathFinding,
			wantArgs: map[string]string{
				"source": "Jeff Skilling",
				"target": "Kenneth Lay",
			},
		},
		{
			name:     "how does X know Y",
			query:    "How does Jeff Skilling know Kenneth Lay?",
			wantType: QueryTypePathFinding,
			wantArgs: map[string]string{
				"source": "Jeff Skilling",
				"target": "Kenneth Lay",
			},
		},
		{
			name:     "connection path query",
			query:    "Find the connection between Jeff Skilling and Kenneth Lay",
			wantType: QueryTypePathFinding,
			wantArgs: map[string]string{
				"source": "Jeff Skilling",
				"target": "Kenneth Lay",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewPatternMatcher()
			result, err := matcher.Match(tt.query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if result.Type != tt.wantType {
				t.Errorf("Match() type = %v, want %v", result.Type, tt.wantType)
			}
			for key, want := range tt.wantArgs {
				if got := result.Args[key]; got != want {
					t.Errorf("Match() args[%s] = %v, want %v", key, got, want)
				}
			}
		})
	}
}

// TestConceptSearchPattern tests pattern matching for "Emails about X" queries
func TestConceptSearchPattern(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantType QueryType
		wantArgs map[string]string
	}{
		{
			name:     "emails about X",
			query:    "Emails about energy trading",
			wantType: QueryTypeConceptSearch,
			wantArgs: map[string]string{"concept": "energy trading"},
		},
		{
			name:     "show me emails about X",
			query:    "Show me emails about California crisis",
			wantType: QueryTypeConceptSearch,
			wantArgs: map[string]string{"concept": "California crisis"},
		},
		{
			name:     "find emails related to X",
			query:    "Find emails related to financial reporting",
			wantType: QueryTypeConceptSearch,
			wantArgs: map[string]string{"concept": "financial reporting"},
		},
		{
			name:     "search for X",
			query:    "Search for discussions about quarterly earnings",
			wantType: QueryTypeConceptSearch,
			wantArgs: map[string]string{"concept": "quarterly earnings"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewPatternMatcher()
			result, err := matcher.Match(tt.query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if result.Type != tt.wantType {
				t.Errorf("Match() type = %v, want %v", result.Type, tt.wantType)
			}
			for key, want := range tt.wantArgs {
				if got := result.Args[key]; got != want {
					t.Errorf("Match() args[%s] = %v, want %v", key, got, want)
				}
			}
		})
	}
}

// TestAggregationPattern tests pattern matching for "How many emails did X send?" queries
func TestAggregationPattern(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantType QueryType
		wantArgs map[string]string
	}{
		{
			name:     "how many emails did X send",
			query:    "How many emails did Jeff Skilling send?",
			wantType: QueryTypeAggregation,
			wantArgs: map[string]string{
				"entity":      "Jeff Skilling",
				"aggregation": "count",
				"relType":     "SENT",
			},
		},
		{
			name:     "count emails from X",
			query:    "Count emails from Kenneth Lay",
			wantType: QueryTypeAggregation,
			wantArgs: map[string]string{
				"entity":      "Kenneth Lay",
				"aggregation": "count",
				"relType":     "SENT",
			},
		},
		{
			name:     "how many people did X email",
			query:    "How many people did Jeff Skilling email?",
			wantType: QueryTypeAggregation,
			wantArgs: map[string]string{
				"entity":      "Jeff Skilling",
				"aggregation": "count",
				"relType":     "COMMUNICATES_WITH",
			},
		},
		{
			name:     "how many organizations mentioned",
			query:    "How many organizations are mentioned in emails?",
			wantType: QueryTypeAggregation,
			wantArgs: map[string]string{
				"aggregation": "count",
				"entityType":  "organization",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewPatternMatcher()
			result, err := matcher.Match(tt.query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if result.Type != tt.wantType {
				t.Errorf("Match() type = %v, want %v", result.Type, tt.wantType)
			}
			for key, want := range tt.wantArgs {
				if got := result.Args[key]; got != want {
					t.Errorf("Match() args[%s] = %v, want %v", key, got, want)
				}
			}
		})
	}
}

// TestAmbiguityHandling tests handling of ambiguous queries
func TestAmbiguityHandling(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		wantAmbiguous bool
		minOptions    int
	}{
		{
			name:          "ambiguous person name",
			query:         "Who is John?",
			wantAmbiguous: true,
			minOptions:    2,
		},
		{
			name:          "unclear query intent",
			query:         "Tell me about energy",
			wantAmbiguous: true,
			minOptions:    2,
		},
		{
			name:          "clear query should not be ambiguous",
			query:         "Who is Jeff Skilling?",
			wantAmbiguous: false,
			minOptions:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewPatternMatcher()
			result, err := matcher.Match(tt.query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if result.Ambiguous != tt.wantAmbiguous {
				t.Errorf("Match() ambiguous = %v, want %v", result.Ambiguous, tt.wantAmbiguous)
			}
			if tt.wantAmbiguous && len(result.Options) < tt.minOptions {
				t.Errorf("Match() options count = %d, want at least %d", len(result.Options), tt.minOptions)
			}
		})
	}
}

// TestInvalidQueries tests handling of invalid or unsupported queries
func TestInvalidQueries(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantError bool
		wantType  QueryType
	}{
		{
			name:      "empty query",
			query:     "",
			wantError: true,
		},
		{
			name:      "whitespace only",
			query:     "   ",
			wantError: true,
		},
		{
			name:      "gibberish returns unknown type",
			query:     "asdf qwer zxcv",
			wantError: false,
			wantType:  QueryTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewPatternMatcher()
			result, err := matcher.Match(tt.query)
			if tt.wantError {
				if err == nil {
					t.Error("Match() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Match() unexpected error = %v", err)
				}
				if tt.wantType != "" && result.Type != tt.wantType {
					t.Errorf("Match() type = %v, want %v", result.Type, tt.wantType)
				}
			}
		})
	}
}

// TestPromptContainsContext tests that queries include context markers
func TestPromptContainsContext(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		contextKeyword string
	}{
		{
			name:           "pronoun indicates context needed",
			query:          "What did he do?",
			contextKeyword: "he",
		},
		{
			name:           "reference indicates context",
			query:          "Tell me more about that",
			contextKeyword: "that",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(strings.ToLower(tt.query), tt.contextKeyword) {
				t.Errorf("Query should contain context keyword %q", tt.contextKeyword)
			}
		})
	}
}
