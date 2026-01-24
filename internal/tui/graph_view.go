package tui

import (
	"fmt"
	"math"
	"strings"

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
		case "tab":
			// Cycle through nodes
			if len(m.graph.Nodes) > 0 {
				m.cursor = (m.cursor + 1) % len(m.graph.Nodes)
			}
		case "enter":
			// Select current node
			if m.cursor >= 0 && m.cursor < len(m.graph.Nodes) {
				m.selectedNode = m.graph.Nodes[m.cursor].ID
			}
		case "e":
			// Expand node - would trigger loading neighbors
			// This would send a command to load data
		case "b":
			// Back to entity list
			m.selectedNode = -1
		}
	}

	return m, nil
}

// View renders the graph view
func (m *GraphViewModel) View(width, height int) string {
	if len(m.graph.Nodes) == 0 {
		return "No graph data loaded.\nSelect an entity from the entity list to visualize.\n"
	}

	rendered := renderGraph(m.graph)

	// Add selection indicator
	if m.cursor >= 0 && m.cursor < len(m.graph.Nodes) {
		selected := m.graph.Nodes[m.cursor]
		rendered += fmt.Sprintf("\n\nSelected: [%s: %s]", selected.Type, selected.Name)
	}

	return rendered
}

// LoadSubgraph loads a subgraph centered on an entity
func (m *GraphViewModel) LoadSubgraph(entityID int) {
	// This would be called by the main app to load data
	// For now, it's a placeholder
	m.centerNodeID = entityID
}

// renderGraph renders the graph as ASCII art
func renderGraph(g *Graph) string {
	if len(g.Nodes) == 0 {
		return "Empty graph"
	}

	// Limit nodes to max 50
	nodes := g.Nodes
	if len(nodes) > 50 {
		nodes = nodes[:50]
	}

	var b strings.Builder

	// Calculate layout
	rows, cols := calculateTreeLayout(len(nodes))

	// Render nodes in tree layout
	nodeIdx := 0
	for row := 0; row < rows && nodeIdx < len(nodes); row++ {
		// Render nodes in this row
		rowNodes := []string{}
		for col := 0; col < cols && nodeIdx < len(nodes); col++ {
			node := nodes[nodeIdx]
			nodeStr := formatNode(node.Type, node.Name)
			rowNodes = append(rowNodes, nodeStr)
			nodeIdx++
		}
		b.WriteString(strings.Join(rowNodes, "  "))
		b.WriteString("\n")

		// Render edges for this row
		if row < rows-1 {
			for col := 0; col < len(rowNodes); col++ {
				// Find edges from this node
				sourceIdx := row*cols + col
				if sourceIdx < len(nodes) {
					sourceID := nodes[sourceIdx].ID
					for _, edge := range g.Edges {
						if edge.FromID == sourceID {
							edgeStr := formatEdge(edge.RelType)
							b.WriteString(edgeStr)
							b.WriteString(" ")
						}
					}
				}
			}
			b.WriteString("\n")
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
