package specsync

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Syncer manages specs synchronization
type Syncer struct {
	client    Client
	companyID int32
	taskID    string
	taskDir   string
	interval  time.Duration

	// Concurrency control
	runMu      sync.Mutex // Single-flight guard for sync runs
	specsLocks sync.Map   // map[specName]*sync.Mutex - Note: grows unbounded over time, acceptable for typical session lengths
}

// NewSyncer creates a new Syncer instance
func NewSyncer(client Client, companyID int32, taskID, taskDir string, interval time.Duration) *Syncer {
	if interval == 0 {
		interval = DefaultSyncInterval
	}
	return &Syncer{
		client:    client,
		taskDir:   taskDir,
		companyID: companyID,
		taskID:    taskID,
		interval:  interval,
	}
}

// TriggerOnce runs a single sync operation
// Returns immediately if a sync is already in progress
func (s *Syncer) TriggerOnce(_ context.Context) *SyncResult {
	// Try to acquire the run lock
	if !s.runMu.TryLock() {
		slog.Debug("Sync already in progress, skipping")
		// Another sync is in progress, skip
		return &SyncResult{Skipped: -1} // -1 indicates skipped due to concurrent run
	}
	defer s.runMu.Unlock()

	// Using background context to avoid cancelling on command termination.
	result := s.runSync(context.Background())
	if len(result.Errors) > 0 {
		slog.Error("Failed to sync specs", "errors", result.Errors)
	}
	return result
}

// RunBackground starts the background sync loop
// Blocks until context is canceled
func (s *Syncer) RunBackground(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.TriggerOnce(ctx)
		}
	}
}
