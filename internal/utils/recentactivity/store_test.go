package recentactivity

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/devplaninc/devplan-cli/internal/pb/config"
	"github.com/devplaninc/devplan-cli/internal/utils/metadata"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStoreUpdateAndLoad(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	store := newStore(tempDir)

	now := time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC)
	require.NoError(t, store.updateTask("task-123", now, "unit"))

	cfg, err := store.load()
	require.NoError(t, err)
	require.Len(t, cfg.GetEntries(), 1)
	assert.Equal(t, "task-123", cfg.GetEntries()[0].GetTaskId())
	assert.Equal(t, "unit", cfg.GetEntries()[0].GetSource())

	data, err := os.ReadFile(filepath.Join(tempDir, configFileName))
	require.NoError(t, err)

	var decoded config.RecentActivityConfig
	u := protojson.UnmarshalOptions{DiscardUnknown: true}
	require.NoError(t, u.Unmarshal(data, &decoded))
	require.Len(t, decoded.GetEntries(), 1)
	assert.Equal(t, "task-123", decoded.GetEntries()[0].GetTaskId())
}

func TestDeduplicationAndSorting(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	store := newStore(tempDir)

	base := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, store.updateTask("task-a", base, ""))
	require.NoError(t, store.updateTask("task-b", base.Add(2*time.Hour), ""))
	require.NoError(t, store.updateTask("task-a", base.Add(4*time.Hour), "update"))

	cfg, err := store.load()
	require.NoError(t, err)
	require.Len(t, cfg.GetEntries(), 2)
	assert.Equal(t, "task-a", cfg.GetEntries()[0].GetTaskId())
	assert.Equal(t, "task-b", cfg.GetEntries()[1].GetTaskId())
	assert.Equal(t, "update", cfg.GetEntries()[0].GetSource())
}

func TestPruneMaxEntries(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	store := newStore(tempDir)

	base := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	total := defaultMaxEntries + 5
	for i := 0; i < total; i++ {
		taskID := fmt.Sprintf("task-%02d", i)
		ts := base.Add(time.Duration(i) * time.Minute)
		require.NoError(t, store.updateTask(taskID, ts, ""))
	}

	cfg, err := store.load()
	require.NoError(t, err)
	require.Len(t, cfg.GetEntries(), defaultMaxEntries)

	found := make(map[string]bool, len(cfg.GetEntries()))
	for _, entry := range cfg.GetEntries() {
		found[entry.GetTaskId()] = true
	}

	for i := 0; i < 5; i++ {
		assert.False(t, found[fmt.Sprintf("task-%02d", i)])
	}
	for i := 5; i < total; i++ {
		assert.True(t, found[fmt.Sprintf("task-%02d", i)])
	}
}

func TestConcurrentUpdates(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	store := newStore(tempDir)

	taskCount := 10
	updatesPerTask := 4
	errCh := make(chan error, taskCount*updatesPerTask)
	var wg sync.WaitGroup

	for i := 0; i < taskCount; i++ {
		taskID := fmt.Sprintf("task-%d", i)
		for j := 0; j < updatesPerTask; j++ {
			wg.Add(1)
			go func(id string, offset int) {
				defer wg.Done()
				ts := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(offset) * time.Second)
				errCh <- store.updateTask(id, ts, "concurrent")
			}(taskID, j)
		}
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	cfg, err := store.load()
	require.NoError(t, err)
	require.NotEmpty(t, cfg.GetEntries())
	require.LessOrEqual(t, len(cfg.GetEntries()), defaultMaxEntries)

	seen := make(map[string]bool)
	for _, entry := range cfg.GetEntries() {
		require.NotEmpty(t, entry.GetTaskId())
		require.False(t, seen[entry.GetTaskId()])
		seen[entry.GetTaskId()] = true
	}
}

func TestTaskActivityIndexFiltersNonTaskEntries(t *testing.T) {
	cfg := config.RecentActivityConfig_builder{
		Entries: []*config.RecentActivityEntry{
			config.RecentActivityEntry_builder{
				Type:         config.RecentActivityEntryType_ENTRY_TYPE_TASK,
				TaskId:       "task-1",
				LastActivity: timestamppb.New(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)),
			}.Build(),
			config.RecentActivityEntry_builder{
				Type:         config.RecentActivityEntryType_ENTRY_TYPE_PROJECT,
				TaskId:       "proj-ignored",
				LastActivity: timestamppb.New(time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)),
			}.Build(),
		},
	}.Build()

	index := taskActivityIndex(cfg)
	require.Len(t, index, 1)
	_, ok := index["task-1"]
	assert.True(t, ok)
}

func TestSortFeaturesByActivity(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	featureAPath := filepath.Join(tempDir, "feature-a")
	featureBPath := filepath.Join(tempDir, "feature-b")
	featureNoMetaPath := filepath.Join(tempDir, "feature-nometa")

	require.NoError(t, os.MkdirAll(featureAPath, 0755))
	require.NoError(t, os.MkdirAll(featureBPath, 0755))
	require.NoError(t, os.MkdirAll(featureNoMetaPath, 0755))

	require.NoError(t, metadata.EnsureMetadataSetup(featureAPath, metadata.Metadata{TaskID: "task-a"}))
	require.NoError(t, metadata.EnsureMetadataSetup(featureBPath, metadata.Metadata{TaskID: "task-b"}))

	features := []workspace.ClonedFeature{
		{DirName: "a", FullPath: featureAPath},
		{DirName: "b", FullPath: featureBPath},
		{DirName: "c", FullPath: featureNoMetaPath},
	}

	activity := map[string]time.Time{
		"task-b": time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC).Add(2 * time.Hour),
		"task-a": time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC).Add(1 * time.Hour),
	}

	sorted := sortFeaturesByActivity(features, activity)
	require.Len(t, sorted, 3)
	assert.Equal(t, featureBPath, sorted[0].FullPath)
	assert.Equal(t, featureAPath, sorted[1].FullPath)
	assert.Equal(t, featureNoMetaPath, sorted[2].FullPath)
}
