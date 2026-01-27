package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// stubLLMClient is a simple stub implementation for development
type stubLLMClient struct{}

// NewStubLLMClient creates a new stub LLM client for development
func NewStubLLMClient() LLMClient {
	return &stubLLMClient{}
}

// GenerateCompletion generates a simple stub response for development
func (s *stubLLMClient) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	// Extract the user query from the prompt (it's the last line typically)
	lines := strings.Split(prompt, "\n")
	userQuery := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" && !strings.HasPrefix(lines[i], "You are") {
			userQuery = strings.TrimSpace(lines[i])
			break
		}
	}

	// Simple pattern matching for common query types
	lowerQuery := strings.ToLower(userQuery)

	// Entity lookup patterns
	if strings.Contains(lowerQuery, "who is") || strings.Contains(lowerQuery, "what is") || strings.Contains(lowerQuery, "tell me about") {
		// Extract entity name (simple heuristic)
		var entityName string
		if strings.Contains(lowerQuery, "who is ") {
			entityName = strings.TrimSpace(strings.Split(lowerQuery, "who is ")[1])
		} else if strings.Contains(lowerQuery, "what is ") {
			entityName = strings.TrimSpace(strings.Split(lowerQuery, "what is ")[1])
		} else if strings.Contains(lowerQuery, "tell me about ") {
			entityName = strings.TrimSpace(strings.Split(lowerQuery, "tell me about ")[1])
		}

		// Clean up common punctuation
		entityName = strings.Trim(entityName, "?.,;")

		response := map[string]interface{}{
			"action": "entity_lookup",
			"entity": entityName,
		}
		jsonResp, _ := json.Marshal(response)
		return string(jsonResp), nil
	}

	// Relationship patterns
	if strings.Contains(lowerQuery, "who did") && (strings.Contains(lowerQuery, "email") || strings.Contains(lowerQuery, "send")) {
		// Pattern: "who did X email" or "who did X send to"
		parts := strings.Split(lowerQuery, "who did ")
		if len(parts) > 1 {
			entityPart := strings.Split(parts[1], " email")[0]
			entityPart = strings.Split(entityPart, " send")[0]
			entityName := strings.TrimSpace(entityPart)

			response := map[string]interface{}{
				"action":   "relationship",
				"entity":   entityName,
				"rel_type": "SENT",
			}
			jsonResp, _ := json.Marshal(response)
			return string(jsonResp), nil
		}
	}

	if strings.Contains(lowerQuery, "who emailed") || strings.Contains(lowerQuery, "who sent") {
		// Pattern: "who emailed X" or "who sent to X"
		var entityName string
		if strings.Contains(lowerQuery, "who emailed ") {
			entityName = strings.TrimSpace(strings.Split(lowerQuery, "who emailed ")[1])
		} else if strings.Contains(lowerQuery, "who sent ") {
			parts := strings.Split(lowerQuery, "who sent ")
			if len(parts) > 1 {
				entityName = strings.TrimSpace(strings.Replace(parts[1], "to ", "", 1))
			}
		}

		entityName = strings.Trim(entityName, "?.,;")

		response := map[string]interface{}{
			"action":   "relationship",
			"entity":   entityName,
			"rel_type": "RECEIVED",
		}
		jsonResp, _ := json.Marshal(response)
		return string(jsonResp), nil
	}

	// Path finding patterns
	if strings.Contains(lowerQuery, "connection between") || strings.Contains(lowerQuery, "path between") {
		// Pattern: "connection between X and Y" or "path between X and Y"
		var source, target string

		if strings.Contains(lowerQuery, "between ") && strings.Contains(lowerQuery, " and ") {
			parts := strings.Split(lowerQuery, "between ")
			if len(parts) > 1 {
				betweenPart := parts[1]
				entities := strings.Split(betweenPart, " and ")
				if len(entities) >= 2 {
					source = strings.TrimSpace(entities[0])
					target = strings.TrimSpace(entities[1])
					target = strings.Trim(target, "?.,;")
				}
			}
		}

		if source != "" && target != "" {
			response := map[string]interface{}{
				"action": "path_finding",
				"source": source,
				"target": target,
			}
			jsonResp, _ := json.Marshal(response)
			return string(jsonResp), nil
		}
	}

	// Semantic search patterns
	if strings.Contains(lowerQuery, "search for") || strings.Contains(lowerQuery, "find") {
		searchText := userQuery
		if strings.Contains(lowerQuery, "search for ") {
			searchText = strings.TrimSpace(strings.Split(lowerQuery, "search for ")[1])
		} else if strings.Contains(lowerQuery, "find ") {
			searchText = strings.TrimSpace(strings.Split(lowerQuery, "find ")[1])
		}

		searchText = strings.Trim(searchText, "?.,;")

		response := map[string]interface{}{
			"action": "semantic_search",
			"text":   searchText,
		}
		jsonResp, _ := json.Marshal(response)
		return string(jsonResp), nil
	}

	// Aggregation patterns
	if strings.Contains(lowerQuery, "how many") {
		// Pattern: "how many emails did X send"
		var entityName string
		var relType string

		if strings.Contains(lowerQuery, " did ") {
			parts := strings.Split(lowerQuery, " did ")
			if len(parts) > 1 {
				entityPart := strings.Split(parts[1], " send")[0]
				entityPart = strings.Split(entityPart, " receive")[0]
				entityName = strings.TrimSpace(entityPart)

				if strings.Contains(lowerQuery, "send") {
					relType = "SENT"
				} else if strings.Contains(lowerQuery, "receive") {
					relType = "RECEIVED"
				}
			}
		}

		if entityName != "" && relType != "" {
			response := map[string]interface{}{
				"action":   "aggregation",
				"entity":   entityName,
				"rel_type": relType,
			}
			jsonResp, _ := json.Marshal(response)
			return string(jsonResp), nil
		}
	}

	// Default: treat as a direct question
	response := map[string]interface{}{
		"action": "answer",
		"answer": fmt.Sprintf("I understand you're asking: '%s'. This is a stub response. In production, a real LLM would process this query.", userQuery),
	}
	jsonResp, _ := json.Marshal(response)
	return string(jsonResp), nil
}

// GenerateEmbedding generates a stub embedding (zeros) for development
func (s *stubLLMClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Return a stub embedding of zeros with dimension 384 (common for sentence transformers)
	embedding := make([]float32, 384)
	return embedding, nil
}
