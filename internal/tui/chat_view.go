package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ChatViewModel represents the chat view state
type ChatViewModel struct {
	messages []ChatMessage
	input    string
	cursor   int
	loading  bool
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string // "user" or "assistant"
	Content string
}

// NewChatViewModel creates a new chat view model
func NewChatViewModel() *ChatViewModel {
	return &ChatViewModel{
		messages: []ChatMessage{
			{
				Role:    "assistant",
				Content: "Hello! I can help you explore the Enron email graph. Try asking about people, relationships, or concepts.",
			},
		},
		input:   "",
		cursor:  0,
		loading: false,
	}
}

// Update handles messages for chat view
func (m *ChatViewModel) Update(msg tea.Msg) (*ChatViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.input != "" && !m.loading {
				// Send message
				m.messages = append(m.messages, ChatMessage{
					Role:    "user",
					Content: m.input,
				})
				m.input = ""
				m.loading = true
				// Would trigger LLM processing here
			}
		case "ctrl+l":
			// Clear history
			m.messages = []ChatMessage{
				{
					Role:    "assistant",
					Content: "Chat cleared. How can I help you?",
				},
			}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if len(msg.String()) == 1 && !m.loading {
				m.input += msg.String()
			}
		}
	}

	return m, nil
}

// View renders the chat view
func (m *ChatViewModel) View(width, height int) string {
	var b strings.Builder

	// Render message history
	messageHeight := height - 4
	b.WriteString("Chat History:\n")
	b.WriteString(strings.Repeat("=", width-2) + "\n")

	// Show last N messages that fit
	startIdx := 0
	if len(m.messages) > messageHeight/3 {
		startIdx = len(m.messages) - messageHeight/3
	}

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

		b.WriteString(style.Render(prefix + msg.Content))
		b.WriteString("\n\n")
	}

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("AI: Thinking..."))
		b.WriteString("\n\n")
	}

	// Render input box
	b.WriteString(strings.Repeat("=", width-2) + "\n")
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	b.WriteString(inputStyle.Render("> " + m.input + "_"))
	b.WriteString("\n")

	return b.String()
}
