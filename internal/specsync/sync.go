package specsync

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/devplaninc/webapp/golang/pb/api/devplan/services/web/company"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/artifacts"
)

const DefaultSyncInterval = 10 * time.Second

// shouldUpload determines if an artifact needs to be uploaded
func shouldUpload(spec Spec, specs []*artifacts.SpecDetails) bool {
	var existing *artifacts.SpecDetails
	for _, s := range specs {
		if s.GetName() == spec.Name {
			existing = s
			break
		}
	}
	if existing == nil {
		// Spec not on server, needs upload
		return true
	}
	// Upload if checksums differ
	return existing.GetChecksum() != spec.Checksum
}

func (s *Syncer) uploadSpec(ctx context.Context, spec Spec) error {
	// Get or create mutex for this spec
	lockInterface, _ := s.specsLocks.LoadOrStore(spec.Path, &sync.Mutex{})
	lock := lockInterface.(*sync.Mutex)

	lock.Lock()
	defer lock.Unlock()

	// Check if context is canceled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create upload request
	req := company.UploadSpecRequest_builder{
		Name:     spec.Name,
		Content:  string(spec.Content),
		Checksum: spec.Checksum,
	}.Build()
	return s.client.UploadTaskSpec(s.companyID, s.taskID, req)
}

// runSync executes one sync run and returns the result
func (s *Syncer) runSync(ctx context.Context) *SyncResult {
	slog.Debug("Running sync")
	defer slog.Debug("Sync finished")
	result := &SyncResult{}

	// Discover local specs
	localSpecs, err := DiscoverTaskSpecs(s.taskDir)
	if err != nil {
		result.Failed = 1
		result.Errors = append(result.Errors, err)
		return result
	}

	if len(localSpecs) == 0 {
		slog.Debug("Running sync: no local specs")
		return result
	}

	serverSpecsResp, err := s.client.GetTaskSpecs(s.companyID, s.taskID)
	if err != nil {
		result.Failed = len(localSpecs)
		result.Errors = append(result.Errors, err)
		return result
	}

	// Process each local spec
	for _, localSpec := range localSpecs {
		if !shouldUpload(localSpec, serverSpecsResp.GetSpecs()) {
			result.Skipped++
			continue
		}
		slog.Debug("Uploading spec", "path", localSpec.Path)

		err := s.uploadSpec(ctx, localSpec)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, err)
			slog.Info("Failed to upload spec", "path", localSpec.Path, "err", err)
		} else {
			slog.Info("Spec uploaded", "path", localSpec.Path)
			result.Uploaded++
		}
	}

	return result
}
