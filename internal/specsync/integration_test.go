package specsync

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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

// serverSpec represents a spec stored on the test server
type serverSpec struct {
	Name     string `json:"name"`
	Checksum string `json:"checksum"`
}

// testServer creates a mock server for artifact sync testing
type testServer struct {
	server      *httptest.Server
	specs       []serverSpec
	uploads     map[string]string // name -> content
	uploadCount atomic.Int32
	mu          gosync.Mutex
}

func newTestServer() *testServer {
	ts := &testServer{
		specs:   []serverSpec{},
		uploads: make(map[string]string),
	}

	mux := http.NewServeMux()

	// Specs endpoint
	mux.HandleFunc("/api/v1/company/", func(w http.ResponseWriter, r *http.Request) {
		// GET returns existing specs
		if r.Method == "GET" {
			ts.mu.Lock()
			response := struct {
				Specs []serverSpec `json:"specs"`
			}{Specs: ts.specs}
			ts.mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// POST uploads a spec
		if r.Method == "POST" {
			var req struct {
				Name     string `json:"name"`
				Content  string `json:"content"`
				Checksum string `json:"checksum"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			ts.mu.Lock()
			ts.uploads[req.Name] = req.Content
			ts.mu.Unlock()
			ts.uploadCount.Add(1)
			w.WriteHeader(http.StatusOK)
			return
		}

		http.NotFound(w, r)
	})

	ts.server = httptest.NewServer(mux)
	return ts
}

func (ts *testServer) close() {
	ts.server.Close()
}

// testClient implements Client using the test server
type testClient struct {
	baseURL string
}

func (c *testClient) GetTaskSpecs(companyID int32, taskID string) (*company.GetTaskSpecsResponse, error) {
	resp, err := http.Get(c.baseURL + "/api/v1/company/1/dev/task/task-123/specs")
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var result struct {
		Specs []serverSpec `json:"specs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to SpecDetails
	specs := make([]*artifacts.SpecDetails, len(result.Specs))
	for i, s := range result.Specs {
		spec := &artifacts.SpecDetails{}
		spec.SetName(s.Name)
		spec.SetChecksum(s.Checksum)
		specs[i] = spec
	}
	return company.GetTaskSpecsResponse_builder{Specs: specs}.Build(), nil
}

func (c *testClient) UploadTaskSpec(companyID int32, taskID string, req *company.UploadSpecRequest) error {
	body := struct {
		Name     string `json:"name"`
		Content  string `json:"content"`
		Checksum string `json:"checksum"`
	}{
		Name:     req.GetName(),
		Content:  req.GetContent(),
		Checksum: req.GetChecksum(),
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/company/1/dev/task/task-123/specs", bytes.NewReader(data))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}
	return nil
}

func TestIntegration_FullSyncFlow(t *testing.T) {
	ts := newTestServer()
	defer ts.close()

	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	err := os.MkdirAll(specsDir, 0755)
	require.NoError(t, err)

	// Create test artifacts
	file1 := filepath.Join(specsDir, "prd.md")
	content1 := []byte("# PRD\nProduct requirements")
	err = os.WriteFile(file1, content1, 0644)
	require.NoError(t, err)

	file2 := filepath.Join(specsDir, "plan.md")
	content2 := []byte("# Plan\nTask plan")
	err = os.WriteFile(file2, content2, 0644)
	require.NoError(t, err)

	file3 := filepath.Join(specsDir, "code.md")
	content3 := []byte("# Code\nTask code")
	err = os.WriteFile(file3, content3, 0644)
	require.NoError(t, err)

	// Set up server with one existing spec (to test skip behavior)
	checksum := CalculateChecksumBytes(content2)
	ts.specs = append(ts.specs, serverSpec{Name: "plan.md", Checksum: checksum})

	// Create client and syncer
	client := &testClient{baseURL: ts.server.URL}
	syncer := NewSyncer(client, 1, "task-123", specsDir, time.Second)

	// Run sync
	result := syncer.TriggerOnce(context.Background())

	// Verify results
	assert.Equal(t, 1, result.Uploaded) // code.md should be uploaded
	assert.Equal(t, 1, result.Skipped)  // plan.md should be skipped, prd.md should be ignored as input spec
	assert.Equal(t, 0, result.Failed)

	// Verify upload was sent
	assert.Equal(t, int32(1), ts.uploadCount.Load())
}

func TestIntegration_BackgroundSyncLifecycle(t *testing.T) {
	// Create test server
	ts := newTestServer()
	defer ts.close()

	// Create test specs directory
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	err := os.MkdirAll(specsDir, 0755)
	require.NoError(t, err)

	// Create test artifact
	testFile := filepath.Join(specsDir, "test.md")
	err = os.WriteFile(testFile, []byte("# Test"), 0644)
	require.NoError(t, err)

	// Create client and syncer with short interval
	client := &testClient{baseURL: ts.server.URL}
	syncer := NewSyncer(client, 1, "task-123", specsDir, 50*time.Millisecond)

	// Start background sync
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		syncer.RunBackground(ctx)
		close(done)
	}()

	// Wait for a few sync cycles
	time.Sleep(200 * time.Millisecond)

	// Cancel context to stop background sync
	cancel()

	// Wait for goroutine to finish
	select {
	case <-done:
		// Background sync stopped cleanly
	case <-time.After(time.Second):
		t.Fatal("Background sync did not stop within timeout")
	}

	// Verify some uploads occurred
	assert.Greater(t, ts.uploadCount.Load(), int32(0))

	// Store upload count before stopping
	countBeforeStop := ts.uploadCount.Load()

	// Wait and verify no more uploads occur
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, countBeforeStop, ts.uploadCount.Load())
}

func TestIntegration_ConcurrentMCPAndBackground(t *testing.T) {
	// Create test server
	ts := newTestServer()
	defer ts.close()

	// Create test specs directory
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	err := os.MkdirAll(specsDir, 0755)
	require.NoError(t, err)

	// Create test artifact
	testFile := filepath.Join(specsDir, "test.md")
	err = os.WriteFile(testFile, []byte("# Test"), 0644)
	require.NoError(t, err)

	// Create client and syncer
	client := &testClient{baseURL: ts.server.URL}
	syncer := NewSyncer(client, 1, "task-123", specsDir, 50*time.Millisecond)

	// Start background sync
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go syncer.RunBackground(ctx)

	// Simulate concurrent MCP triggers
	var wg gosync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			syncer.TriggerOnce(ctx)
		}()
		time.Sleep(10 * time.Millisecond)
	}
	wg.Wait()

	// Verify no errors and all syncs completed
	// Due to single-flight guard, not all triggers may result in uploads
	assert.Greater(t, ts.uploadCount.Load(), int32(0))
}

func TestIntegration_DiscoveryWithNestedDirs(t *testing.T) {
	t.Parallel()
	// Create nested specs directory structure
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	storyDir := filepath.Join(specsDir, "STORY-123")
	taskDir := filepath.Join(storyDir, "TASK-456")

	err := os.MkdirAll(taskDir, 0755)
	require.NoError(t, err)

	// Create artifacts at different levels
	err = os.WriteFile(filepath.Join(specsDir, "prd.md"), []byte("# PRD"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(storyDir, "requirements.md"), []byte("# Story"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(taskDir, "requirements.md"), []byte("# Task"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(taskDir, "plan.md"), []byte("# Task"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(taskDir, "coding.md"), []byte("# Task"), 0644)
	require.NoError(t, err)

	// Discover artifacts
	specs, err := DiscoverTaskSpecs(taskDir)
	require.NoError(t, err)

	// Verify all artifacts found with correct names
	names := make(map[string]bool)
	for _, s := range specs {
		names[s.Name] = true
	}
	assert.Len(t, names, 2)
	assert.Contains(t, names, "plan.md")
	assert.Contains(t, names, "coding.md")
}

func TestIntegration_EmptySpecsDir(t *testing.T) {
	// Create empty specs directory
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	err := os.MkdirAll(specsDir, 0755)
	require.NoError(t, err)

	// Create mock client
	client := &mockClient{
		specs: []*artifacts.SpecDetails{},
	}

	// Create syncer
	syncer := NewSyncer(client, 1, "task-123", specsDir, time.Second)

	// Trigger sync
	result := syncer.TriggerOnce(context.Background())

	// Verify no errors with empty directory
	assert.Equal(t, 0, result.Uploaded)
	assert.Equal(t, 0, result.Skipped)
	assert.Equal(t, 0, result.Failed)
}

// makeSpecDetails is defined in sync_test.go and shared across test files
