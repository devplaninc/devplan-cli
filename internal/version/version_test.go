package version

import (
	"testing"
)

func TestIsAutoUpdateDisabled(t *testing.T) {
	// Test cases for auto-update disabled check
	testCases := []struct {
		name             string
		flagValue        string
		expectedDisabled bool
	}{
		{
			name:             "Auto-update enabled (default)",
			flagValue:        "false",
			expectedDisabled: false,
		},
		{
			name:             "Auto-update disabled",
			flagValue:        "true",
			expectedDisabled: true,
		},
		{
			name:             "Auto-update enabled (empty string)",
			flagValue:        "",
			expectedDisabled: false,
		},
		{
			name:             "Auto-update enabled (invalid value)",
			flagValue:        "invalid",
			expectedDisabled: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Store original value
			originalValue := DisableAutoUpdate

			// Set test value
			DisableAutoUpdate = tc.flagValue

			// Test the function
			result := IsAutoUpdateDisabled()

			// Verify result
			if result != tc.expectedDisabled {
				t.Errorf("Expected IsAutoUpdateDisabled() = %v, got %v", tc.expectedDisabled, result)
			}

			// Restore original value
			DisableAutoUpdate = originalValue
		})
	}
}

func TestVersionFunctions(t *testing.T) {
	// Test that version functions return expected values
	version := GetVersion()
	if version == "" {
		t.Error("GetVersion() returned empty string")
	}

	commitHash := GetCommitHash()
	if commitHash == "" {
		t.Error("GetCommitHash() returned empty string")
	}

	buildDate := GetBuildDate()
	if buildDate == "" {
		t.Error("GetBuildDate() returned empty string")
	}
}
