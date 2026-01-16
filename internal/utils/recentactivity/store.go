package recentactivity

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/devplaninc/devplan-cli/internal/pb/config"
	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	configFileName    = "recent-activity.json"
	defaultMaxEntries = 20
	defaultFileMode   = 0644
)

type store struct {
	path       string
	maxEntries int
	logger     *slog.Logger
}

func newStore(configDir string) *store {
	return &store{
		path:       filepath.Join(configDir, configFileName),
		maxEntries: defaultMaxEntries,
		logger:     slog.Default(),
	}
}

func newDefaultStore() (*store, error) {
	dir, err := prefs.GetConfigDir()
	if err != nil {
		return nil, err
	}
	return newStore(dir), nil
}

func recordTaskActivity(taskID string, ts time.Time, source string) error {
	store, err := newDefaultStore()
	if err != nil {
		return err
	}
	return store.updateTask(taskID, ts, source)
}

func RecordTaskActivity(taskID string, source string) error {
	return recordTaskActivity(taskID, time.Now(), source)
}

func (s *store) load() (*config.RecentActivityConfig, error) {
	return s.readConfig()
}

func (s *store) updateTask(taskID string, ts time.Time, source string) error {
	if taskID == "" {
		return fmt.Errorf("taskID is required")
	}

	cfg, err := s.readConfig()
	if err != nil {
		return err
	}
	updated := applyTaskUpdate(cfg, taskID, ts, source, s.maxEntries)
	return s.writeConfig(updated)
}

func (s *store) readConfig() (*config.RecentActivityConfig, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return newEmptyConfig(), nil
		}
		return nil, err
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return newEmptyConfig(), nil
	}

	cfg := &config.RecentActivityConfig{}
	u := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := u.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *store) writeConfig(cfg *config.RecentActivityConfig) error {
	if cfg == nil {
		cfg = newEmptyConfig()
	}
	entries := append([]*config.RecentActivityEntry(nil), cfg.GetEntries()...)
	sortEntriesByActivity(entries)
	cfg.SetEntries(entries)

	m := protojson.MarshalOptions{Indent: "  "}
	payload, err := m.Marshal(cfg)
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	mode := os.FileMode(defaultFileMode)
	if info, err := os.Stat(s.path); err == nil {
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}

	tempFile, err := os.CreateTemp(dir, "recent-activity-*")
	if err != nil {
		return err
	}
	tempName := tempFile.Name()
	defer func() {
		_ = os.Remove(tempName)
	}()

	if err := os.Chmod(tempName, mode); err != nil {
		_ = tempFile.Close()
		return err
	}

	if _, err := tempFile.Write(payload); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tempName, s.path); err != nil {
		return err
	}
	return nil
}

func newEmptyConfig() *config.RecentActivityConfig {
	return config.RecentActivityConfig_builder{}.Build()
}

func applyTaskUpdate(
	cfg *config.RecentActivityConfig,
	taskID string,
	ts time.Time,
	source string,
	maxEntries int,
) *config.RecentActivityConfig {
	if cfg == nil {
		cfg = newEmptyConfig()
	}

	entries := make([]*config.RecentActivityEntry, 0, len(cfg.GetEntries())+1)
	for _, entry := range cfg.GetEntries() {
		if entry == nil {
			continue
		}
		entryType := entry.GetType()
		if (entryType == config.RecentActivityEntryType_ENTRY_TYPE_TASK ||
			entryType == config.RecentActivityEntryType_ENTRY_TYPE_UNSPECIFIED) &&
			entry.GetTaskId() == taskID {
			continue
		}
		entries = append(entries, entry)
	}

	entries = append(entries, config.RecentActivityEntry_builder{
		Type:         config.RecentActivityEntryType_ENTRY_TYPE_TASK,
		TaskId:       taskID,
		LastActivity: timestamppb.New(ts),
		Source:       source,
	}.Build())

	sortEntriesByActivity(entries)
	if len(entries) > maxEntries {
		entries = entries[:maxEntries]
	}

	return config.RecentActivityConfig_builder{
		Entries: entries,
	}.Build()
}

func sortEntriesByActivity(entries []*config.RecentActivityEntry) {
	if len(entries) == 0 {
		return
	}
	sort.SliceStable(entries, func(i, j int) bool {
		left := activityTime(entries[i])
		right := activityTime(entries[j])
		if left.Equal(right) {
			return false
		}
		return left.After(right)
	})
}

func activityTime(entry *config.RecentActivityEntry) time.Time {
	return entry.GetLastActivity().AsTime()
}
