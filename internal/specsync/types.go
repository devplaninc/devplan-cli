package specsync

import (
	"github.com/devplaninc/webapp/golang/pb/api/devplan/services/web/company"
)

// Client interface for artifact operations
type Client interface {
	GetTaskSpecs(companyID int32, taskID string) (*company.GetTaskSpecsResponse, error)
	UploadTaskSpec(companyID int32, taskID string, req *company.UploadSpecRequest) error
}

// SyncResult holds results of a sync run
type SyncResult struct {
	Uploaded int
	Skipped  int
	Failed   int
	Errors   []error
}

// Spec represents a local artifact file
type Spec struct {
	Name     string
	Path     string
	Checksum string
	Content  []byte
}
