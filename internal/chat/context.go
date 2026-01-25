package chat

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// chatContext implements the Context interface for conversation management
type chatContext struct {
	history         []HistoryEntry
	trackedEntities map[string]TrackedEntity
	maxHistory      int
}

// NewContext creates a new conversation context
func NewContext() Context {
	return &chatContext{
		history:         make([]HistoryEntry, 0, 5),
		trackedEntities: make(map[string]TrackedEntity),
		maxHistory:      5,
	}
}

// AddQuery adds a query and response to the conversation history
func (c *chatContext) AddQuery(query, response string) {
	entry := HistoryEntry{
		Query:     query,
		Response:  response,
		Timestamp: time.Now(),
	}

	c.history = append(c.history, entry)

	// Keep only last maxHistory entries
	if len(c.history) > c.maxHistory {
		c.history = c.history[len(c.history)-c.maxHistory:]
	}
}

// GetHistory returns the conversation history
func (c *chatContext) GetHistory() []HistoryEntry {
	return c.history
}

// TrackEntity adds an entity to the tracked entities list
func (c *chatContext) TrackEntity(name, entityType string, id int) {
	c.trackedEntities[name] = TrackedEntity{
		Name:      name,
		Type:      entityType,
		ID:        id,
		Timestamp: time.Now(),
	}
}

// GetTrackedEntities returns all tracked entities
func (c *chatContext) GetTrackedEntities() map[string]TrackedEntity {
	return c.trackedEntities
}

// ResolvePronoun attempts to resolve a pronoun to a tracked entity
func (c *chatContext) ResolvePronoun(query string) (*TrackedEntity, bool) {
	queryLower := strings.ToLower(query)

	// Check for pronouns
	pronouns := []string{"he", "she", "it", "they", "him", "her", "them", "his", "hers", "their"}
	hasPronoun := false
	for _, pronoun := range pronouns {
		// Match pronoun as a whole word
		pattern := regexp.MustCompile(`\b` + pronoun + `\b`)
		if pattern.MatchString(queryLower) {
			hasPronoun = true
			break
		}
	}

	if !hasPronoun {
		return nil, false
	}

	// Return the most recently mentioned entity
	var mostRecent *TrackedEntity
	var mostRecentTime time.Time

	for _, entity := range c.trackedEntities {
		if mostRecent == nil || entity.Timestamp.After(mostRecentTime) {
			entityCopy := entity
			mostRecent = &entityCopy
			mostRecentTime = entity.Timestamp
		}
	}

	if mostRecent != nil {
		return mostRecent, true
	}

	return nil, false
}

// BuildPromptContext builds a context string for the LLM prompt
func (c *chatContext) BuildPromptContext(query string) string {
	var builder strings.Builder

	// Add conversation history
	if len(c.history) > 0 {
		builder.WriteString("Previous conversation:\n")
		for _, entry := range c.history {
			builder.WriteString(fmt.Sprintf("User: %s\n", entry.Query))
			builder.WriteString(fmt.Sprintf("Assistant: %s\n", entry.Response))
		}
		builder.WriteString("\n")
	}

	// Add tracked entities
	if len(c.trackedEntities) > 0 {
		builder.WriteString("Mentioned entities:\n")
		for _, entity := range c.trackedEntities {
			builder.WriteString(fmt.Sprintf("- %s (%s, ID: %d)\n", entity.Name, entity.Type, entity.ID))
		}
		builder.WriteString("\n")
	}

	// Resolve pronouns if present
	if resolved, ok := c.ResolvePronoun(query); ok {
		builder.WriteString(fmt.Sprintf("Note: Pronouns in the query likely refer to %s (%s)\n\n", resolved.Name, resolved.Type))
	}

	builder.WriteString(fmt.Sprintf("Current query: %s\n", query))

	return builder.String()
}

// Clear clears all conversation history and tracked entities
func (c *chatContext) Clear() {
	c.history = make([]HistoryEntry, 0, 5)
	c.trackedEntities = make(map[string]TrackedEntity)
}

// contextData is the internal structure for serialization
type contextData struct {
	History         []HistoryEntry           `json:"history"`
	TrackedEntities map[string]TrackedEntity `json:"tracked_entities"`
	MaxHistory      int                      `json:"max_history"`
}

// Serialize serializes the context to JSON bytes
func (c *chatContext) Serialize() ([]byte, error) {
	data := contextData{
		History:         c.history,
		TrackedEntities: c.trackedEntities,
		MaxHistory:      c.maxHistory,
	}
	return json.Marshal(data)
}

// Deserialize deserializes the context from JSON bytes
func (c *chatContext) Deserialize(data []byte) error {
	var contextData contextData
	if err := json.Unmarshal(data, &contextData); err != nil {
		return err
	}

	c.history = contextData.History
	c.trackedEntities = contextData.TrackedEntities
	c.maxHistory = contextData.MaxHistory

	return nil
}

// GetLastMentionedEntity returns the most recently mentioned entity
func (c *chatContext) GetLastMentionedEntity() (*TrackedEntity, bool) {
	if len(c.trackedEntities) == 0 {
		return nil, false
	}

	var mostRecent *TrackedEntity
	var mostRecentTime time.Time

	for _, entity := range c.trackedEntities {
		if mostRecent == nil || entity.Timestamp.After(mostRecentTime) {
			entityCopy := entity
			mostRecent = &entityCopy
			mostRecentTime = entity.Timestamp
		}
	}

	return mostRecent, mostRecent != nil
}
