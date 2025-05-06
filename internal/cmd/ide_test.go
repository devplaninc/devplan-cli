package cmd

import (
	"testing"

	"github.com/spf13/viper"
)

func TestGetSetDefaultIDE(t *testing.T) {
	// Setup
	viper.Reset()
	viper.Set("default_ide", "")

	// Test getDefaultIDE with no default set
	ide := getDefaultIDE()
	if ide != "" {
		t.Errorf("Expected empty default IDE, got %s", ide)
	}

	// Test setting and getting default IDE in memory
	testIDE := "cursor"
	viper.Set("default_ide", testIDE)

	// Test getDefaultIDE with default set
	ide = getDefaultIDE()
	if ide != testIDE {
		t.Errorf("Expected default IDE %s, got %s", testIDE, ide)
	}

	// Test setting a different IDE
	testIDE = "intellij"
	viper.Set("default_ide", testIDE)

	// Verify the new IDE is set
	ide = getDefaultIDE()
	if ide != testIDE {
		t.Errorf("Expected default IDE %s, got %s", testIDE, ide)
	}
}

func TestIDESelectionModel(t *testing.T) {
	// Test InitIDESelectionModel
	model := InitIDESelectionModel()

	// Verify initial state
	if len(model.choices) != 3 {
		t.Errorf("Expected 3 choices, got %d", len(model.choices))
	}

	if model.cursor != 0 {
		t.Errorf("Expected cursor at position 0, got %d", model.cursor)
	}

	if model.selected != "" {
		t.Errorf("Expected empty selection, got %s", model.selected)
	}

	if model.quitting {
		t.Errorf("Expected quitting to be false")
	}

	// Test View method
	view := model.View()
	if view == "" {
		t.Errorf("Expected non-empty view")
	}

	// Verify the view contains all IDE types
	for _, ide := range []string{IDETypeCursor, IDETypeIntelliJ, IDETypeJunie} {
		if !contains(view, ide) {
			t.Errorf("View does not contain IDE type %s", ide)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
