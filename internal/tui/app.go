package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ViewMode represents the current view state
type ViewMode int

const (
	ViewEntityList ViewMode = iota
	ViewGraphView
	ViewDetailView
	ViewChatView
)

// Model is the main application state container for Bubble Tea
type Model struct {
	currentView ViewMode
	width       int
	height      int

	// View-specific states
	entityList *EntityListModel
	graphView  *GraphViewModel
	detailView *DetailViewModel
	chatView   *ChatViewModel

	// Shared state
	selectedEntityID int
	err              error
}

// NewModel creates a new TUI model
func NewModel() Model {
	return Model{
		currentView: ViewEntityList,
		entityList:  NewEntityListModel(),
		graphView:   NewGraphViewModel(),
		detailView:  NewDetailViewModel(),
		chatView:    NewChatViewModel(),
	}
}

// LoadEntities loads entities into the entity list
func (m *Model) LoadEntities(entities []Entity) {
	m.entityList.entities = entities
}

// Init initializes the model (required by Bubble Tea)
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model (required by Bubble Tea)
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Quit unless in chat view (where 'q' is input)
			if m.currentView != ViewChatView {
				return m, tea.Quit
			}
		case "tab":
			// Switch between views
			m.currentView = (m.currentView + 1) % 4
			return m, nil
		case "1":
			m.currentView = ViewEntityList
			return m, nil
		case "2":
			m.currentView = ViewGraphView
			return m, nil
		case "3":
			m.currentView = ViewDetailView
			return m, nil
		case "4":
			m.currentView = ViewChatView
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Delegate to current view's update handler
	switch m.currentView {
	case ViewEntityList:
		return m.updateEntityList(msg)
	case ViewGraphView:
		return m.updateGraphView(msg)
	case ViewDetailView:
		return m.updateDetailView(msg)
	case ViewChatView:
		return m.updateChatView(msg)
	}

	return m, nil
}

// View renders the current view (required by Bubble Tea)
func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	// Render header
	header := m.renderHeader()

	// Render current view
	var content string
	switch m.currentView {
	case ViewEntityList:
		content = m.entityList.View(m.width, m.height-4)
	case ViewGraphView:
		content = m.graphView.View(m.width, m.height-4)
	case ViewDetailView:
		content = m.detailView.View(m.width, m.height-4)
	case ViewChatView:
		content = m.chatView.View(m.width, m.height-4)
	}

	// Render footer
	footer := m.renderFooter()

	return header + "\n" + content + "\n" + footer
}

// renderHeader renders the application header with view tabs
func (m Model) renderHeader() string {
	tabs := []string{"1:Entities", "2:Graph", "3:Details", "4:Chat"}
	header := "Enron Knowledge Graph Explorer\n"

	tabBar := ""
	for i, tab := range tabs {
		if ViewMode(i) == m.currentView {
			tabBar += "[" + tab + "] "
		} else {
			tabBar += " " + tab + "  "
		}
	}

	return header + tabBar
}

// renderFooter renders the application footer with keybindings
func (m Model) renderFooter() string {
	help := "Tab: Switch View | Q: Quit"

	switch m.currentView {
	case ViewEntityList:
		help += " | ↑↓: Navigate | F: Filter | /: Search | Enter: Details"
	case ViewGraphView:
		help += " | Tab: Select Node | E: Expand | Enter: Details | B: Back"
	case ViewDetailView:
		help += " | V: Visualize | B: Back"
	case ViewChatView:
		help += " | Enter: Send | Ctrl+L: Clear"
	}

	return help
}

// View-specific update methods (delegated from main Update)

func (m Model) updateEntityList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	updatedList, cmd := m.entityList.Update(msg)
	m.entityList = updatedList

	// Check if entity was selected
	if m.entityList.selectedID > 0 {
		m.selectedEntityID = m.entityList.selectedID
		m.currentView = ViewDetailView
		m.detailView.LoadEntity(m.selectedEntityID)
	}

	return m, cmd
}

func (m Model) updateGraphView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	updatedGraph, cmd := m.graphView.Update(msg)
	m.graphView = updatedGraph

	// Check if entity was selected
	if m.graphView.selectedNode > 0 {
		m.selectedEntityID = m.graphView.selectedNode
		m.currentView = ViewDetailView
		m.detailView.LoadEntity(m.selectedEntityID)
	}

	return m, cmd
}

func (m Model) updateDetailView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	updatedDetail, cmd := m.detailView.Update(msg)
	m.detailView = updatedDetail

	// Check for view transitions
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "v":
			// Switch to graph view
			m.currentView = ViewGraphView
			m.graphView.LoadSubgraph(m.selectedEntityID)
		case "b", "esc":
			// Back to entity list
			m.currentView = ViewEntityList
		}
	}

	return m, cmd
}

func (m Model) updateChatView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	updatedChat, cmd := m.chatView.Update(msg)
	m.chatView = updatedChat
	return m, cmd
}
