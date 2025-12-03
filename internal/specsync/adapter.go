package specsync

import (
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/services/web/company"
)

// ClientAdapter wraps devplan.Client to implement Client
type ClientAdapter struct {
	client *devplan.Client
}

// NewClientAdapter creates a new adapter for the devplan client
func NewClientAdapter(client *devplan.Client) *ClientAdapter {
	return &ClientAdapter{client: client}
}

func (a *ClientAdapter) GetTaskSpecs(companyID int32, taskID string) (*company.GetTaskSpecsResponse, error) {
	return a.client.GetTaskSpecs(companyID, taskID)
}

func (a *ClientAdapter) UploadTaskSpec(companyID int32, taskID string, req *company.UploadSpecRequest) error {
	return a.client.UploadTaskSpec(companyID, taskID, req)
}
