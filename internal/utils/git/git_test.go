package git

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"github.com/stretchr/testify/assert"
)

func TestGitCommandVerbose(t *testing.T) {
	// Keep original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	tests := []struct {
		name     string
		verbose  bool
		expected bool
	}{
		{
			name:     "Verbose false - no output",
			verbose:  false,
			expected: false,
		},
		{
			name:     "Verbose true - has output",
			verbose:  true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldVerbose := prefs.Verbose
			prefs.Verbose = tt.verbose
			defer func() { prefs.Verbose = oldVerbose }()

			r, w, _ := os.Pipe()
			os.Stdout = w

			gitCommand("status")

			w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			if tt.expected {
				assert.Contains(t, output, "> git status")
			} else {
				assert.NotContains(t, output, "> git status")
			}
		})
	}
}

func TestGetFullName(t *testing.T) {
	tests := []struct {
		url      string
		expected string
		wantErr  bool
	}{
		// HTTPS URLs
		{
			url:      "https://github.com/devplaninc/devplan-cli.git",
			expected: "devplaninc/devplan-cli",
		},
		{
			url:      "https://github.com/devplaninc/devplan-cli",
			expected: "devplaninc/devplan-cli",
		},
		{
			url:      "https://gitlab.com/devplaninc/devplan-cli.git",
			expected: "devplaninc/devplan-cli",
		},
		{
			url:      "https://bitbucket.org/devplaninc/devplan-cli.git",
			expected: "devplaninc/devplan-cli",
		},

		{
			url:      "git@github.com:devplaninc/devplan-cli.git",
			expected: "devplaninc/devplan-cli",
		},
		{
			url:      "git@gitlab.com:devplaninc/devplan-cli.git",
			expected: "devplaninc/devplan-cli",
		},
		{
			url:      "git@bitbucket.org:devplaninc/devplan-cli.git",
			expected: "devplaninc/devplan-cli",
		},

		{
			url:      "https://github.com/devplaninc/subgroup/devplan-cli.git",
			expected: "devplaninc/subgroup/devplan-cli",
		},
		{
			url:      "https://gitlab.com/devplaninc/group/subgroup/devplan-cli.git",
			expected: "devplaninc/group/subgroup/devplan-cli",
		},

		{
			url:      "https://git.example.com/devplaninc/devplan-cli.git",
			expected: "devplaninc/devplan-cli",
		},
		{
			url:      "git@git.example.com:devplaninc/devplan-cli.git",
			expected: "devplaninc/devplan-cli",
		},

		{
			url:      "https://git.example.com:8080/devplaninc/devplan-cli.git",
			expected: "devplaninc/devplan-cli",
		},
		{
			url:      "https://user:password@github.com/devplaninc/devplan-cli.git",
			expected: "devplaninc/devplan-cli",
		},
		{
			url:     "not-a-valid-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got, err := GetFullName(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, "", got)
			} else {
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
