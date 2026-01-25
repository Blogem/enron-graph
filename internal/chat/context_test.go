package chat

import (
	"testing"
	"time"
)

// TestConversationHistoryStorage tests storing last 5 queries
func TestConversationHistoryStorage(t *testing.T) {
	ctx := NewContext()

	// Add 3 queries
	ctx.AddQuery("Who is Jeff Skilling?", "Jeff Skilling is the CEO...")
	ctx.AddQuery("What did he do?", "He led the company...")
	ctx.AddQuery("When did he join?", "He joined in 1990...")

	history := ctx.GetHistory()
	if len(history) != 3 {
		t.Errorf("GetHistory() length = %d, want 3", len(history))
	}

	// Verify order (most recent last)
	if history[0].Query != "Who is Jeff Skilling?" {
		t.Errorf("First query = %s, want 'Who is Jeff Skilling?'", history[0].Query)
	}
	if history[2].Query != "When did he join?" {
		t.Errorf("Last query = %s, want 'When did he join?'", history[2].Query)
	}
}

// TestConversationHistoryLimit tests that only last 5 queries are kept
func TestConversationHistoryLimit(t *testing.T) {
	ctx := NewContext()

	// Add 7 queries
	for i := 1; i <= 7; i++ {
		query := "Query " + string(rune('0'+i))
		response := "Response " + string(rune('0'+i))
		ctx.AddQuery(query, response)
	}

	history := ctx.GetHistory()
	if len(history) != 5 {
		t.Errorf("GetHistory() length = %d, want 5 (limit)", len(history))
	}

	// Verify oldest queries were dropped (should start at query 3)
	if history[0].Query != "Query 3" {
		t.Errorf("First query after limit = %s, want 'Query 3'", history[0].Query)
	}
	if history[4].Query != "Query 7" {
		t.Errorf("Last query = %s, want 'Query 7'", history[4].Query)
	}
}

// TestEntityTracking tests tracking entities mentioned in conversation
func TestEntityTracking(t *testing.T) {
	ctx := NewContext()

	// Track entities
	ctx.TrackEntity("Jeff Skilling", "person", 1)
	ctx.TrackEntity("Kenneth Lay", "person", 2)
	ctx.TrackEntity("Enron", "organization", 3)

	// Get tracked entities
	entities := ctx.GetTrackedEntities()
	if len(entities) != 3 {
		t.Errorf("GetTrackedEntities() length = %d, want 3", len(entities))
	}

	// Verify entity details
	jeff, exists := entities["Jeff Skilling"]
	if !exists {
		t.Fatal("Entity 'Jeff Skilling' not found in tracked entities")
	}
	if jeff.Type != "person" {
		t.Errorf("Entity type = %s, want 'person'", jeff.Type)
	}
	if jeff.ID != 1 {
		t.Errorf("Entity ID = %d, want 1", jeff.ID)
	}
}

// TestPronounResolution tests resolving pronouns to tracked entities
func TestPronounResolution(t *testing.T) {
	ctx := NewContext()

	// Set up context: user asked about Jeff Skilling
	ctx.TrackEntity("Jeff Skilling", "person", 1)
	ctx.AddQuery("Who is Jeff Skilling?", "Jeff Skilling is the CEO...")

	// Resolve pronouns
	tests := []struct {
		name     string
		query    string
		wantName string
		wantID   int
	}{
		{
			name:     "he pronoun",
			query:    "What did he do?",
			wantName: "Jeff Skilling",
			wantID:   1,
		},
		{
			name:     "his pronoun",
			query:    "What was his role?",
			wantName: "Jeff Skilling",
			wantID:   1,
		},
		{
			name:     "him pronoun",
			query:    "Tell me about him",
			wantName: "Jeff Skilling",
			wantID:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, ok := ctx.ResolvePronoun(tt.query)
			if !ok {
				t.Fatal("ResolvePronoun() returned false, want true")
			}
			if entity.Name != tt.wantName {
				t.Errorf("ResolvePronoun() name = %s, want %s", entity.Name, tt.wantName)
			}
			if entity.ID != tt.wantID {
				t.Errorf("ResolvePronoun() ID = %d, want %d", entity.ID, tt.wantID)
			}
		})
	}
}

// TestMultipleEntityPronounResolution tests pronoun resolution with multiple entities
func TestMultipleEntityPronounResolution(t *testing.T) {
	ctx := NewContext()

	// Track multiple entities
	ctx.TrackEntity("Jeff Skilling", "person", 1)
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	ctx.TrackEntity("Kenneth Lay", "person", 2)
	ctx.AddQuery("Who is Jeff Skilling?", "Jeff Skilling is the CEO...")
	ctx.AddQuery("Who is Kenneth Lay?", "Kenneth Lay is the chairman...")

	// "He" should refer to the most recently mentioned entity (Kenneth Lay)
	entity, ok := ctx.ResolvePronoun("What did he do?")
	if !ok {
		t.Fatal("ResolvePronoun() returned false, want true")
	}
	if entity.Name != "Kenneth Lay" {
		t.Errorf("ResolvePronoun() name = %s, want 'Kenneth Lay' (most recent)", entity.Name)
	}
}

// TestOrganizationPronounResolution tests pronoun resolution for organizations
func TestOrganizationPronounResolution(t *testing.T) {
	ctx := NewContext()

	// Track organization
	ctx.TrackEntity("Enron", "organization", 3)
	ctx.AddQuery("What is Enron?", "Enron is an energy company...")

	// "It" should refer to organization
	entity, ok := ctx.ResolvePronoun("When was it founded?")
	if !ok {
		t.Fatal("ResolvePronoun() returned false, want true")
	}
	if entity.Name != "Enron" {
		t.Errorf("ResolvePronoun() name = %s, want 'Enron'", entity.Name)
	}
}

// TestContextInjection tests injecting context into LLM prompts
func TestContextInjection(t *testing.T) {
	ctx := NewContext()

	// Build context
	ctx.AddQuery("Who is Jeff Skilling?", "Jeff Skilling is the CEO...")
	ctx.TrackEntity("Jeff Skilling", "person", 1)

	// Get context for new query
	newQuery := "What did he do?"
	contextStr := ctx.BuildPromptContext(newQuery)

	// Verify context includes history
	if contextStr == "" {
		t.Error("BuildPromptContext() returned empty string")
	}

	// Context should mention previous query
	if len(ctx.GetHistory()) > 0 && contextStr == newQuery {
		t.Error("BuildPromptContext() should include conversation history")
	}
}

// TestContextClear tests clearing conversation context
func TestContextClear(t *testing.T) {
	ctx := NewContext()

	// Add some context
	ctx.AddQuery("Who is Jeff Skilling?", "Jeff Skilling is the CEO...")
	ctx.TrackEntity("Jeff Skilling", "person", 1)

	// Clear context
	ctx.Clear()

	// Verify everything is cleared
	if len(ctx.GetHistory()) != 0 {
		t.Errorf("After Clear(), history length = %d, want 0", len(ctx.GetHistory()))
	}
	if len(ctx.GetTrackedEntities()) != 0 {
		t.Errorf("After Clear(), tracked entities length = %d, want 0", len(ctx.GetTrackedEntities()))
	}
}

// TestTimestampTracking tests that queries are timestamped
func TestTimestampTracking(t *testing.T) {
	ctx := NewContext()

	before := time.Now()
	ctx.AddQuery("Test query", "Test response")
	after := time.Now()

	history := ctx.GetHistory()
	if len(history) != 1 {
		t.Fatalf("GetHistory() length = %d, want 1", len(history))
	}

	timestamp := history[0].Timestamp
	if timestamp.Before(before) || timestamp.After(after) {
		t.Errorf("Timestamp %v not between %v and %v", timestamp, before, after)
	}
}

// TestContextSerialization tests serializing/deserializing context
func TestContextSerialization(t *testing.T) {
	ctx := NewContext()

	// Add some context
	ctx.AddQuery("Who is Jeff Skilling?", "Jeff Skilling is the CEO...")
	ctx.TrackEntity("Jeff Skilling", "person", 1)

	// Serialize
	data, err := ctx.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Deserialize into new context
	newCtx := NewContext()
	if err := newCtx.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	// Verify data preserved
	if len(newCtx.GetHistory()) != len(ctx.GetHistory()) {
		t.Errorf("After deserialize, history length = %d, want %d",
			len(newCtx.GetHistory()), len(ctx.GetHistory()))
	}
	if len(newCtx.GetTrackedEntities()) != len(ctx.GetTrackedEntities()) {
		t.Errorf("After deserialize, entities length = %d, want %d",
			len(newCtx.GetTrackedEntities()), len(ctx.GetTrackedEntities()))
	}
}

// TestGetLastMentionedEntity tests getting the most recently mentioned entity
func TestGetLastMentionedEntity(t *testing.T) {
	ctx := NewContext()

	// No entities tracked yet
	_, ok := ctx.GetLastMentionedEntity()
	if ok {
		t.Error("GetLastMentionedEntity() returned true with no entities, want false")
	}

	// Track entities in sequence
	ctx.TrackEntity("Jeff Skilling", "person", 1)
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	ctx.TrackEntity("Kenneth Lay", "person", 2)

	// Should return most recent
	entity, ok := ctx.GetLastMentionedEntity()
	if !ok {
		t.Fatal("GetLastMentionedEntity() returned false, want true")
	}
	if entity.Name != "Kenneth Lay" {
		t.Errorf("GetLastMentionedEntity() name = %s, want 'Kenneth Lay'", entity.Name)
	}
}

// TestEntityTrackingWithUpdates tests updating tracked entity info
func TestEntityTrackingWithUpdates(t *testing.T) {
	ctx := NewContext()

	// Track entity initially
	ctx.TrackEntity("Jeff Skilling", "person", 1)

	// Track same entity again (should update, not duplicate)
	ctx.TrackEntity("Jeff Skilling", "person", 1)

	entities := ctx.GetTrackedEntities()
	if len(entities) != 1 {
		t.Errorf("After tracking same entity twice, length = %d, want 1", len(entities))
	}
}

// TestContextWithEmptyHistory tests handling empty conversation history
func TestContextWithEmptyHistory(t *testing.T) {
	ctx := NewContext()

	// Get history when empty
	history := ctx.GetHistory()
	if history == nil {
		t.Error("GetHistory() returned nil, want empty slice")
	}
	if len(history) != 0 {
		t.Errorf("GetHistory() length = %d, want 0", len(history))
	}

	// Build context with no history
	contextStr := ctx.BuildPromptContext("First query")
	if contextStr == "" {
		t.Error("BuildPromptContext() returned empty string for first query")
	}
}
