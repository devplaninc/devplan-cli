package ide

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectIDEDirectories(t *testing.T) {
	// Test cases for different IDE directories
	testCases := []struct {
		dirs []string
		ides []Assistant
	}{
		{[]string{".cursor"}, []Assistant{CursorAI}},
		{[]string{".junie"}, []Assistant{JunieAI}},
		{[]string{".windsurf"}, []Assistant{WindsurfAI}},
		{[]string{".cursor", ".junie"}, []Assistant{CursorAI, JunieAI}},
		{nil, nil},
	}

	// Test each IDE directory
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%+v", tc.dirs), func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "ide-detect-test")
			require.NoError(t, err)
			defer func(path string) {
				require.NoError(t, os.RemoveAll(path))
			}(tempDir)
			for _, dir := range tc.dirs {
				idePath := filepath.Join(tempDir, dir)
				require.NoError(t, os.Mkdir(idePath, 0755))
				_, err := os.Stat(idePath)
				require.NoError(t, err)
			}
			ides, err := DetectAssistantOnPath(tempDir)
			require.NoError(t, err)
			assert.Equal(t, tc.ides, ides)
		})
	}
}
