package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EntityListModel represents the entity list view state
type EntityListModel struct {
	entities    []Entity
	cursor      int
	page        int
	pageSize    int
	filterType  string
	searchQuery string
	selectedID  int
	isSearching bool
	isFiltering bool
}

// Entity represents a discovered entity for display
type Entity struct {
	ID         int
	Type       string
	Name       string
	Confidence float64
}

// NewEntityListModel creates a new entity list model
func NewEntityListModel() *EntityListModel {
	return &EntityListModel{
		entities:    []Entity{},
		cursor:      0,
		page:        0,
		pageSize:    20,
		filterType:  "",
		searchQuery: "",
		selectedID:  0,
		isSearching: false,
		isFiltering: false,
	}
}

// Update handles messages for entity list view
func (m *EntityListModel) Update(msg tea.Msg) (*EntityListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.isSearching {
			return m.handleSearchInput(msg)
		}
		if m.isFiltering {
			return m.handleFilterInput(msg)
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			visible := m.getVisibleEntities()
			if m.cursor < len(visible)-1 {
				m.cursor++
			}
		case "pgup":
			m.cursor -= m.pageSize
			if m.cursor < 0 {
				m.cursor = 0
			}
		case "pgdown":
			visible := m.getVisibleEntities()
			m.cursor += m.pageSize
			if m.cursor >= len(visible) {
				m.cursor = len(visible) - 1
			}
		case "f":
			m.isFiltering = true
			m.filterType = ""
		case "/":
			m.isSearching = true
			m.searchQuery = ""
		case "enter":
			visible := m.getVisibleEntities()
			if m.cursor >= 0 && m.cursor < len(visible) {
				m.selectedID = visible[m.cursor].ID
			}
		case "esc":
			m.filterType = ""
			m.searchQuery = ""
			m.cursor = 0
		}
	}

	return m, nil
}

// handleSearchInput handles keyboard input during search
func (m *EntityListModel) handleSearchInput(msg tea.KeyMsg) (*EntityListModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.isSearching = false
		m.cursor = 0
	case "esc":
		m.isSearching = false
		m.searchQuery = ""
		m.cursor = 0
	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
		}
	}
	return m, nil
}

// handleFilterInput handles keyboard input during filtering
func (m *EntityListModel) handleFilterInput(msg tea.KeyMsg) (*EntityListModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.isFiltering = false
		m.cursor = 0
	case "esc":
		m.isFiltering = false
		m.filterType = ""
		m.cursor = 0
	case "backspace":
		if len(m.filterType) > 0 {
			m.filterType = m.filterType[:len(m.filterType)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.filterType += msg.String()
		}
	}
	return m, nil
}

// View renders the entity list view
func (m *EntityListModel) View(width, height int) string {
	if len(m.entities) == 0 {
		return "No entities found. Load data first.\n"
	}

	var b strings.Builder

	// Show filter/search status
	if m.filterType != "" {
		b.WriteString(fmt.Sprintf("Filter: %s\n", m.filterType))
	}
	if m.searchQuery != "" {
		b.WriteString(fmt.Sprintf("Search: %s\n", m.searchQuery))
	}
	if m.isSearching {
		b.WriteString(fmt.Sprintf("Search: %s_\n", m.searchQuery))
	}
	if m.isFiltering {
		b.WriteString(fmt.Sprintf("Filter: %s_\n", m.filterType))
	}

	// Render table
	visible := m.getVisibleEntities()
	table := m.renderTable(visible, width, height-5)
	b.WriteString(table)

	// Show pagination info
	b.WriteString(fmt.Sprintf("\nShowing %d of %d entities\n", len(visible), len(m.entities)))

	return b.String()
}

// renderTable renders the entity table
func (m *EntityListModel) renderTable(entities []Entity, width, height int) string {
	if len(entities) == 0 {
		return "No matching entities found.\n"
	}

	// Calculate pagination
	start, end := calculatePagination(len(entities), m.pageSize, m.page)
	pageEntities := entities[start:end]

	// Render header
	headerStyle := lipgloss.NewStyle().Bold(true)
	header := headerStyle.Render(
		fmt.Sprintf("%-6s %-15s %-40s %-10s", "ID", "Type", "Name", "Confidence"),
	)

	var rows []string
	rows = append(rows, header)
	rows = append(rows, strings.Repeat("-", width-2))

	// Render rows
	for i, entity := range pageEntities {
		rowStyle := lipgloss.NewStyle()
		if i == m.cursor {
			rowStyle = rowStyle.Background(lipgloss.Color("62"))
		}

		confidence := fmt.Sprintf("%.0f%%", entity.Confidence*100)
		row := rowStyle.Render(
			fmt.Sprintf("%-6d %-15s %-40s %-10s",
				entity.ID,
				entity.Type,
				truncate(entity.Name, 40),
				confidence,
			),
		)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// getVisibleEntities returns entities after applying filters and search
func (m *EntityListModel) getVisibleEntities() []Entity {
	filtered := m.entities

	// Apply type filter
	if m.filterType != "" {
		filtered = filterEntitiesByType(filtered, m.filterType)
	}

	// Apply search
	if m.searchQuery != "" {
		filtered = searchEntitiesByName(filtered, m.searchQuery)
	}

	return filtered
}

// Helper functions

func renderEntityTable(entities []Entity) string {
	if len(entities) == 0 {
		return "No entities"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-6s %-15s %-40s %-10s\n", "ID", "Type", "Name", "Confidence"))
	b.WriteString(strings.Repeat("-", 75) + "\n")

	for _, entity := range entities {
		confidence := fmt.Sprintf("%.0f%%", entity.Confidence*100)
		b.WriteString(fmt.Sprintf("%-6d %-15s %-40s %-10s\n",
			entity.ID,
			entity.Type,
			truncate(entity.Name, 40),
			confidence,
		))
	}

	return b.String()
}

func calculatePagination(totalItems, pageSize, currentPage int) (start, end int) {
	start = currentPage * pageSize
	if start > totalItems {
		start = totalItems
	}

	end = start + pageSize
	if end > totalItems {
		end = totalItems
	}

	return start, end
}

func filterEntitiesByType(entities []Entity, filterType string) []Entity {
	if filterType == "" {
		return entities
	}

	var filtered []Entity
	for _, entity := range entities {
		if entity.Type == filterType {
			filtered = append(filtered, entity)
		}
	}
	return filtered
}

func searchEntitiesByName(entities []Entity, searchQuery string) []Entity {
	if searchQuery == "" {
		return entities
	}

	query := strings.ToLower(searchQuery)
	var results []Entity

	for _, entity := range entities {
		name := strings.ToLower(entity.Name)
		if strings.Contains(name, query) {
			results = append(results, entity)
		}
	}

	return results
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
