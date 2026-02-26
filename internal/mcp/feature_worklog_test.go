package mcp

import (
	"testing"

	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/worklog"
)

func TestFeatureWorkLogItemBuilder(t *testing.T) {
	featureID := "feat-123"
	companyID := int32(42)

	wlType := getWorkloadType("research")
	item := worklog.WorkLogItem_builder{
		Message:           "test message",
		CompanyId:         &companyID,
		FeatureId:         &featureID,
		Type:              wlType,
		Stage:             "started",
		ActionDescription: "Research",
		AgentName:         "Claude",
	}.Build()

	if item.GetFeatureId() != "feat-123" {
		t.Fatalf("expected FeatureId 'feat-123', got %q", item.GetFeatureId())
	}
	if item.GetCompanyId() != 42 {
		t.Fatalf("expected CompanyId 42, got %d", item.GetCompanyId())
	}
	if item.GetType() != worklog.WorkLogType_RESEARCH {
		t.Fatalf("expected type RESEARCH, got %v", item.GetType())
	}
	if item.HasTaskId() {
		t.Fatalf("expected TaskId to not be set, but it was: %q", item.GetTaskId())
	}
	if item.GetMessage() != "test message" {
		t.Fatalf("expected message 'test message', got %q", item.GetMessage())
	}
	if item.GetAgentName() != "Claude" {
		t.Fatalf("expected AgentName 'Claude', got %q", item.GetAgentName())
	}
}
