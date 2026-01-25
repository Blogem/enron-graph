package tui

import (
	"strings"
	"testing"

	"github.com/Blogem/enron-graph/internal/graph"
	tea "github.com/charmbracelet/bubbletea"
)

// TestTUIDisplaysEntitiesAndRelationships verifies T110: TUI displays entities and relationships
func TestTUIDisplaysEntitiesAndRelationships(t *testing.T) {
	// Create model with test data
	model := NewModel(graph.NewMockRepository())

	// Load test entities
	testEntities := []Entity{
		{ID: 1, Type: "person", Name: "Jeff Skilling", Confidence: 0.95},
		{ID: 2, Type: "organization", Name: "Enron Corp", Confidence: 0.88},
		{ID: 3, Type: "person", Name: "Kenneth Lay", Confidence: 0.92},
	}
	model.LoadEntities(testEntities)

	// Initialize with window size
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model = updated.(Model)

	// Render the view
	view := model.View()

	// Verify entities are displayed
	if !strings.Contains(view, "Jeff Skilling") {
		t.Error("TUI should display entity 'Jeff Skilling'")
	}
	if !strings.Contains(view, "Enron Corp") {
		t.Error("TUI should display entity 'Enron Corp'")
	}
	if !strings.Contains(view, "Kenneth Lay") {
		t.Error("TUI should display entity 'Kenneth Lay'")
	}

	// Verify entity types are displayed
	if !strings.Contains(view, "person") {
		t.Error("TUI should display entity type 'person'")
	}
	if !strings.Contains(view, "organization") {
		t.Error("TUI should display entity type 'organization'")
	}

	// Verify confidence scores are displayed
	if !strings.Contains(view, "95%") || !strings.Contains(view, "88%") || !strings.Contains(view, "92%") {
		t.Error("TUI should display confidence scores")
	}

	// Verify navigation hints are present
	if !strings.Contains(view, "Navigate") || !strings.Contains(view, "Filter") || !strings.Contains(view, "Search") {
		t.Error("TUI should display navigation hints")
	}
}

// TestGraphViewRendersNodesAndEdges verifies T111: Graph view renders ASCII graph with nodes/edges
func TestGraphViewRendersNodesAndEdges(t *testing.T) {
	// Create graph view with test data
	graphView := NewGraphViewModel()

	// Create test graph
	testGraph := &Graph{
		Nodes: []GraphNode{
			{ID: 1, Type: "person", Name: "Jeff Skilling"},
			{ID: 2, Type: "organization", Name: "Enron Corp"},
			{ID: 3, Type: "person", Name: "Kenneth Lay"},
		},
		Edges: []GraphEdge{
			{FromID: 1, ToID: 2, RelType: "WORKED_AT"},
			{FromID: 3, ToID: 2, RelType: "CEO_OF"},
		},
	}
	graphView.graph = testGraph

	// Render the graph view
	view := graphView.View(120, 40)

	// Debug: print the view to see what's actually rendered
	t.Logf("Graph view output:\n%s", view)

	// Verify nodes are rendered (implementation uses "[Type] Name" format in structured display)
	if !strings.Contains(view, "person") {
		t.Error("Graph view should render nodes with entity types")
	}
	if !strings.Contains(view, "Jeff Skilling") {
		t.Error("Graph view should display node 'Jeff Skilling'")
	}
	if !strings.Contains(view, "Enron Corp") {
		t.Error("Graph view should display node 'Enron Corp'")
	}

	// Verify edges are rendered (implementation uses relationship sections)
	if !strings.Contains(view, "Outgoing") || !strings.Contains(view, "WORKED_AT") {
		t.Error("Graph view should render outgoing relationships")
	}
}

// TestSelectingNodeShowsProperties verifies T112: Selecting node shows properties and expansion option
func TestSelectingNodeShowsProperties(t *testing.T) {
	// Create model with test data
	model := NewModel(graph.NewMockRepository())
	testEntities := []Entity{
		{ID: 1, Type: "person", Name: "Jeff Skilling", Confidence: 0.95},
		{ID: 2, Type: "organization", Name: "Enron Corp", Confidence: 0.88},
	}
	model.LoadEntities(testEntities)

	// Initialize with window size
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model = updated.(Model)

	// Simulate selecting an entity (press Enter on entity list)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	// Should now be in detail view
	if model.currentView != ViewDetailView {
		t.Error("Selecting entity should switch to detail view")
	}

	// Render detail view
	view := model.View()

	// Verify entity properties are displayed
	if !strings.Contains(view, "Properties") {
		t.Error("Detail view should show 'Properties' section")
	}
	if !strings.Contains(view, "Type:") {
		t.Error("Detail view should show entity type")
	}
	if !strings.Contains(view, "Name:") {
		t.Error("Detail view should show entity name")
	}
	if !strings.Contains(view, "Confidence:") {
		t.Error("Detail view should show confidence score")
	}

	// Verify expansion/visualization option is available
	if !strings.Contains(view, "Visualize") || !strings.Contains(view, "V:") {
		t.Error("Detail view should show visualization option (V key)")
	}

	// Verify relationships section is present
	if !strings.Contains(view, "Relationships") {
		t.Error("Detail view should show 'Relationships' section")
	}
}

// TestFilteringByEntityType verifies T113: Controls for filtering by entity type work
func TestFilteringByEntityType(t *testing.T) {
	// Create model with mixed entity types
	model := NewModel(graph.NewMockRepository())
	testEntities := []Entity{
		{ID: 1, Type: "person", Name: "Jeff Skilling", Confidence: 0.95},
		{ID: 2, Type: "organization", Name: "Enron Corp", Confidence: 0.88},
		{ID: 3, Type: "person", Name: "Kenneth Lay", Confidence: 0.92},
		{ID: 4, Type: "concept", Name: "energy trading", Confidence: 0.85},
		{ID: 5, Type: "person", Name: "Andrew Fastow", Confidence: 0.90},
	}
	model.LoadEntities(testEntities)

	// Initialize with window size
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model = updated.(Model)

	// Verify initial view shows all entities
	view := model.View()
	t.Logf("Initial view:\n%s", view)
	if !strings.Contains(view, "of 5 entities") {
		t.Error("Should initially show all 5 entities")
	}

	// Simulate pressing 'F' key to activate filter
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	model = updated.(Model)

	// Verify filter mode is activated
	if !model.entityList.isFiltering {
		t.Error("Pressing 'F' should activate filter mode")
	}

	// Navigate down to "person" option (sorted: All, concept, organization, person)
	// Filter cursor starts at 0 (All), so press down 3 times to get to "person"
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)

	// Press Enter to apply filter
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	// Verify filter is applied
	if model.entityList.filterType != "person" {
		t.Errorf("Filter type should be 'person', got '%s'", model.entityList.filterType)
	}

	// Render filtered view
	view = model.View()
	t.Logf("Filtered view:\n%s", view)

	// Verify only person entities are shown (3 filtered results)
	if !strings.Contains(view, "of 3 entities") {
		t.Error("Filter should show 3 person entities")
	}

	// Verify person entities are visible
	if !strings.Contains(view, "Jeff Skilling") {
		t.Error("Filtered view should show 'Jeff Skilling'")
	}
	if !strings.Contains(view, "Kenneth Lay") {
		t.Error("Filtered view should show 'Kenneth Lay'")
	}
	if !strings.Contains(view, "Andrew Fastow") {
		t.Error("Filtered view should show 'Andrew Fastow'")
	}

	// Verify correct number of visible entities
	visible := model.entityList.getVisibleEntities()
	if len(visible) != 3 {
		t.Errorf("Should have 3 visible entities after filter, got %d", len(visible))
	}

	// Verify all visible entities are of type "person"
	for _, entity := range visible {
		if entity.Type != "person" {
			t.Errorf("Visible entity '%s' should be type 'person', got '%s'", entity.Name, entity.Type)
		}
	}
}

// TestNavigationBetweenViews verifies view switching works
func TestNavigationBetweenViews(t *testing.T) {
	// Create model
	model := NewModel(graph.NewMockRepository())
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model = updated.(Model)

	// Initially should be in entity list view
	if model.currentView != ViewEntityList {
		t.Error("Should start in entity list view")
	}

	// Press '2' to switch to graph view
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(Model)
	if model.currentView != ViewGraphView {
		t.Error("Pressing '2' should switch to graph view")
	}

	// Press '3' to switch to detail view
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model = updated.(Model)
	if model.currentView != ViewDetailView {
		t.Error("Pressing '3' should switch to detail view")
	}

	// Press '4' to switch to chat view
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	model = updated.(Model)
	if model.currentView != ViewChatView {
		t.Error("Pressing '4' should switch to chat view")
	}

	// Press '1' to switch back to entity list
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	model = updated.(Model)
	if model.currentView != ViewEntityList {
		t.Error("Pressing '1' should switch back to entity list view")
	}

	// Press Tab to cycle to next view
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	if model.currentView != ViewGraphView {
		t.Error("Tab should cycle to next view (graph view)")
	}
}
