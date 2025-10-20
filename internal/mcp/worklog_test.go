package mcp

import (
	"testing"

	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/worklog"
)

func TestGetWorkloadType_KnownValuesCaseInsensitive(t *testing.T) {
	cases := []struct {
		in   string
		want worklog.WorkLogType
	}{
		{"research", worklog.WorkLogType_RESEARCH},
		{"Research", worklog.WorkLogType_RESEARCH},
		{"RESEARCH", worklog.WorkLogType_RESEARCH},
		{"planning", worklog.WorkLogType_PLANNING},
		{"Planning", worklog.WorkLogType_PLANNING},
		{"PLANNING", worklog.WorkLogType_PLANNING},
		{"execution", worklog.WorkLogType_EXECUTION},
		{"Execution", worklog.WorkLogType_EXECUTION},
		{"EXECUTION", worklog.WorkLogType_EXECUTION},
	}

	for _, tc := range cases {
		got := getWorkloadType(tc.in)
		if got != tc.want {
			t.Fatalf("getWorkloadType(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestGetWorkloadType_UnknownDefaultsToUnspecified(t *testing.T) {
	cases := []string{"", "unknown", "foo", "123"}
	for _, in := range cases {
		got := getWorkloadType(in)
		if got != worklog.WorkLogType_WORK_LOG_TYPE_UNSPECIFIED {
			t.Fatalf("getWorkloadType(%q) = %v, want UNSPECIFIED", in, got)
		}
	}
}
