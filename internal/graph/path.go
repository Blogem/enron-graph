package graph

import (
	"context"
	"fmt"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/relationship"
)

// PathNode represents a node in the path search
type PathNode struct {
	EntityID int
	Parent   *PathNode
	Relation *ent.Relationship
	Depth    int
}

// findShortestPathBFS uses breadth-first search to find the shortest path
func (r *entRepository) findShortestPathBFS(ctx context.Context, fromID, toID int) ([]*ent.Relationship, error) {
	if fromID == toID {
		return []*ent.Relationship{}, nil
	}

	visited := make(map[int]bool)
	queue := []*PathNode{{EntityID: fromID, Parent: nil, Relation: nil, Depth: 0}}
	visited[fromID] = true

	// Maximum depth to prevent infinite loops
	maxDepth := 10

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Check depth limit
		if current.Depth >= maxDepth {
			continue
		}

		// Get all relationships for current entity
		rels, err := r.client.Relationship.Query().
			Where(
				relationship.Or(
					relationship.FromIDEQ(current.EntityID),
					relationship.ToIDEQ(current.EntityID),
				),
			).
			All(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to query relationships: %w", err)
		}

		// Explore each relationship
		for _, rel := range rels {
			var neighborID int

			// Determine neighbor ID
			if rel.FromID == current.EntityID {
				neighborID = rel.ToID
			} else {
				neighborID = rel.FromID
			}

			// Check if we found the target
			if neighborID == toID {
				// Reconstruct path
				path := []*ent.Relationship{rel}
				node := current
				for node.Parent != nil {
					path = append([]*ent.Relationship{node.Relation}, path...)
					node = node.Parent
				}
				return path, nil
			}

			// Add to queue if not visited
			if !visited[neighborID] {
				visited[neighborID] = true
				queue = append(queue, &PathNode{
					EntityID: neighborID,
					Parent:   current,
					Relation: rel,
					Depth:    current.Depth + 1,
				})
			}
		}
	}

	// No path found
	return nil, fmt.Errorf("no path found between entities %d and %d", fromID, toID)
}
