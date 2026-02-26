package mcp

import (
	"context"
	"log/slog"

	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/utils/recentactivity"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/worklog"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type FeatureWorkLogReportInput struct {
	Message           string `json:"message" jsonschema:"message describing what happened"`
	FeatureID         string `json:"featureId,omitempty" jsonschema:"optional feature identifier"`
	CompanyID         int32  `json:"companyId,omitempty" jsonschema:"required company identifier"`
	Type              string `json:"type,omitempty" jsonschema:"optional worklog type. If present should be one of 'full_workflow', 'research', 'planning', 'coding', 'review', 'address_review', 'analysis', 'commit', 'finalize'"`
	Stage             string `json:"stage,omitempty" jsonschema:"optional worklog stage. If present should be one of 'started', 'running', 'ended', 'error'"`
	ActionDescription string `json:"actionDescription,omitempty" jsonschema:"description of the action. Should be a 1-2 words description of what this work log is about"`
	AgentName         string `json:"agentName,omitempty" jsonschema:"Based on your current execution environment, report the appropriate agent name, e.g. Cursor or Claude or Codex, etc. Use 'Cursor' when operating within Cursor IDE, 'Claude' when operating within ClaudeCode, etc.'"`
}

type FeatureWorkLogReportOutput struct {
}

func (s *Server) reportFeatureWorkLog(_ context.Context, _ *mcp.CallToolRequest, input FeatureWorkLogReportInput) (*mcp.CallToolResult, FeatureWorkLogReportOutput, error) {
	cl := devplan.NewClient(devplan.Config{})
	wlType := getWorkloadType(input.Type)
	customType := ""
	if wlType == worklog.WorkLogType_WORK_LOG_TYPE_UNSPECIFIED {
		customType = input.Type
	}
	item := worklog.WorkLogItem_builder{
		Message:           input.Message,
		CompanyId:         &input.CompanyID,
		FeatureId:         &input.FeatureID,
		Type:              wlType,
		CustomType:        customType,
		Stage:             input.Stage,
		ActionDescription: input.ActionDescription,
		AgentName:         input.AgentName,
	}.Build()
	_, err := cl.SubmitWorklogItem(input.CompanyID, item)
	slog.Info("Reporting feature worklog item", "item", item)

	// Record feature activity using the story ID (feature ID) so the feature
	// workspace sorts to the top in list/switch/clean.
	if input.FeatureID != "" {
		if raErr := recentactivity.RecordTaskActivity(input.FeatureID, "worklog"); raErr != nil {
			slog.Debug("Failed to record recent feature activity", "featureID", input.FeatureID, "err", raErr)
		}
	}

	return nil, FeatureWorkLogReportOutput{}, err
}
