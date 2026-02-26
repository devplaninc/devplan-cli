package converters

import (
	"testing"

	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFeatureDetails_WithRepoNames(t *testing.T) {
	doc := &documents.DocumentEntity{}
	doc.SetDetails(`{"repoNames": ["org/repo1", "org/repo2"]}`)

	details, err := GetFeatureDetails(doc)
	require.NoError(t, err)
	require.NotNil(t, details)
	assert.Equal(t, []string{"org/repo1", "org/repo2"}, details.GetRepoNames())
}

func TestGetFeatureDetails_EmptyDetails(t *testing.T) {
	doc := &documents.DocumentEntity{}
	doc.SetDetails("")

	details, err := GetFeatureDetails(doc)
	require.NoError(t, err)
	assert.Nil(t, details)
}

func TestGetFeatureDetails_NoRepoNames(t *testing.T) {
	doc := &documents.DocumentEntity{}
	doc.SetDetails(`{}`)

	details, err := GetFeatureDetails(doc)
	require.NoError(t, err)
	require.NotNil(t, details)
	assert.Empty(t, details.GetRepoNames())
}

func TestGetTaskDetails_WithRepoName(t *testing.T) {
	doc := &documents.DocumentEntity{}
	doc.SetDetails(`{"repoName": "org/repo"}`)

	details, err := GetTaskDetails(doc)
	require.NoError(t, err)
	require.NotNil(t, details)
	assert.Equal(t, "org/repo", details.GetRepoName())
}

func TestGetTaskDetails_EmptyDetails(t *testing.T) {
	doc := &documents.DocumentEntity{}
	doc.SetDetails("")

	details, err := GetTaskDetails(doc)
	require.NoError(t, err)
	assert.Nil(t, details)
}
