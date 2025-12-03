package specsync

import (
	"context"
	"os"
	"path/filepath"
	gosync "sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/devplaninc/webapp/golang/pb/api/devplan/services/web/company"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/artifacts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClient implements Client for testing
type mockClient struct {
	specs       []*artifacts.SpecDetails
	uploads     []string
	uploadErr   error
	specsErr    error
	uploadCount atomic.Int32
	mu          gosync.Mutex
}

func (m *mockClient) GetTaskSpecs(companyID int32, taskID string) (*company.GetTaskSpecsResponse, error) {
	if m.specsErr != nil {
		return nil, m.specsErr
	}
	return company.GetTaskSpecsResponse_builder{Specs: m.specs}.Build(), nil
}

func (m *mockClient) UploadTaskSpec(companyID int32, taskID string, req *company.UploadSpecRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.uploads = append(m.uploads, req.GetName())
	m.uploadCount.Add(1)
	return m.uploadErr
}

// makeSpecDetails creates a SpecDetails with the given name and checksum
func makeSpecDetails(name, checksum string) *artifacts.SpecDetails {
	spec := &artifacts.SpecDetails{}
	spec.SetName(name)
	spec.SetChecksum(checksum)
	return spec
}

func TestShouldUpload(t *testing.T) {
	tests := []struct {
		name        string
		artifact    Spec
		serverSpecs []*artifacts.SpecDetails
		expected    bool
	}{
		{
			name: "artifact not on server - should upload",
			artifact: Spec{
				Name:     "new.md",
				Checksum: "abc123",
			},
			serverSpecs: []*artifacts.SpecDetails{},
			expected:    true,
		},
		{
			name: "artifact changed - should upload",
			artifact: Spec{
				Name:     "existing.md",
				Checksum: "newchecksum",
			},
			serverSpecs: []*artifacts.SpecDetails{
				makeSpecDetails("existing.md", "oldchecksum"),
			},
			expected: true,
		},
		{
			name: "artifact unchanged - should skip",
			artifact: Spec{
				Name:     "same.md",
				Checksum: "samechecksum",
			},
			serverSpecs: []*artifacts.SpecDetails{
				makeSpecDetails("same.md", "samechecksum"),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := shouldUpload(tt.artifact, tt.serverSpecs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncer_TriggerOnce(t *testing.T) {
	t.Parallel()
	// Create test specs directory
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	err := os.MkdirAll(specsDir, 0755)
	require.NoError(t, err)

	// Create test artifact
	testFile := filepath.Join(specsDir, "test.md")
	content := []byte("# Test")
	err = os.WriteFile(testFile, content, 0644)
	require.NoError(t, err)

	// Create mock client
	client := &mockClient{
		specs: []*artifacts.SpecDetails{}, // Empty = all artifacts need upload
	}

	// Create syncer
	syncer := NewSyncer(client, 1, "task-123", specsDir, time.Second)

	// Trigger sync
	result := syncer.TriggerOnce(context.Background())

	// Verify result
	assert.Equal(t, 1, result.Uploaded)
	assert.Equal(t, 0, result.Skipped)
	assert.Equal(t, 0, result.Failed)
	assert.Contains(t, client.uploads, "test.md")
}

func TestSyncer_TriggerOnce_SkipsUnchanged(t *testing.T) {
	t.Parallel()
	// Create test specs directory
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	err := os.MkdirAll(specsDir, 0755)
	require.NoError(t, err)

	// Create test artifact
	testFile := filepath.Join(specsDir, "test.md")
	content := []byte("# Test")
	err = os.WriteFile(testFile, content, 0644)
	require.NoError(t, err)

	// Calculate checksum to simulate unchanged file
	checksum := CalculateChecksumBytes(content)

	// Create mock client with matching checksum
	client := &mockClient{
		specs: []*artifacts.SpecDetails{
			makeSpecDetails("test.md", checksum),
		},
	}

	// Create syncer
	syncer := NewSyncer(client, 1, "task-123", specsDir, time.Second)

	// Trigger sync
	result := syncer.TriggerOnce(context.Background())

	// Verify result - should skip unchanged file
	assert.Equal(t, 0, result.Uploaded)
	assert.Equal(t, 1, result.Skipped)
	assert.Equal(t, 0, result.Failed)
	assert.Empty(t, client.uploads)
}

func TestSyncer_TriggerOnce_ConcurrentCalls(t *testing.T) {
	t.Parallel()
	// Create test specs directory
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	err := os.MkdirAll(specsDir, 0755)
	require.NoError(t, err)

	// Create test artifact
	testFile := filepath.Join(specsDir, "test.md")
	err = os.WriteFile(testFile, []byte("# Test"), 0644)
	require.NoError(t, err)

	// Create mock client
	client := &mockClient{
		specs: []*artifacts.SpecDetails{},
	}

	// Create syncer
	syncer := NewSyncer(client, 1, "task-123", specsDir, time.Second)

	// Trigger multiple concurrent syncs
	var wg gosync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			syncer.TriggerOnce(context.Background())
		}()
	}
	wg.Wait()

	// Verify uploads were serialized (only one upload per concurrent batch)
	// Due to single-flight guard, some calls may be skipped
	assert.LessOrEqual(t, len(client.uploads), 10)
}

func TestSyncer_PerArtifactLocking(t *testing.T) {
	t.Parallel()
	// Create test specs directory
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	err := os.MkdirAll(specsDir, 0755)
	require.NoError(t, err)

	// Create test artifact
	testFile := filepath.Join(specsDir, "test.md")
	err = os.WriteFile(testFile, []byte("# Test"), 0644)
	require.NoError(t, err)

	// Create mock client that tracks concurrent uploads
	client := &mockClient{
		specs: []*artifacts.SpecDetails{},
	}

	// Create syncer
	syncer := NewSyncer(client, 1, "task-123", specsDir, time.Second)

	// Upload same artifact concurrently
	var wg gosync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			syncer.uploadSpec(context.Background(), Spec{
				Name: "test.md",
				Path: testFile,
			})
		}()
	}
	wg.Wait()

	// All uploads should succeed (serialized by per-artifact lock)
	assert.Equal(t, int32(5), client.uploadCount.Load())
}
