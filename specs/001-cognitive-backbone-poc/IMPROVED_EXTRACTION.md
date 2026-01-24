# Improved Entity Type Discovery

## What Changed

The extraction prompt has been enhanced to enable dynamic discovery of entity types beyond the basic set (person, organization, concept). The system now:

1. **Learns from previous extractions** - Queries the database for already-discovered entity types
2. **Enriches the prompt** - Includes discovered types to guide the LLM
3. **Allows flexibility** - LLM can reuse existing types or create new ones as appropriate
4. **Builds a richer taxonomy** - Over time, creates a domain-specific entity type system

## How It Works

### Before (Static Types)
```
Extract:
1. Persons
2. Organizations  
3. Concepts
4. Other
```

### After (Dynamic Discovery)
```
Previously discovered entity types in this dataset:
- location
- project
- event
- product
- financial_instrument

IMPORTANT: Reuse these types when entities fit. You can also discover new types.

You have FULL FLEXIBILITY to identify entity types including:
- person, organization, concept (standard)
- project, location, event, document, product, regulation, technology, financial_instrument
- ANY OTHER TYPE that emerges from the content
```

## Testing

### Run Without Data Cleanup

To test the progressive type discovery, run the test multiple times without cleaning data:

```bash
# First run - discovers initial types
SKIP_CLEANUP=true ./tests/integration/test_user_story_2.sh

# Second run - reuses discovered types and may find more
SKIP_CLEANUP=true ./tests/integration/test_user_story_2.sh

# Third run - further refinement
SKIP_CLEANUP=true ./tests/integration/test_user_story_2.sh
```

### Query Discovered Types

Use the included script to see what types have been discovered:

```bash
./scripts/query_discovered_types.sh
```

This shows:
- Summary by type with counts and average confidence
- Sample entities for each type
- Timeline of when types were first discovered

### Example Output

After running the test 3 times, the system discovered these types:

| Type | Count | Avg Confidence | Examples |
|------|-------|----------------|----------|
| person | 5 | 1.00 | jeff.skilling, kenneth.lay, andrew.fastow |
| concept | 2 | 0.95 | Energy Trading Strategy, Financial Projections |
| organization | 2 | 0.90 | Enron Corp, Enron Energy Services |
| **location** | 1 | 0.90 | California market |
| **project** | 1 | 0.80 | International Expansion |
| **event** | 1 | 0.70 | Next board meeting |
| **product** | 1 | 0.80 | Energy Trading Division |

The bold types are newly discovered beyond the standard set!

## SQL Queries

### Get all entity types
```sql
SELECT DISTINCT type_category, COUNT(*) 
FROM discovered_entities 
GROUP BY type_category 
ORDER BY type_category;
```

### Get entities of a specific type
```sql
SELECT name, confidence_score, properties 
FROM discovered_entities 
WHERE type_category = 'project';
```

### See type discovery timeline
```sql
SELECT 
    type_category,
    MIN(created_at) as first_discovered,
    COUNT(*) as total_entities
FROM discovered_entities
GROUP BY type_category
ORDER BY first_discovered;
```

## Benefits

1. **Domain Adaptation** - System learns domain-specific entity types (e.g., "financial_instrument" for Enron)
2. **Consistency** - Once a type is discovered, it's reused across all future extractions
3. **Rich Taxonomy** - Builds up a comprehensive entity type system over time
4. **Flexible** - Not limited to pre-defined types - can discover whatever is relevant

## Implementation Details

### Code Changes

- [repository.go](../internal/graph/repository.go#L21) - Added `GetDistinctEntityTypes()` method
- [prompts.go](../internal/extractor/prompts.go#L9-L77) - Enhanced prompt with discovered types
- [extractor.go](../internal/extractor/extractor.go#L101-L164) - Fetches and uses discovered types
- [prompts.go](../internal/extractor/prompts.go#L140-L166) - Improved JSON parsing for array responses

### New Structures

```go
type ExtractedEntity struct {
    Type       string                 `json:"type"`
    Name       string                 `json:"name"`
    Properties map[string]interface{} `json:"properties,omitempty"`
    Confidence float64                `json:"confidence"`
}

type ExtractionResult struct {
    Entities []ExtractedEntity `json:"entities"`
}
```

All entities now use a unified structure with a dynamic `type` field instead of separate arrays for persons/organizations/concepts/other.
