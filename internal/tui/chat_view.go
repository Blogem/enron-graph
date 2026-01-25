package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/Blogem/enron-graph/pkg/llm"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ChatViewModel represents the chat view state
type ChatViewModel struct {
	messages  []ChatMessage
	input     string
	cursor    int
	loading   bool
	llmClient llm.Client
	ctx       context.Context
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string // "user" or "assistant"
	Content string
}

// llmResponseMsg is sent when LLM responds
type llmResponseMsg struct {
	response string
	err      error
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
		ctx:     context.Background(),
	}
}

// SetLLMClient sets the LLM client for the chat view
func (m *ChatViewModel) SetLLMClient(client llm.Client) {
	m.llmClient = client
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

				// Call LLM asynchronously
				return m, m.callLLM(userMessage)
			}
		case "ctrl+l":
			// Clear history
			m.messages = []ChatMessage{
				{
					Role:    "assistant",
					Content: "Chat cleared. How can I help you?",
				},
			}
			m.loading = false
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
			m.messages = append(m.messages, ChatMessage{
				Role:    "assistant",
				Content: msg.response,
			})
		}
	}

	return m, nil
}

// callLLM creates a command that calls the LLM asynchronously
func (m *ChatViewModel) callLLM(userMessage string) tea.Cmd {
	return func() tea.Msg {
		if m.llmClient == nil {
			return llmResponseMsg{
				response: "",
				err:      fmt.Errorf("LLM client not configured. Please ensure Ollama is running."),
			}
		}

		// Build context-aware prompt
		prompt := m.buildPrompt(userMessage)

		// Call LLM
		response, err := m.llmClient.GenerateCompletion(m.ctx, prompt)
		return llmResponseMsg{
			response: response,
			err:      err,
		}
	}
}

// buildPrompt builds a context-aware prompt for the LLM
func (m *ChatViewModel) buildPrompt(userMessage string) string {
	var b strings.Builder

	b.WriteString("You are an AI assistant helping users explore the Enron email knowledge graph. ")
	b.WriteString("The graph contains entities (people, organizations, locations, concepts) and relationships between them. ")
	b.WriteString("Provide helpful, concise answers about the graph structure and entities.\n\n")

	// Include recent conversation history (last 3 exchanges)
	historyStart := len(m.messages) - 6
	if historyStart < 0 {
		historyStart = 0
	}

	if historyStart > 0 {
		b.WriteString("Recent conversation:\n")
		for i := historyStart; i < len(m.messages); i++ {
			msg := m.messages[i]
			if msg.Role == "user" {
				b.WriteString(fmt.Sprintf("User: %s\n", msg.Content))
			} else {
				b.WriteString(fmt.Sprintf("Assistant: %s\n", msg.Content))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("User: %s\n", userMessage))
	b.WriteString("Assistant:")

	return b.String()
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
