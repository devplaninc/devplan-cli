package ide

import (
	"testing"

	"github.com/devplaninc/devplan-cli/internal/utils/prompts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCurrentFeatureRules_NoTruncation(t *testing.T) {
	// Create a document with content under the limit
	shortContent := "Short content"
	docID := "test-doc-id"
	docTitle := "Test Document"

	doc := documents.DocumentEntity_builder{
		Id:      docID,
		Title:   docTitle,
		Content: &shortContent,
	}.Build()

	// Create base rule
	baseRule := Rule{
		Header:   "Base Header",
		Footer:   "Base Footer",
		NoPrefix: false,
	}

	paths := rulePaths{
		dir: "test-dir",
		ext: "md",
	}

	// Test rule generation
	rules, err := generateCurrentFeatureRules(paths, baseRule, &prompts.Info{Doc: doc})

	// Assertions
	require.NoError(t, err)
	require.Len(t, rules, 1)

	rule := rules[0]
	assert.Equal(t, ruleFileNameCurrentFeaturePrefix, rule.Name)
	assert.Equal(t, shortContent, rule.Content)
	assert.Equal(t, "Base Header", rule.Header) // Should not have truncation header
	assert.Equal(t, "Base Footer", rule.Footer)
}
