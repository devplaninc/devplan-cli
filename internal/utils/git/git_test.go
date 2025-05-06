package git

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
