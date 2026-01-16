package recentactivity

import (
	"log/slog"
	"sort"
	"time"

	"github.com/devplaninc/devplan-cli/internal/pb/config"
	"github.com/devplaninc/devplan-cli/internal/utils/metadata"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
)

func SortClonedFeatures(features []workspace.ClonedFeature) []workspace.ClonedFeature {
	store, err := newDefaultStore()
	if err != nil {
		slog.Debug("Failed to create recent activity store", "err", err)
		return features
	}
	cfg, err := store.load()
	if err != nil {
		slog.Debug("Failed to load recent activity config", "err", err)
		return features
	}
	index := taskActivityIndex(cfg)
	return sortFeaturesByActivity(features, index)
}

func taskActivityIndex(cfg *config.RecentActivityConfig) map[string]time.Time {
	index := make(map[string]time.Time)
	if cfg == nil {
		return index
	}
	for _, entry := range cfg.GetEntries() {
		entryType := entry.GetType()
		if entryType != config.RecentActivityEntryType_ENTRY_TYPE_TASK &&
			entryType != config.RecentActivityEntryType_ENTRY_TYPE_UNSPECIFIED {
			continue
		}
		taskID := entry.GetTaskId()
		if taskID == "" {
			continue
		}
		ts := activityTime(entry)
		if existing, ok := index[taskID]; !ok || ts.After(existing) {
			index[taskID] = ts
		}
	}
	return index
}

func sortFeaturesByActivity(
	features []workspace.ClonedFeature,
	activity map[string]time.Time,
) []workspace.ClonedFeature {
	if len(features) == 0 || len(activity) == 0 {
		return features
	}

	items := make([]featureActivity, 0, len(features))
	for _, feature := range features {
		item := featureActivity{
			feature: feature,
		}
		meta, err := metadata.ReadMetadata(feature.FullPath)
		if err != nil {
			slog.Debug("Failed to read feature metadata", "path", feature.FullPath, "err", err)
		} else if meta != nil && meta.TaskID != "" {
			if ts, ok := activity[meta.TaskID]; ok {
				item.hasActivity = true
				item.activityTime = ts
			}
		}
		items = append(items, item)
	}

	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		if left.hasActivity && right.hasActivity {
			if left.activityTime.Equal(right.activityTime) {
				return false
			}
			return left.activityTime.After(right.activityTime)
		}
		if left.hasActivity != right.hasActivity {
			return left.hasActivity
		}
		return false
	})

	sorted := make([]workspace.ClonedFeature, len(items))
	for i, item := range items {
		sorted[i] = item.feature
	}
	return sorted
}

type featureActivity struct {
	feature      workspace.ClonedFeature
	hasActivity  bool
	activityTime time.Time
}
