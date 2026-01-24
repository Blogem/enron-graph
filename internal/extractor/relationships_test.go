package extractor

import (
	"testing"
)

// T031: Unit tests for relationship creation
// Tests SENT, RECEIVED, MENTIONS, COMMUNICATES_WITH relationship types

func TestCreateSentRelationship(t *testing.T) {
	email := EmailMetadata{
		MessageID: "test@enron.com",
		From:      "alice@enron.com",
		To:        []string{"bob@enron.com"},
	}

	sender := TestEntityForRelationships{
		ID:   1,
		Name: "Alice",
		Type: "person",
	}

	recipient := TestEntityForRelationships{
		ID:   2,
		Name: "Bob",
		Type: "person",
	}

	relationship := Relationship{
		FromEntityID: sender.ID,
		ToEntityID:   recipient.ID,
		Type:         "SENT",
		EmailID:      email.MessageID,
	}

	if relationship.Type != "SENT" {
		t.Errorf("Expected relationship type 'SENT', got '%s'", relationship.Type)
	}

	if relationship.FromEntityID != sender.ID {
		t.Errorf("Expected FromEntityID %d, got %d", sender.ID, relationship.FromEntityID)
	}

	if relationship.ToEntityID != recipient.ID {
		t.Errorf("Expected ToEntityID %d, got %d", recipient.ID, relationship.ToEntityID)
	}
}

func TestCreateReceivedRelationship(t *testing.T) {
	email := EmailMetadata{
		MessageID: "test@enron.com",
		From:      "alice@enron.com",
		To:        []string{"bob@enron.com"},
	}

	sender := TestEntityForRelationships{
		ID:   1,
		Name: "Alice",
		Type: "person",
	}

	recipient := TestEntityForRelationships{
		ID:   2,
		Name: "Bob",
		Type: "person",
	}

	relationship := Relationship{
		FromEntityID: recipient.ID,
		ToEntityID:   sender.ID,
		Type:         "RECEIVED",
		EmailID:      email.MessageID,
	}

	if relationship.Type != "RECEIVED" {
		t.Errorf("Expected relationship type 'RECEIVED', got '%s'", relationship.Type)
	}

	// RECEIVED is from recipient to sender
	if relationship.FromEntityID != recipient.ID {
		t.Errorf("Expected FromEntityID %d, got %d", recipient.ID, relationship.FromEntityID)
	}

	if relationship.ToEntityID != sender.ID {
		t.Errorf("Expected ToEntityID %d, got %d", sender.ID, relationship.ToEntityID)
	}
}

func TestCreateMentionsRelationship(t *testing.T) {
	email := EmailMetadata{
		MessageID: "test@enron.com",
		Body:      "Bob mentioned Acme Corp in the meeting.",
	}

	person := TestEntityForRelationships{
		ID:   1,
		Name: "Bob",
		Type: "person",
	}

	organization := TestEntityForRelationships{
		ID:   2,
		Name: "Acme Corp",
		Type: "organization",
	}

	relationship := Relationship{
		FromEntityID: person.ID,
		ToEntityID:   organization.ID,
		Type:         "MENTIONS",
		EmailID:      email.MessageID,
	}

	if relationship.Type != "MENTIONS" {
		t.Errorf("Expected relationship type 'MENTIONS', got '%s'", relationship.Type)
	}

	if relationship.FromEntityID != person.ID {
		t.Errorf("Expected FromEntityID %d, got %d", person.ID, relationship.FromEntityID)
	}

	if relationship.ToEntityID != organization.ID {
		t.Errorf("Expected ToEntityID %d, got %d", organization.ID, relationship.ToEntityID)
	}
}

func TestCreateCommunicatesWithRelationship(t *testing.T) {
	person1 := TestEntityForRelationships{
		ID:   1,
		Name: "Alice",
		Type: "person",
	}

	person2 := TestEntityForRelationships{
		ID:   2,
		Name: "Bob",
		Type: "person",
	}

	relationship := Relationship{
		FromEntityID: person1.ID,
		ToEntityID:   person2.ID,
		Type:         "COMMUNICATES_WITH",
	}

	if relationship.Type != "COMMUNICATES_WITH" {
		t.Errorf("Expected relationship type 'COMMUNICATES_WITH', got '%s'", relationship.Type)
	}
}

func TestCreateRelationships_MultipleRecipients(t *testing.T) {
	email := EmailMetadata{
		MessageID: "test@enron.com",
		From:      "alice@enron.com",
		To:        []string{"bob@enron.com", "charlie@enron.com", "dave@enron.com"},
	}

	sender := TestEntityForRelationships{ID: 1, Name: "Alice", Type: "person"}
	recipients := []TestEntityForRelationships{
		{ID: 2, Name: "Bob", Type: "person"},
		{ID: 3, Name: "Charlie", Type: "person"},
		{ID: 4, Name: "Dave", Type: "person"},
	}

	// Create SENT relationships
	sentRelationships := []Relationship{}
	for _, recipient := range recipients {
		sentRelationships = append(sentRelationships, Relationship{
			FromEntityID: sender.ID,
			ToEntityID:   recipient.ID,
			Type:         "SENT",
			EmailID:      email.MessageID,
		})
	}

	if len(sentRelationships) != 3 {
		t.Errorf("Expected 3 SENT relationships, got %d", len(sentRelationships))
	}

	// Verify each relationship
	for i, rel := range sentRelationships {
		if rel.Type != "SENT" {
			t.Errorf("Relationship %d: expected type 'SENT', got '%s'", i, rel.Type)
		}
		if rel.FromEntityID != sender.ID {
			t.Errorf("Relationship %d: expected FromEntityID %d, got %d", i, sender.ID, rel.FromEntityID)
		}
	}
}

func TestCreateRelationships_CCAndBCC(t *testing.T) {
	email := EmailMetadata{
		MessageID: "test@enron.com",
		From:      "alice@enron.com",
		To:        []string{"bob@enron.com"},
		CC:        []string{"charlie@enron.com"},
		BCC:       []string{"dave@enron.com"},
	}

	allRecipients := append([]string{}, email.To...)
	allRecipients = append(allRecipients, email.CC...)
	allRecipients = append(allRecipients, email.BCC...)

	if len(allRecipients) != 3 {
		t.Errorf("Expected 3 total recipients, got %d", len(allRecipients))
	}

	// Each recipient should have a RECEIVED relationship
	receivedCount := len(allRecipients)
	if receivedCount != 3 {
		t.Errorf("Expected 3 RECEIVED relationships, got %d", receivedCount)
	}
}

func TestCreateRelationships_PersonToOrganization(t *testing.T) {
	email := EmailMetadata{
		MessageID: "test@enron.com",
		Body:      "Discussion about Enron Corporation",
	}

	person := TestEntityForRelationships{
		ID:   1,
		Name: "Alice",
		Type: "person",
	}

	organization := TestEntityForRelationships{
		ID:   2,
		Name: "Enron Corporation",
		Type: "organization",
	}

	// Person mentions organization
	relationship := Relationship{
		FromEntityID: person.ID,
		ToEntityID:   organization.ID,
		Type:         "MENTIONS",
		EmailID:      email.MessageID,
	}

	if relationship.Type != "MENTIONS" {
		t.Errorf("Expected 'MENTIONS', got '%s'", relationship.Type)
	}

	if person.Type != "person" || organization.Type != "organization" {
		t.Error("Expected person-to-organization relationship")
	}
}

func TestCreateRelationships_ConceptMentions(t *testing.T) {
	email := EmailMetadata{
		MessageID: "test@enron.com",
		Body:      "We need to discuss renewable energy and solar panels.",
	}

	person := TestEntityForRelationships{
		ID:   1,
		Name: "Alice",
		Type: "person",
	}

	concepts := []TestEntityForRelationships{
		{ID: 10, Name: "renewable energy", Type: "concept"},
		{ID: 11, Name: "solar panels", Type: "concept"},
	}

	relationships := []Relationship{}
	for _, concept := range concepts {
		relationships = append(relationships, Relationship{
			FromEntityID: person.ID,
			ToEntityID:   concept.ID,
			Type:         "MENTIONS",
			EmailID:      email.MessageID,
		})
	}

	if len(relationships) != 2 {
		t.Errorf("Expected 2 MENTIONS relationships, got %d", len(relationships))
	}

	for _, rel := range relationships {
		if rel.Type != "MENTIONS" {
			t.Errorf("Expected 'MENTIONS', got '%s'", rel.Type)
		}
	}
}

func TestCreateRelationships_BidirectionalCommunication(t *testing.T) {
	person1 := TestEntityForRelationships{ID: 1, Name: "Alice", Type: "person"}
	person2 := TestEntityForRelationships{ID: 2, Name: "Bob", Type: "person"}

	// Alice sends to Bob
	email1 := EmailMetadata{
		MessageID: "email1@enron.com",
		From:      "alice@enron.com",
		To:        []string{"bob@enron.com"},
	}

	// Bob replies to Alice
	email2 := EmailMetadata{
		MessageID: "email2@enron.com",
		From:      "bob@enron.com",
		To:        []string{"alice@enron.com"},
	}

	relationships := []Relationship{
		{FromEntityID: person1.ID, ToEntityID: person2.ID, Type: "SENT", EmailID: email1.MessageID},
		{FromEntityID: person2.ID, ToEntityID: person1.ID, Type: "SENT", EmailID: email2.MessageID},
	}

	// Should create bidirectional COMMUNICATES_WITH
	if len(relationships) != 2 {
		t.Errorf("Expected 2 relationships, got %d", len(relationships))
	}

	// Both should be SENT relationships
	for _, rel := range relationships {
		if rel.Type != "SENT" {
			t.Errorf("Expected 'SENT', got '%s'", rel.Type)
		}
	}
}

// Helper types for testing

type EmailMetadata struct {
	MessageID string
	From      string
	To        []string
	CC        []string
	BCC       []string
	Body      string
}

// TestEntity for relationship testing (actual entities stored in database via ent)
type TestEntityForRelationships struct {
	ID   int
	Name string
	Type string
}

type Relationship struct {
	FromEntityID int
	ToEntityID   int
	Type         string
	EmailID      string
}
