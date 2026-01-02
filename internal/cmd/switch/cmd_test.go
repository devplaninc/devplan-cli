package switch_cmd

import (
	"testing"

	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetIDENames(t *testing.T) {
	// Mock installed IDEs
	installedIDEs := map[ide.IDE]string{
		ide.PyCharm:  "/path/to/pycharm",
		ide.WebStorm: "/path/to/webstorm",
		ide.IntelliJ: "/path/to/intellij",
		ide.Cursor:   "/path/to/cursor",
		ide.Claude:   "/path/to/claude",
		"other":      "/path/to/other",
		"another":    "/path/to/another",
	}

	tests := []struct {
		name     string
		lastIDE  string
		expected []ide.IDE
	}{
		{
			name:    "No last IDE",
			lastIDE: "",
			expected: []ide.IDE{
				ide.Claude,
				ide.Cursor,
				ide.WebStorm,
				ide.PyCharm,
				"another",
				ide.IntelliJ,
				"other",
			},
		},
		{
			name:    "Last IDE is PyCharm",
			lastIDE: string(ide.PyCharm),
			expected: []ide.IDE{
				ide.PyCharm,
				ide.Claude,
				ide.Cursor,
				ide.WebStorm,
				"another",
				ide.IntelliJ,
				"other",
			},
		},
		{
			name:    "Last IDE is other",
			lastIDE: "other",
			expected: []ide.IDE{
				"other",
				ide.Claude,
				ide.Cursor,
				ide.WebStorm,
				ide.PyCharm,
				"another",
				ide.IntelliJ,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set(prefs.LastIDEKey, tt.lastIDE)
			actual := getIDENames(installedIDEs)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
