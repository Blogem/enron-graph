package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DetailViewModel represents the detail view state
type DetailViewModel struct {
	entity        *EntityDetail
	relationships []Relationship
	cursor        int
	selectedID    int
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
		entity:        nil,
		relationships: []Relationship{},
		cursor:        0,
		selectedID:    0,
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
			// Select relationship
			if m.cursor >= 0 && m.cursor < len(m.relationships) {
				m.selectedID = m.relationships[m.cursor].ToID
			}
		case "v":
			// Trigger visualization (handled by parent)
		case "b", "esc":
			// Back (handled by parent)
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
	b.WriteString("  V: Visualize in Graph View\n")
	b.WriteString("  R: Browse Related Entities\n")
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
