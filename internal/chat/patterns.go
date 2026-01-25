package chat

import (
	"fmt"
	"regexp"
	"strings"
)

// patternMatcher implements the PatternMatcher interface
type patternMatcher struct {
	entityLookupPatterns  []*regexp.Regexp
	relationshipPatterns  []*regexp.Regexp
	pathFindingPatterns   []*regexp.Regexp
	conceptSearchPatterns []*regexp.Regexp
	aggregationPatterns   []*regexp.Regexp
}

// NewPatternMatcher creates a new pattern matcher
func NewPatternMatcher() PatternMatcher {
	return &patternMatcher{
		pathFindingPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^how\s+(?:are|is)\s+(.+?)\s+and\s+(.+?)\s+connected\??$`),
			regexp.MustCompile(`(?i)^what\s+is\s+the\s+relationship\s+between\s+(.+?)\s+and\s+(.+?)\??$`),
			regexp.MustCompile(`(?i)^how\s+does\s+(.+?)\s+know\s+(.+?)\??$`),
			regexp.MustCompile(`(?i)^find\s+(?:the\s+)?connection\s+between\s+(.+?)\s+and\s+(.+?)$`),
		},
		conceptSearchPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^emails?\s+about\s+(.+?)$`),
			regexp.MustCompile(`(?i)^show\s+me\s+emails?\s+about\s+(.+?)$`),
			regexp.MustCompile(`(?i)^find\s+emails?\s+related\s+to\s+(.+?)$`),
			regexp.MustCompile(`(?i)^search\s+for\s+(?:discussions?\s+about\s+)?(.+?)$`),
		},
		aggregationPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^how\s+many\s+emails?\s+did\s+(.+?)\s+send\??$`),
			regexp.MustCompile(`(?i)^count\s+emails?\s+from\s+(.+?)$`),
			regexp.MustCompile(`(?i)^how\s+many\s+emails?\s+did\s+(.+?)\s+receive\??$`),
			regexp.MustCompile(`(?i)^how\s+many\s+people\s+did\s+(.+?)\s+email\??$`),
			regexp.MustCompile(`(?i)^how\s+many\s+(organizations?|orgs?)\s+are\s+mentioned\s+in\s+emails?\??$`),
		},
		relationshipPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^who\s+did\s+(.+?)\s+email\??$`),
			regexp.MustCompile(`(?i)^who\s+emailed\s+(.+?)\??$`),
			regexp.MustCompile(`(?i)^who\s+did\s+(.+?)\s+communicate\s+with\??$`),
			regexp.MustCompile(`(?i)^what\s+(?:organizations?|orgs?)\s+did\s+(.+?)\s+mention\??$`),
		},
		entityLookupPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^who\s+is\s+(.+?)\??$`),
			regexp.MustCompile(`(?i)^what\s+is\s+(.+?)\??$`),
			regexp.MustCompile(`(?i)^tell\s+me\s+about\s+(.+?)$`),
			regexp.MustCompile(`(?i)^show\s+me\s+(.+?)$`),
		},
	}
}

// Match matches a query against known patterns and returns the match result
func (m *patternMatcher) Match(query string) (*MatchResult, error) {
	// Validate input
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	// Check for ambiguity first
	if isAmbiguous(query) {
		return &MatchResult{
			Type:      QueryTypeAmbiguous,
			Ambiguous: true,
			Options:   getDisambiguationOptions(query),
		}, nil
	}

	// Try path finding patterns first (before entity lookup to avoid conflicts)
	for _, pattern := range m.pathFindingPatterns {
		if matches := pattern.FindStringSubmatch(query); matches != nil {
			source := strings.TrimSpace(matches[1])
			target := strings.TrimSpace(matches[2])
			source = strings.TrimRight(source, "?.!")
			target = strings.TrimRight(target, "?.!")

			return &MatchResult{
				Type: QueryTypePathFinding,
				Args: map[string]string{
					"source": source,
					"target": target,
				},
			}, nil
		}
	}

	// Try concept search patterns (before entity lookup)
	for _, pattern := range m.conceptSearchPatterns {
		if matches := pattern.FindStringSubmatch(query); matches != nil {
			concept := strings.TrimSpace(matches[1])
			concept = strings.TrimRight(concept, "?.!")

			return &MatchResult{
				Type: QueryTypeConceptSearch,
				Args: map[string]string{"concept": concept},
			}, nil
		}
	}

	// Try aggregation patterns (before entity lookup)
	for _, pattern := range m.aggregationPatterns {
		if matches := pattern.FindStringSubmatch(query); matches != nil {
			result := &MatchResult{
				Type: QueryTypeAggregation,
				Args: map[string]string{"aggregation": "count"},
			}

			if len(matches) > 1 && matches[1] != "" {
				entity := strings.TrimSpace(matches[1])
				entity = strings.TrimRight(entity, "?.!")
				result.Args["entity"] = entity

				// Determine what to count
				if strings.Contains(strings.ToLower(query), "send") || strings.Contains(strings.ToLower(query), "from") {
					result.Args["relType"] = "SENT"
				} else if strings.Contains(strings.ToLower(query), "receive") {
					result.Args["relType"] = "RECEIVED"
				} else if strings.Contains(strings.ToLower(query), "people") {
					result.Args["relType"] = "COMMUNICATES_WITH"
				}
			}

			// Check for entity type aggregations
			if strings.Contains(strings.ToLower(query), "organization") {
				result.Args["entityType"] = "organization"
			}

			return result, nil
		}
	}

	// Try relationship patterns
	for _, pattern := range m.relationshipPatterns {
		if matches := pattern.FindStringSubmatch(query); matches != nil {
			entity := strings.TrimSpace(matches[1])
			entity = strings.TrimRight(entity, "?.!")

			result := &MatchResult{
				Type: QueryTypeRelationship,
				Args: map[string]string{"entity": entity},
			}

			// Determine relationship type and direction
			if strings.Contains(strings.ToLower(query), "did") && strings.Contains(strings.ToLower(query), "email") {
				result.Args["relType"] = "SENT"
				result.Args["direction"] = "outgoing"
			} else if strings.Contains(strings.ToLower(query), "emailed") {
				result.Args["relType"] = "RECEIVED"
				result.Args["direction"] = "incoming"
			} else if strings.Contains(strings.ToLower(query), "communicate") {
				result.Args["relType"] = "COMMUNICATES_WITH"
			} else if strings.Contains(strings.ToLower(query), "mention") {
				result.Args["relType"] = "MENTIONS"
				result.Args["entityType"] = "organization"
			}

			return result, nil
		}
	}

	// Try entity lookup patterns last
	for _, pattern := range m.entityLookupPatterns {
		if matches := pattern.FindStringSubmatch(query); matches != nil {
			name := strings.TrimSpace(matches[1])
			// Remove trailing punctuation
			name = strings.TrimRight(name, "?.!")
			return &MatchResult{
				Type: QueryTypeEntityLookup,
				Args: map[string]string{"name": name},
			}, nil
		}
	}

	// No pattern matched - return unknown
	return &MatchResult{
		Type: QueryTypeUnknown,
		Args: map[string]string{},
	}, nil
}

// isAmbiguous checks if a query is ambiguous
func isAmbiguous(query string) bool {
	queryLower := strings.ToLower(query)

	// Check for single-name queries that might be ambiguous
	if strings.HasPrefix(queryLower, "who is ") {
		name := strings.TrimPrefix(queryLower, "who is ")
		name = strings.TrimRight(name, "?.!")
		name = strings.TrimSpace(name)

		// Single common first names are ambiguous
		commonNames := []string{"john", "mike", "david", "robert", "james", "mary", "susan"}
		for _, common := range commonNames {
			if name == common {
				return true
			}
		}
	}

	// Check for vague "tell me about" queries
	if strings.HasPrefix(queryLower, "tell me about ") {
		subject := strings.TrimPrefix(queryLower, "tell me about ")
		subject = strings.TrimRight(subject, "?.!")
		subject = strings.TrimSpace(subject)

		// Generic terms are ambiguous
		genericTerms := []string{"energy", "finance", "trading", "power"}
		for _, term := range genericTerms {
			if subject == term {
				return true
			}
		}
	}

	return false
}

// getDisambiguationOptions returns possible interpretations for ambiguous queries
func getDisambiguationOptions(query string) []string {
	queryLower := strings.ToLower(query)

	if strings.HasPrefix(queryLower, "who is john") {
		return []string{
			"Did you mean the person 'John Smith'?",
			"Did you mean the person 'John Doe'?",
			"Did you mean the organization 'Johnson & Co'?",
		}
	}

	if strings.Contains(queryLower, "energy") {
		return []string{
			"Search for entities related to 'energy'",
			"Find emails about 'energy'",
			"Look up organization 'Energy Corp'",
		}
	}

	// Default disambiguation
	return []string{
		"Could you be more specific?",
		"Please provide more details",
	}
}
