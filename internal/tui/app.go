package tui

import (
	"context"

	"github.com/Blogem/enron-graph/internal/chat"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/pkg/llm"
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

	// Repository for data access
	repo      graph.Repository
	ctx       context.Context
	llmClient llm.Client

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
func NewModel(repo graph.Repository) Model {
	// Create chat components
	chatContext := chat.NewContext()
	var chatHandler chat.Handler

	// Handler will be set when LLM client is available
	chatView := NewChatViewModel(chatHandler, chatContext)

	return Model{
		currentView: ViewEntityList,
		repo:        repo,
		ctx:         context.Background(),
		llmClient:   nil, // Will be set later if available
		entityList:  NewEntityListModel(),
		graphView:   NewGraphViewModel(),
		detailView:  NewDetailViewModel(),
		chatView:    chatView,
	}
}

// LoadEntities loads entities into the entity list
func (m *Model) LoadEntities(entities []Entity) {
	m.entityList.entities = entities
}

// SetLLMClient sets the LLM client for chat functionality
func (m *Model) SetLLMClient(client llm.Client) {
	m.llmClient = client

	// Create chat handler with LLM client and repository adapter
	chatRepo := newChatRepositoryAdapter(m.repo)
	chatHandler := chat.NewHandler(client, chatRepo)
	m.chatView.SetHandler(chatHandler)
}

// Init initializes the model (required by Bubble Tea)
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model (required by Bubble Tea)
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global keybindings first
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			// Quit unless in chat view (where 'q' is input)
			if m.currentView != ViewChatView {
				return m, tea.Quit
			}
		case "tab":
			// Switch between views with Tab
			m.currentView = (m.currentView + 1) % 4
			// Reset view-specific state to prevent interference
			m.entityList.selectedID = 0
			m.graphView.selectedNode = -1
			return m, nil
		case "shift+tab":
			// Switch between views backwards with Shift+Tab
			m.currentView = (m.currentView + 3) % 4
			// Reset view-specific state to prevent interference
			m.entityList.selectedID = 0
			m.graphView.selectedNode = -1
			return m, nil
		case "1":
			m.currentView = ViewEntityList
			// Reset selectedID to prevent auto-jump to details
			m.entityList.selectedID = 0
			// Reset graph view state
			m.graphView.selectedNode = -1
			return m, nil
		case "2":
			m.currentView = ViewGraphView
			// Reset entity list selection state
			m.entityList.selectedID = 0
			// Load graph data for selected entity if available
			if m.selectedEntityID > 0 {
				m.graphView.LoadSubgraphData(m.ctx, m.repo, m.selectedEntityID)
			}
			return m, nil
		case "3":
			m.currentView = ViewDetailView
			// Reset other view states
			m.entityList.selectedID = 0
			m.graphView.selectedNode = -1
			return m, nil
		case "4", "c":
			m.currentView = ViewChatView
			// Reset other view states
			m.entityList.selectedID = 0
			m.graphView.selectedNode = -1
			return m, nil
		}
		// If not a global key, fall through to view-specific handling
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
	help := "Tab: Switch View | C: Chat | Q: Quit"

	switch m.currentView {
	case ViewEntityList:
		help += " | ↑↓: Navigate | F: Filter | /: Search | Enter: Details"
	case ViewGraphView:
		help += " | ↑↓: Navigate Nodes | E: Expand | Enter: Details | B: Back"
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

		// Load entity details from repository
		entity, err := m.repo.FindEntityByID(m.ctx, m.selectedEntityID)
		if err == nil {
			m.detailView.LoadEntityData(entity.ID, entity.TypeCategory, entity.Name,
				entity.Properties, entity.ConfidenceScore)

			// Load relationships using the entity's type category
			relationships, err := m.repo.FindRelationshipsByEntity(m.ctx, entity.TypeCategory, entity.ID)
			if err == nil {
				m.detailView.LoadRelationships(m.ctx, m.repo, relationships)
			}
		} else {
			m.detailView.LoadEntity(m.selectedEntityID)
		}

		// Reset selectedID after transitioning to prevent auto-jump on next navigation
		m.entityList.selectedID = 0
	}

	return m, cmd
}

func (m Model) updateGraphView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	updatedGraph, cmd := m.graphView.Update(msg)
	m.graphView = updatedGraph

	// Check for view transitions (only handle at parent level)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "b", "esc":
			// Back to entity list
			m.currentView = ViewEntityList
			m.graphView.selectedNode = -1
			return m, cmd
		}
	}

	// Check if user wants to expand a node
	if m.graphView.expandNode > 0 {
		// Reload graph centered on the expanded node
		m.graphView.LoadSubgraphData(m.ctx, m.repo, m.graphView.expandNode)
		// Reset cursor to center node (0)
		m.graphView.cursor = 0
		// Clear expand request
		m.graphView.expandNode = 0
		return m, cmd
	}

	// Check if entity was selected
	if m.graphView.selectedNode > 0 {
		m.selectedEntityID = m.graphView.selectedNode
		m.currentView = ViewDetailView

		// Load entity details from repository
		entity, err := m.repo.FindEntityByID(m.ctx, m.selectedEntityID)
		if err == nil {
			m.detailView.LoadEntityData(entity.ID, entity.TypeCategory, entity.Name,
				entity.Properties, entity.ConfidenceScore)

			// Load relationships using the entity's type category
			relationships, err := m.repo.FindRelationshipsByEntity(m.ctx, entity.TypeCategory, entity.ID)
			if err == nil {
				m.detailView.LoadRelationships(m.ctx, m.repo, relationships)
			}
		} else {
			m.detailView.LoadEntity(m.selectedEntityID)
		}

		// Reset selectedNode after transitioning to prevent auto-jump on next navigation
		m.graphView.selectedNode = -1
	}

	return m, cmd
}

func (m Model) updateDetailView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	updatedDetail, cmd := m.detailView.Update(msg)
	m.detailView = updatedDetail

	// Check if user selected a relationship to navigate to
	if m.detailView.selectedEntityID > 0 {
		m.selectedEntityID = m.detailView.selectedEntityID

		// Load entity details from repository
		entity, err := m.repo.FindEntityByID(m.ctx, m.selectedEntityID)
		if err == nil {
			m.detailView.LoadEntityData(entity.ID, entity.TypeCategory, entity.Name,
				entity.Properties, entity.ConfidenceScore)

			// Load relationships using the entity's type category
			relationships, err := m.repo.FindRelationshipsByEntity(m.ctx, entity.TypeCategory, entity.ID)
			if err == nil {
				m.detailView.LoadRelationships(m.ctx, m.repo, relationships)
			}
		} else {
			m.detailView.LoadEntity(m.selectedEntityID)
		}

		// Reset selectedEntityID after loading new details
		m.detailView.selectedEntityID = 0
		// Reset cursor to top of new entity's relationships
		m.detailView.cursor = 0
		return m, cmd
	}

	// Check for view transitions (only handle at parent level)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "v":
			// Switch to graph view with loaded data
			m.currentView = ViewGraphView
			m.graphView.LoadSubgraphData(m.ctx, m.repo, m.selectedEntityID)
			return m, cmd
		case "b", "esc":
			// Back to entity list - clear selection state
			m.currentView = ViewEntityList
			m.entityList.selectedID = 0
			return m, cmd
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
