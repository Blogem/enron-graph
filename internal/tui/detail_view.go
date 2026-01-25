package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/graph"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DetailViewModel represents the detail view state
type DetailViewModel struct {
	entity           *EntityDetail
	relationships    []Relationship
	cursor           int
	selectedID       int
	selectedEntityID int // Entity to navigate to (from relationships)
}

// EntityDetail represents detailed entity information
type EntityDetail struct {
	ID         int
	Type       string
	Name       string
	Properties map[string]interface{}
	Confidence float64
}

// Relationship represents a relationship to display
type Relationship struct {
	ID     int
	Type   string
	ToID   int
	ToName string
	ToType string
}

// NewDetailViewModel creates a new detail view model
func NewDetailViewModel() *DetailViewModel {
	return &DetailViewModel{
		entity:           nil,
		relationships:    []Relationship{},
		cursor:           0,
		selectedID:       0,
		selectedEntityID: 0,
	}
}

// Update handles messages for detail view
func (m *DetailViewModel) Update(msg tea.Msg) (*DetailViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.relationships)-1 {
				m.cursor++
			}
		case "enter":
			// Navigate to related entity details
			if m.cursor >= 0 && m.cursor < len(m.relationships) {
				m.selectedEntityID = m.relationships[m.cursor].ToID
			}
			// Navigation keys (v, b, esc) are handled by parent app.go
		}
	}

	return m, nil
}

// View renders the detail view
func (m *DetailViewModel) View(width, height int) string {
	if m.entity == nil {
		return "No entity selected.\n"
	}

	var b strings.Builder

	// Render entity header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	b.WriteString(headerStyle.Render(fmt.Sprintf("Entity: %s", m.entity.Name)))
	b.WriteString("\n\n")

	// Render properties
	b.WriteString("Properties:\n")
	b.WriteString(strings.Repeat("-", 50) + "\n")
	b.WriteString(fmt.Sprintf("  ID:         %d\n", m.entity.ID))
	b.WriteString(fmt.Sprintf("  Type:       %s\n", m.entity.Type))
	b.WriteString(fmt.Sprintf("  Name:       %s\n", m.entity.Name))
	b.WriteString(fmt.Sprintf("  Confidence: %.0f%%\n", m.entity.Confidence*100))

	if len(m.entity.Properties) > 0 {
		b.WriteString("\n  Custom Properties:\n")
		for key, value := range m.entity.Properties {
			b.WriteString(fmt.Sprintf("    %s: %v\n", key, value))
		}
	}

	// Render relationships
	b.WriteString("\n\nRelationships:\n")
	b.WriteString(strings.Repeat("-", 50) + "\n")

	if len(m.relationships) == 0 {
		b.WriteString("  No relationships found.\n")
	} else {
		for i, rel := range m.relationships {
			style := lipgloss.NewStyle()
			if i == m.cursor {
				style = style.Background(lipgloss.Color("62"))
			}

			relStr := fmt.Sprintf("  %s → [%s: %s]",
				rel.Type,
				rel.ToType,
				rel.ToName,
			)
			b.WriteString(style.Render(relStr))
			b.WriteString("\n")
		}
	}

	// Render actions
	b.WriteString("\n\nActions:\n")
	b.WriteString("  ↑↓: Navigate Relationships\n")
	b.WriteString("  Enter: View Selected Entity Details\n")
	b.WriteString("  V: Visualize in Graph View\n")
	b.WriteString("  B/Esc: Back to Entity List\n")

	return b.String()
}

// LoadEntity loads entity details (placeholder - would fetch from repo)
func (m *DetailViewModel) LoadEntity(entityID int) {
	// This would be called by the main app to load data from repository
	// For now, it's a placeholder
	m.entity = &EntityDetail{
		ID:         entityID,
		Type:       "person",
		Name:       "Loading...",
		Properties: make(map[string]interface{}),
		Confidence: 0.0,
	}
	m.relationships = []Relationship{}
}

// LoadEntityData loads entity details with actual data
func (m *DetailViewModel) LoadEntityData(id int, entityType, name string, properties map[string]interface{}, confidence float64) {
	m.entity = &EntityDetail{
		ID:         id,
		Type:       entityType,
		Name:       name,
		Properties: properties,
		Confidence: confidence,
	}
	m.relationships = []Relationship{}
}

// LoadRelationships loads relationships for the current entity
func (m *DetailViewModel) LoadRelationships(ctx context.Context, repo graph.Repository, rels []*ent.Relationship) {
	relationships := make([]Relationship, 0, len(rels))

	for _, rel := range rels {
		// Determine target entity ID (could be to_id or from_id depending on direction)
		var targetID int
		var targetType string

		if rel.FromID == m.entity.ID {
			targetID = rel.ToID
			targetType = rel.ToType
		} else {
			targetID = rel.FromID
			targetType = rel.FromType
		}

		// Skip relationships to emails
		if targetType == "email" {
			continue
		}

		// Fetch target entity name
		targetEntity, err := repo.FindEntityByID(ctx, targetID)
		targetName := "Unknown"
		if err == nil {
			targetName = targetEntity.Name
		}

		relationships = append(relationships, Relationship{
			ID:     rel.ID,
			Type:   rel.Type,
			ToID:   targetID,
			ToName: targetName,
			ToType: targetType,
		})
	}

	m.relationships = relationships
}
