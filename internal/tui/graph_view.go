package tui

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/Blogem/enron-graph/internal/graph"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// GraphViewModel represents the graph visualization view state
type GraphViewModel struct {
	graph        *Graph
	selectedNode int
	cursor       int
	centerNodeID int
	maxNodes     int
	expandNode   int         // ID of node to expand (center graph on)
	orderedNodes []GraphNode // Cached ordered list for consistent cursor navigation
}

// Graph represents nodes and edges for visualization
type Graph struct {
	Nodes []GraphNode
	Edges []GraphEdge
}

// GraphNode represents a node in the graph
type GraphNode struct {
	ID   int
	Type string
	Name string
}

// GraphEdge represents an edge in the graph
type GraphEdge struct {
	FromID  int
	ToID    int
	RelType string
}

// NewGraphViewModel creates a new graph view model
func NewGraphViewModel() *GraphViewModel {
	return &GraphViewModel{
		graph: &Graph{
			Nodes: []GraphNode{},
			Edges: []GraphEdge{},
		},
		selectedNode: -1,
		cursor:       0,
		centerNodeID: 0,
		maxNodes:     50,
	}
}

// Update handles messages for graph view
func (m *GraphViewModel) Update(msg tea.Msg) (*GraphViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			// Navigate to previous node
			if len(m.orderedNodes) > 0 {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.orderedNodes) - 1
				}
			}
		case "down", "j":
			// Navigate to next node
			if len(m.orderedNodes) > 0 {
				m.cursor = (m.cursor + 1) % len(m.orderedNodes)
			}
		case "enter":
			// Select current node
			if m.cursor >= 0 && m.cursor < len(m.orderedNodes) {
				m.selectedNode = m.orderedNodes[m.cursor].ID
			}
		case "e":
			// Expand node - center graph on selected node
			if m.cursor >= 0 && m.cursor < len(m.orderedNodes) {
				// Debug output to file
				f, _ := os.OpenFile("/tmp/tui-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if f != nil {
					fmt.Fprintf(f, "DEBUG EXPAND: cursor=%d, orderedNodes length=%d, nodeID=%d, nodeName=%s\n",
						m.cursor, len(m.orderedNodes), m.orderedNodes[m.cursor].ID, m.orderedNodes[m.cursor].Name)
					f.Close()
				}
				m.expandNode = m.orderedNodes[m.cursor].ID
			}
			// Navigation keys (b, esc) are handled by parent app.go
		}
	}

	return m, nil
}

// View renders the graph view
func (m *GraphViewModel) View(width, height int) string {
	if len(m.graph.Nodes) == 0 {
		return "No graph data loaded.\nSelect an entity from the entity list to visualize.\n"
	}

	return m.renderGraphWithSelection(width, height)
}

// LoadSubgraph loads a subgraph centered on an entity
func (m *GraphViewModel) LoadSubgraph(entityID int) {
	// This would be called by the main app to load data
	// For now, it's a placeholder
	m.centerNodeID = entityID
}

// LoadSubgraphData loads a subgraph centered on an entity with actual data
func (m *GraphViewModel) LoadSubgraphData(ctx context.Context, repo graph.Repository, entityID int) {
	m.centerNodeID = entityID

	// Clear existing graph
	m.graph.Nodes = []GraphNode{}
	m.graph.Edges = []GraphEdge{}

	// Load center entity
	entity, err := repo.FindEntityByID(ctx, entityID)
	if err != nil {
		return
	}

	// Add center node
	m.graph.Nodes = append(m.graph.Nodes, GraphNode{
		ID:   entity.ID,
		Type: entity.TypeCategory,
		Name: entity.Name,
	})

	// Load relationships using the entity's type category
	relationships, err := repo.FindRelationshipsByEntity(ctx, entity.TypeCategory, entity.ID)
	if err != nil {
		return
	}

	// Track which entities we've already added to avoid duplicates
	addedEntities := make(map[int]bool)
	addedEntities[entity.ID] = true

	// Add related entities and edges
	for _, rel := range relationships {
		var targetID int
		var fromID, toID int
		var targetType string

		if rel.FromID == entity.ID {
			targetID = rel.ToID
			targetType = rel.ToType
			fromID = entity.ID
			toID = rel.ToID
		} else {
			targetID = rel.FromID
			targetType = rel.FromType
			fromID = rel.FromID
			toID = entity.ID
		}

		// Skip relationships to emails
		if targetType == "email" {
			continue
		}

		// Add target entity if not already added
		if !addedEntities[targetID] {
			targetEntity, err := repo.FindEntityByID(ctx, targetID)
			if err == nil {
				m.graph.Nodes = append(m.graph.Nodes, GraphNode{
					ID:   targetEntity.ID,
					Type: targetEntity.TypeCategory,
					Name: targetEntity.Name,
				})
				addedEntities[targetID] = true
			}
		}

		// Add edge
		m.graph.Edges = append(m.graph.Edges, GraphEdge{
			FromID:  fromID,
			ToID:    toID,
			RelType: rel.Type,
		})
	}
}

// renderGraphWithSelection renders the graph with cursor highlighting
func (m *GraphViewModel) renderGraphWithSelection(width, height int) string {
	if len(m.graph.Nodes) == 0 {
		return "Empty graph"
	}

	var b strings.Builder

	// Build orderedNodes to match the exact rendering order
	if len(m.graph.Nodes) > 0 {
		centerNode := m.graph.Nodes[0]
		m.orderedNodes = []GraphNode{centerNode}

		// Group edges first
		outgoing := make(map[string][]GraphEdge)
		incoming := make(map[string][]GraphEdge)

		for _, edge := range m.graph.Edges {
			if edge.FromID == centerNode.ID {
				outgoing[edge.RelType] = append(outgoing[edge.RelType], edge)
			} else if edge.ToID == centerNode.ID {
				incoming[edge.RelType] = append(incoming[edge.RelType], edge)
			}
		}

		// Build ordered nodes in the same order they'll be rendered
		// First add outgoing nodes (sorted by relationship type)
		relTypes := make([]string, 0, len(outgoing))
		for relType := range outgoing {
			relTypes = append(relTypes, relType)
		}
		sort.Strings(relTypes)

		for _, relType := range relTypes {
			edges := outgoing[relType]
			for _, edge := range edges {
				for _, node := range m.graph.Nodes {
					if node.ID == edge.ToID {
						m.orderedNodes = append(m.orderedNodes, node)
						break
					}
				}
			}
		}

		// Then add incoming nodes (sorted by relationship type)
		relTypes = make([]string, 0, len(incoming))
		for relType := range incoming {
			relTypes = append(relTypes, relType)
		}
		sort.Strings(relTypes)

		for _, relType := range relTypes {
			edges := incoming[relType]
			for _, edge := range edges {
				for _, node := range m.graph.Nodes {
					if node.ID == edge.FromID {
						m.orderedNodes = append(m.orderedNodes, node)
						break
					}
				}
			}
		}
	}

	// Show current selection status
	if m.cursor >= 0 && m.cursor < len(m.orderedNodes) {
		selectedNode := m.orderedNodes[m.cursor]
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		b.WriteString(statusStyle.Render(fmt.Sprintf("Selected: %s (Node %d of %d) | ↑↓: Navigate | Enter: View Details | E: Expand\n\n", selectedNode.Name, m.cursor+1, len(m.orderedNodes))))
	}

	// First node is the center/selected node
	if len(m.graph.Nodes) > 0 {
		centerNode := m.graph.Nodes[0]

		// Render center node prominently
		centerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

		// Highlight if this is the selected node
		if m.cursor == 0 {
			centerStyle = centerStyle.Background(lipgloss.Color("62"))
		}

		centerText := fmt.Sprintf("%s: %s", centerNode.Type, centerNode.Name)
		b.WriteString(centerStyle.Render(centerText))
		b.WriteString("\n\n")

		// Group edges by relationship type
		outgoing := make(map[string][]GraphEdge)
		incoming := make(map[string][]GraphEdge)

		for _, edge := range m.graph.Edges {
			if edge.FromID == centerNode.ID {
				outgoing[edge.RelType] = append(outgoing[edge.RelType], edge)
			} else if edge.ToID == centerNode.ID {
				incoming[edge.RelType] = append(incoming[edge.RelType], edge)
			}
		}

		// Track node index for cursor highlighting
		nodeIndex := 1 // Start at 1 since center is 0

		// Render outgoing relationships
		if len(outgoing) > 0 {
			b.WriteString("Outgoing Relationships:\n")
			b.WriteString(strings.Repeat("─", 60) + "\n")

			// Sort relationship types for consistent ordering
			relTypes := make([]string, 0, len(outgoing))
			for relType := range outgoing {
				relTypes = append(relTypes, relType)
			}
			sort.Strings(relTypes)

			for _, relType := range relTypes {
				edges := outgoing[relType]
				relStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
				b.WriteString(relStyle.Render(fmt.Sprintf("  %s", relType)))
				b.WriteString("\n")

				for _, edge := range edges {
					// Find target node in ordered list
					for _, node := range m.orderedNodes {
						if node.ID == edge.ToID {
							nodeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(getColorForEntityType(node.Type)))

							// Highlight if this node is selected by cursor
							if nodeIndex == m.cursor {
								nodeStyle = nodeStyle.Background(lipgloss.Color("62")).Bold(true)
							}

							b.WriteString(fmt.Sprintf("    └─> %s\n", nodeStyle.Render(fmt.Sprintf("[%s] %s", node.Type, node.Name))))
							nodeIndex++
							break
						}
					}
				}
				b.WriteString("\n")
			}
		}

		// Render incoming relationships
		if len(incoming) > 0 {
			b.WriteString("Incoming Relationships:\n")
			b.WriteString(strings.Repeat("─", 60) + "\n")

			// Sort relationship types for consistent ordering
			relTypes := make([]string, 0, len(incoming))
			for relType := range incoming {
				relTypes = append(relTypes, relType)
			}
			sort.Strings(relTypes)

			for _, relType := range relTypes {
				edges := incoming[relType]
				relStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))
				b.WriteString(relStyle.Render(fmt.Sprintf("  %s", relType)))
				b.WriteString("\n")

				for _, edge := range edges {
					// Find source node in ordered list
					for _, node := range m.orderedNodes {
						if node.ID == edge.FromID {
							nodeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(getColorForEntityType(node.Type)))

							// Highlight if this node is selected by cursor
							if nodeIndex == m.cursor {
								nodeStyle = nodeStyle.Background(lipgloss.Color("62")).Bold(true)
							}

							b.WriteString(fmt.Sprintf("    <─┘ %s\n", nodeStyle.Render(fmt.Sprintf("[%s] %s", node.Type, node.Name))))
							nodeIndex++
							break
						}
					}
				}
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

// renderGraph renders the graph as ASCII art in a radial/hierarchical layout
func renderGraph(g *Graph) string {
	if len(g.Nodes) == 0 {
		return "Empty graph"
	}

	var b strings.Builder

	// First node is the center/selected node
	if len(g.Nodes) > 0 {
		centerNode := g.Nodes[0]

		// Render center node prominently
		centerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

		centerText := fmt.Sprintf("%s: %s", centerNode.Type, centerNode.Name)
		b.WriteString(centerStyle.Render(centerText))
		b.WriteString("\n\n")

		// Group edges by relationship type
		outgoing := make(map[string][]GraphEdge)
		incoming := make(map[string][]GraphEdge)

		for _, edge := range g.Edges {
			if edge.FromID == centerNode.ID {
				outgoing[edge.RelType] = append(outgoing[edge.RelType], edge)
			} else if edge.ToID == centerNode.ID {
				incoming[edge.RelType] = append(incoming[edge.RelType], edge)
			}
		}

		// Render outgoing relationships
		if len(outgoing) > 0 {
			b.WriteString("Outgoing Relationships:\n")
			b.WriteString(strings.Repeat("─", 60) + "\n")

			for relType, edges := range outgoing {
				relStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
				b.WriteString(relStyle.Render(fmt.Sprintf("  %s", relType)))
				b.WriteString("\n")

				for _, edge := range edges {
					// Find target node
					for _, node := range g.Nodes {
						if node.ID == edge.ToID {
							nodeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(getColorForEntityType(node.Type)))
							b.WriteString(fmt.Sprintf("    └─> %s\n", nodeStyle.Render(fmt.Sprintf("[%s] %s", node.Type, node.Name))))
							break
						}
					}
				}
				b.WriteString("\n")
			}
		}

		// Render incoming relationships
		if len(incoming) > 0 {
			b.WriteString("Incoming Relationships:\n")
			b.WriteString(strings.Repeat("─", 60) + "\n")

			for relType, edges := range incoming {
				relStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))
				b.WriteString(relStyle.Render(fmt.Sprintf("  %s", relType)))
				b.WriteString("\n")

				for _, edge := range edges {
					// Find source node
					for _, node := range g.Nodes {
						if node.ID == edge.FromID {
							nodeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(getColorForEntityType(node.Type)))
							b.WriteString(fmt.Sprintf("    <─┘ %s\n", nodeStyle.Render(fmt.Sprintf("[%s] %s", node.Type, node.Name))))
							break
						}
					}
				}
				b.WriteString("\n")
			}
		}

		if len(outgoing) == 0 && len(incoming) == 0 {
			b.WriteString("No relationships found.\n")
		}
	}

	return b.String()
}

// formatNode formats a node as [Type: Name] with color
func formatNode(entityType, entityName string) string {
	color := getColorForEntityType(entityType)
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))

	nodeText := fmt.Sprintf("[%s: %s]", entityType, truncate(entityName, 20))
	return style.Render(nodeText)
}

// formatEdge formats an edge as ---[REL_TYPE]-->
func formatEdge(relType string) string {
	return fmt.Sprintf("---[%s]-->", relType)
}

// calculateTreeLayout calculates rows and columns for tree layout
func calculateTreeLayout(nodeCount int) (rows, columns int) {
	if nodeCount == 0 {
		return 0, 0
	}
	if nodeCount == 1 {
		return 1, 1
	}

	// Approximate square layout
	columns = int(math.Ceil(math.Sqrt(float64(nodeCount))))
	rows = int(math.Ceil(float64(nodeCount) / float64(columns)))

	return rows, columns
}

// getColorForEntityType returns the color for an entity type
func getColorForEntityType(entityType string) string {
	switch entityType {
	case "person":
		return "blue"
	case "organization":
		return "green"
	case "concept":
		return "yellow"
	case "email":
		return "gray"
	default:
		return "white"
	}
}
