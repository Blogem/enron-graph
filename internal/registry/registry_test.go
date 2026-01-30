package registry

import (
	"context"
	"testing"
)

// Test that Register() function stores EntityCreator in map
func TestRegister(t *testing.T) {
	// Clear the registry before test
	PromotedTypes = make(map[string]EntityCreator)

	mockCreator := func(ctx context.Context, data map[string]any) (any, error) {
		return "mock entity", nil
	}

	Register("TestType", mockCreator)

	if _, exists := PromotedTypes["TestType"]; !exists {
		t.Error("Expected TestType to be registered in PromotedTypes map")
	}
}

// Test looking up existing promoted type
func TestLookupExistingType(t *testing.T) {
	// Clear and setup registry
	PromotedTypes = make(map[string]EntityCreator)

	expectedResult := "test entity"
	mockCreator := func(ctx context.Context, data map[string]any) (any, error) {
		return expectedResult, nil
	}

	Register("Person", mockCreator)

	// Lookup the registered type
	creator, exists := PromotedTypes["Person"]
	if !exists {
		t.Fatal("Expected Person to be registered")
	}

	// Call the creator to verify it works
	result, err := creator(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != expectedResult {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}
}

// Test looking up non-existent type (returns nil/false)
func TestLookupNonExistentType(t *testing.T) {
	// Clear registry
	PromotedTypes = make(map[string]EntityCreator)

	_, exists := PromotedTypes["NonExistent"]
	if exists {
		t.Error("Expected NonExistent type to not be in registry")
	}
}

// Test multiple registrations (no collision)
func TestMultipleRegistrations(t *testing.T) {
	// Clear registry
	PromotedTypes = make(map[string]EntityCreator)

	creator1 := func(ctx context.Context, data map[string]any) (any, error) {
		return "entity1", nil
	}
	creator2 := func(ctx context.Context, data map[string]any) (any, error) {
		return "entity2", nil
	}
	creator3 := func(ctx context.Context, data map[string]any) (any, error) {
		return "entity3", nil
	}

	Register("Type1", creator1)
	Register("Type2", creator2)
	Register("Type3", creator3)

	if len(PromotedTypes) != 3 {
		t.Errorf("Expected 3 registered types, got %d", len(PromotedTypes))
	}

	// Verify each type is correctly registered
	types := []string{"Type1", "Type2", "Type3"}
	for _, typeName := range types {
		if _, exists := PromotedTypes[typeName]; !exists {
			t.Errorf("Expected %s to be registered", typeName)
		}
	}
}

// Test that RegisterFinder() function stores EntityFinder in map
func TestRegisterFinder(t *testing.T) {
	// Clear the registry before test
	PromotedFinders = make(map[string]EntityFinder)

	mockFinder := func(ctx context.Context, uniqueID string) (any, error) {
		return "mock entity", nil
	}

	RegisterFinder("TestType", mockFinder)

	if _, exists := PromotedFinders["TestType"]; !exists {
		t.Error("Expected TestType to be registered in PromotedFinders map")
	}
}

// Test looking up existing finder
func TestLookupExistingFinder(t *testing.T) {
	// Clear and setup registry
	PromotedFinders = make(map[string]EntityFinder)

	expectedResult := "test entity"
	mockFinder := func(ctx context.Context, uniqueID string) (any, error) {
		return expectedResult, nil
	}

	RegisterFinder("Person", mockFinder)

	// Lookup the registered finder
	finder, exists := PromotedFinders["Person"]
	if !exists {
		t.Fatal("Expected Person finder to be registered")
	}

	// Call the finder to verify it works
	result, err := finder(context.Background(), "test-unique-id")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != expectedResult {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}
}

// Test looking up non-existent finder (returns nil/false)
func TestLookupNonExistentFinder(t *testing.T) {
	// Clear registry
	PromotedFinders = make(map[string]EntityFinder)

	_, exists := PromotedFinders["NonExistent"]
	if exists {
		t.Error("Expected NonExistent finder to not be in registry")
	}
}

// Test multiple finder registrations
func TestMultipleFinderRegistrations(t *testing.T) {
	// Clear registry
	PromotedFinders = make(map[string]EntityFinder)

	finder1 := func(ctx context.Context, uniqueID string) (any, error) {
		return "entity1", nil
	}
	finder2 := func(ctx context.Context, uniqueID string) (any, error) {
		return "entity2", nil
	}
	finder3 := func(ctx context.Context, uniqueID string) (any, error) {
		return "entity3", nil
	}

	RegisterFinder("Type1", finder1)
	RegisterFinder("Type2", finder2)
	RegisterFinder("Type3", finder3)

	if len(PromotedFinders) != 3 {
		t.Errorf("Expected 3 registered finders, got %d", len(PromotedFinders))
	}

	// Verify each finder is correctly registered
	types := []string{"Type1", "Type2", "Type3"}
	for _, typeName := range types {
		if _, exists := PromotedFinders[typeName]; !exists {
			t.Errorf("Expected %s finder to be registered", typeName)
		}
	}
}
