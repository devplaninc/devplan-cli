package mcp

import (
	"context"
	"strings"

	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/worklog"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type WorkLogReportInput struct {
	Message           string `json:"message" jsonschema:"message describing what happened"`
	TaskID            string `json:"taskId,omitempty" jsonschema:"optional task identifier"`
	CompanyID         int32  `json:"companyId,omitempty" jsonschema:"required company identifier"`
	Type              string `json:"type,omitempty" jsonschema:"optional worklog type. If present should be one of 'full_workflow', 'research', 'planning', 'coding', 'review', 'address_review', 'analysis', 'commit', 'finalize'"`
	Stage             string `json:"stage,omitempty" jsonschema:"optional worklog stage. If present should be one of 'started', 'running', 'ended', 'error'"`
	ActionDescription string `json:"actionDescription,omitempty" jsonschema:"description of the action. Should be a 1-2 words description of what this work lof is about"`
	AgentName         string `json:"agentName,omitempty" jsonschema:"Based on your current execution environment, report the appropriate agent name, e.g. Cursor or Claude or Codex, etc. Use 'Cursor' when operating within Cursor IDE, 'Claude' when operating within ClaudeCode, etc.'"`
}

type WorkLogReportOutput struct {
}

func reportWorkLog(_ context.Context, _ *mcp.CallToolRequest, input WorkLogReportInput) (*mcp.CallToolResult, WorkLogReportOutput, error) {
	cl := devplan.NewClient(devplan.Config{})
	wlType := getWorkloadType(input.Type)
	customType := ""
	if wlType == worklog.WorkLogType_WORK_LOG_TYPE_UNSPECIFIED {
		customType = input.Type
	}
	_, err := cl.SubmitWorklogItem(input.CompanyID, worklog.WorkLogItem_builder{
		Message:           input.Message,
		CompanyId:         &input.CompanyID,
		TaskId:            &input.TaskID,
		Type:              wlType,
		CustomType:        customType,
		Stage:             input.Stage,
		ActionDescription: input.ActionDescription,
		AgentName:         input.AgentName,
	}.Build())
	return nil, WorkLogReportOutput{}, err
}

func getWorkloadType(wlType string) worklog.WorkLogType {
	wlTypeInt, ok := worklog.WorkLogType_value[strings.ToUpper(wlType)]
	if !ok {
		return worklog.WorkLogType_WORK_LOG_TYPE_UNSPECIFIED
	}
	return worklog.WorkLogType(wlTypeInt)
}
