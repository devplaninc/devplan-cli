package picker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTargetCmd_PreRun_TaskValidation(t *testing.T) {
	tests := []struct {
		name      string
		cmd       TargetCmd
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid task with feature",
			cmd: TargetCmd{
				FeatureID: "feature-123",
				TaskID:    "task-456",
			},
			expectErr: false,
		},
		{
			name: "valid task without feature",
			cmd: TargetCmd{
				TaskID: "task-456",
			},
			expectErr: false,
		},
		{
			name: "task with single-shot should fail",
			cmd: TargetCmd{
				TaskID:     "task-456",
				SingleShot: true,
			},
			expectErr: true,
			errMsg:    "--task cannot be used with -s (--single-shot)",
		},
		{
			name: "single-shot with feature should fail",
			cmd: TargetCmd{
				FeatureID:  "feature-123",
				SingleShot: true,
			},
			expectErr: true,
			errMsg:    "-s (--single-shot) cannot be used with -f (--feature)",
		},
		{
			name: "steps without feature should fail",
			cmd: TargetCmd{
				Steps: true,
			},
			expectErr: true,
			errMsg:    "--steps must be used together with -f (--feature)",
		},
		{
			name: "valid steps with feature",
			cmd: TargetCmd{
				FeatureID: "feature-123",
				Steps:     true,
			},
			expectErr: false,
		},
		{
			name: "task with steps and feature should be valid",
			cmd: TargetCmd{
				FeatureID: "feature-123",
				TaskID:    "task-456",
				Steps:     true,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cmd.PreRun(nil, nil)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
