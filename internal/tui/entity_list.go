package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EntityListModel represents the entity list view state
type EntityListModel struct {
	entities     []Entity
	cursor       int
	page         int
	pageSize     int
	filterType   string
	searchQuery  string
	selectedID   int
	isSearching  bool
	isFiltering  bool
	filterCursor int // Cursor position in filter menu
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
		pageSize:    5,
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
				m.updatePage()
			}
		case "down", "j":
			visible := m.getVisibleEntities()
			if m.cursor < len(visible)-1 {
				m.cursor++
				m.updatePage()
			}
		case "pgup":
			m.cursor -= m.pageSize
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.updatePage()
		case "pgdown":
			visible := m.getVisibleEntities()
			m.cursor += m.pageSize
			if m.cursor >= len(visible) {
				m.cursor = len(visible) - 1
			}
			m.updatePage()
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

// getFilterOptions returns the available filter options based on actual entity types
func (m *EntityListModel) getFilterOptions() []string {
	// Start with "All"
	options := []string{"All"}

	// Collect unique entity types
	typeMap := make(map[string]bool)
	for _, entity := range m.entities {
		if entity.Type != "" {
			typeMap[entity.Type] = true
		}
	}

	// Add unique types to options
	for entityType := range typeMap {
		options = append(options, entityType)
	}

	return options
}

// handleFilterInput handles keyboard input during filtering
func (m *EntityListModel) handleFilterInput(msg tea.KeyMsg) (*EntityListModel, tea.Cmd) {
	filterOptions := m.getFilterOptions()

	switch msg.String() {
	case "enter":
		// Apply selected filter
		if m.filterCursor == 0 {
			m.filterType = "" // "All" means no filter
		} else {
			m.filterType = filterOptions[m.filterCursor]
		}
		m.isFiltering = false
		m.cursor = 0
		m.page = 0
	case "esc":
		m.isFiltering = false
		m.cursor = 0
	case "up", "k":
		m.filterCursor--
		if m.filterCursor < 0 {
			m.filterCursor = len(filterOptions) - 1
		}
	case "down", "j":
		m.filterCursor++
		if m.filterCursor >= len(filterOptions) {
			m.filterCursor = 0
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
		b.WriteString(fmt.Sprintf("Search (Esc to cancel): %s_\n", m.searchQuery))
	}
	if m.isFiltering {
		// Show filter menu
		filterOptions := m.getFilterOptions()
		b.WriteString("Filter by type (↑↓ to select, Enter to apply, Esc to cancel):\n")
		for i, option := range filterOptions {
			if i == m.filterCursor {
				b.WriteString(fmt.Sprintf("  > %s\n", option))
			} else {
				b.WriteString(fmt.Sprintf("    %s\n", option))
			}
		}
		b.WriteString("\n")
	}

	// Render table
	visible := m.getVisibleEntities()
	table := m.renderTable(visible, width, height-5)
	b.WriteString(table)

	// Show pagination info
	start, end := calculatePagination(len(visible), m.pageSize, m.page)
	b.WriteString(fmt.Sprintf("\nShowing %d-%d of %d entities (Page %d)\n", start+1, end, len(visible), m.page+1))

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
		// Adjust cursor check for page offset
		if start+i == m.cursor {
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

// updatePage calculates which page the cursor is on
func (m *EntityListModel) updatePage() {
	m.page = m.cursor / m.pageSize
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
