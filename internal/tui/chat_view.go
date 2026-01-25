package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/Blogem/enron-graph/internal/chat"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ChatViewModel represents the chat view state
type ChatViewModel struct {
	messages    []ChatMessage
	input       string
	cursor      int
	loading     bool
	handler     chat.Handler
	chatContext chat.Context
	ctx         context.Context
	// For visualize functionality
	lastResults []chat.Entity
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role          string // "user" or "assistant"
	Content       string
	ResultType    string // "entity", "path", "count", "text"
	Entities      []chat.Entity
	Path          []chat.PathNode
	ShowVisualize bool
}

// llmResponseMsg is sent when LLM responds
type llmResponseMsg struct {
	response   string
	entities   []chat.Entity
	path       []chat.PathNode
	resultType string
	err        error
}

// NewChatViewModel creates a new chat view model with handler and repository
func NewChatViewModel(handler chat.Handler, chatContext chat.Context) *ChatViewModel {
	return &ChatViewModel{
		messages: []ChatMessage{
			{
				Role:    "assistant",
				Content: "Hello! I can help you explore the Enron email graph. Try asking:\n  • \"Who is Jeff Skilling?\"\n  • \"Who did Kenneth Lay email?\"\n  • \"How are Jeff Skilling and Sherron Watkins connected?\"\n  • \"Show me emails about Enron scandal\"",
			},
		},
		input:       "",
		cursor:      0,
		loading:     false,
		handler:     handler,
		chatContext: chatContext,
		ctx:         context.Background(),
		lastResults: []chat.Entity{},
	}
}

// SetHandler sets the chat handler
func (m *ChatViewModel) SetHandler(handler chat.Handler) {
	m.handler = handler
}

// Update handles messages for chat view
func (m *ChatViewModel) Update(msg tea.Msg) (*ChatViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.input != "" && !m.loading {
				// Send message
				userMessage := m.input
				m.messages = append(m.messages, ChatMessage{
					Role:    "user",
					Content: userMessage,
				})
				m.input = ""
				m.loading = true

				// Process query asynchronously using chat handler
				return m, m.processQuery(userMessage)
			}
		case "ctrl+l":
			// Clear history
			m.messages = []ChatMessage{
				{
					Role:    "assistant",
					Content: "Chat cleared. How can I help you?",
				},
			}
			m.chatContext.Clear()
			m.loading = false
			m.lastResults = []chat.Entity{}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if len(msg.String()) == 1 && !m.loading {
				m.input += msg.String()
			}
		}

	case llmResponseMsg:
		// Handle LLM response
		m.loading = false
		if msg.err != nil {
			m.messages = append(m.messages, ChatMessage{
				Role:    "assistant",
				Content: fmt.Sprintf("Error: %v", msg.err),
			})
		} else {
			chatMsg := ChatMessage{
				Role:       "assistant",
				Content:    msg.response,
				ResultType: msg.resultType,
				Entities:   msg.entities,
				Path:       msg.path,
			}

			// Show visualize button for entity or path results
			if msg.resultType == "entity" || msg.resultType == "path" {
				chatMsg.ShowVisualize = true
			}

			m.messages = append(m.messages, chatMsg)
			m.lastResults = msg.entities
		}
	}

	return m, nil
}

// processQuery creates a command that processes query using chat handler
func (m *ChatViewModel) processQuery(userMessage string) tea.Cmd {
	return func() tea.Msg {
		if m.handler == nil {
			return llmResponseMsg{
				response: "",
				err:      fmt.Errorf("Chat handler not configured. Please ensure the system is properly initialized."),
			}
		}

		// Process query through chat handler
		response, err := m.handler.ProcessQuery(m.ctx, userMessage, m.chatContext)
		if err != nil {
			return llmResponseMsg{
				response: "",
				err:      err,
			}
		}

		// Try to extract structured results from response for visualization
		// This is a simple extraction - in production you'd want more sophisticated parsing
		resultType := "text"
		var entities []chat.Entity
		var path []chat.PathNode

		// Check if response contains entity information
		if strings.Contains(strings.ToLower(response), "entity:") ||
			strings.Contains(strings.ToLower(response), "found:") {
			resultType = "entity"
			// Extract tracked entities from context
			trackedEntities := m.chatContext.GetTrackedEntities()
			for _, te := range trackedEntities {
				entities = append(entities, chat.Entity{
					ID:   te.ID,
					Name: te.Name,
					Type: te.Type,
				})
			}
		}

		if strings.Contains(strings.ToLower(response), "path:") ||
			strings.Contains(strings.ToLower(response), "connection:") {
			resultType = "path"
		}

		if strings.Contains(strings.ToLower(response), "count:") {
			resultType = "count"
		}

		return llmResponseMsg{
			response:   response,
			entities:   entities,
			path:       path,
			resultType: resultType,
			err:        nil,
		}
	}
}

// formatEntityCard formats an entity as a card display
func formatEntityCard(entity chat.Entity) string {
	var b strings.Builder

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)

	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	typeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	b.WriteString(nameStyle.Render(entity.Name))
	b.WriteString(" ")
	b.WriteString(typeStyle.Render(fmt.Sprintf("(%s)", entity.Type)))
	b.WriteString("\n")

	if len(entity.Properties) > 0 {
		for key, val := range entity.Properties {
			b.WriteString(fmt.Sprintf("  %s: %v\n", key, val))
		}
	}

	return cardStyle.Render(b.String())
}

// formatPathDisplay formats a relationship path
func formatPathDisplay(path []chat.PathNode) string {
	var b strings.Builder

	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))

	arrowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	for i, node := range path {
		if i > 0 {
			b.WriteString(" ")
			b.WriteString(arrowStyle.Render(fmt.Sprintf("--[%s]-->", node.Relationship)))
			b.WriteString(" ")
		}
		b.WriteString(pathStyle.Render(node.Entity.Name))
	}

	return b.String()
}

// View renders the chat view
func (m *ChatViewModel) View(width, height int) string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	b.WriteString(titleStyle.Render("Chat - Graph Q&A"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", width-2) + "\n")

	// Render message history (scrollable)
	messageHeight := height - 6

	// Calculate which messages to show
	startIdx := 0
	lineCount := 0
	for i := len(m.messages) - 1; i >= 0; i-- {
		msgLines := strings.Count(m.messages[i].Content, "\n") + 3
		if m.messages[i].ShowVisualize {
			msgLines += 2
		}
		if len(m.messages[i].Entities) > 0 {
			msgLines += len(m.messages[i].Entities) * 4
		}
		if lineCount+msgLines > messageHeight {
			startIdx = i + 1
			break
		}
		lineCount += msgLines
	}

	// Render messages
	for i := startIdx; i < len(m.messages); i++ {
		msg := m.messages[i]

		var style lipgloss.Style
		var prefix string
		if msg.Role == "user" {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
			prefix = "You: "
		} else {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			prefix = "AI: "
		}

		b.WriteString(style.Render(prefix))
		b.WriteString(msg.Content)
		b.WriteString("\n")

		// Display entity cards if present (T125)
		if len(msg.Entities) > 0 {
			b.WriteString("\n")
			for _, entity := range msg.Entities {
				b.WriteString(formatEntityCard(entity))
				b.WriteString("\n")
			}
		}

		// Display path if present (T125)
		if len(msg.Path) > 0 {
			b.WriteString("\n")
			b.WriteString(formatPathDisplay(msg.Path))
			b.WriteString("\n")
		}

		// Show visualize button (T125)
		if msg.ShowVisualize {
			visualizeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Italic(true)
			b.WriteString("\n")
			b.WriteString(visualizeStyle.Render("  [Press 'v' to visualize in graph view]"))
			b.WriteString("\n")
		}

		b.WriteString("\n")
	}

	// Show loading indicator (T126)
	if m.loading {
		loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Italic(true)

		spinner := "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
		b.WriteString(loadingStyle.Render(fmt.Sprintf("AI: %c Processing your query...", spinner[0])))
		b.WriteString("\n\n")
	}

	// Render input box
	b.WriteString(strings.Repeat("─", width-2) + "\n")
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	b.WriteString(inputStyle.Render("> " + m.input + "█"))
	b.WriteString("\n")

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	b.WriteString(helpStyle.Render("  [Enter: Send | Ctrl+L: Clear | Esc: Back]"))

	return b.String()
}
