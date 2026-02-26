package spec

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/utils/converters"
	"github.com/devplaninc/devplan-cli/internal/utils/gitws"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/picker"
	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"github.com/devplaninc/devplan-cli/internal/utils/recentactivity"
	"github.com/opensdd/osdd-api/clients/go/osdd/recipes"
	"github.com/opensdd/osdd-core/core"
	"github.com/opensdd/osdd-core/core/executable"
	"github.com/spf13/cobra"
)

var (
	startCmd = createStartCmd()
)

func createStartCmd() *cobra.Command {
	var companyID int32
	var taskID string
	var featureID string
	var ideType string
	var path string
	var branchName string
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start implementation of a task or feature in an AI IDE",
		Long: `Start implementation of a task or feature in an AI IDE.

Use -t/--task to start a single task (clones one repo, creates worktree).
Use -f/--feature to start a feature (clones all referenced repos into a parent folder).

Exactly one of -t or -f must be provided.`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if taskID == "" && featureID == "" {
				return fmt.Errorf("exactly one of --task (-t) or --feature (-f) must be provided")
			}
			if taskID != "" && featureID != "" {
				return fmt.Errorf("--task (-t) and --feature (-f) are mutually exclusive")
			}
			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			ctx := context.Background()
			cl := devplan.NewClient(devplan.Config{})

			var workspacePath string
			var execRecipe *recipes.ExecutableRecipe

			if featureID != "" {
				res := runStartFeature(ctx, cl, companyID, featureID, path)
				workspacePath = res.WorkspacePath
				execRecipe = res.ExecRecipe
				if err := recentactivity.RecordTaskActivity(featureID, "spec_start"); err != nil {
					slog.Debug("Failed to record recent feature activity", "featureID", featureID, "err", err)
				}
			} else {
				docResp, err := cl.GetDocument(companyID, taskID)
				check(err)
				task := docResp.GetDocument()
				if err := recentactivity.RecordTaskActivity(taskID, "spec_start"); err != nil {
					slog.Debug("Failed to record recent task activity", "taskID", taskID, "err", err)
				}

				workspacePath = path
				if workspacePath == "" {
					details, err := converters.GetTaskDetails(task)
					check(err)
					cloneRes, err := gitws.InteractiveClone(ctx, &picker.TargetCmd{
						CompanyID: companyID,
						ProjectID: task.GetProjectId(),
						FeatureID: task.GetParentId(),
						TaskID:    taskID,
						IDEName:   ideType,
						Yes:       true,
					}, details.GetRepoName(), branchName)
					check(err)
					workspacePath = cloneRes.RepoPath
				}

				execRecipe, err = cl.GetTaskExecRecipe(companyID, taskID)
				check(err)
			}

			// Common recipe execution
			if execRecipe.GetEntryPoint() == nil {
				execRecipe.SetEntryPoint(&recipes.EntryPoint{})
			}
			execRecipe.GetEntryPoint().SetWorkspace(recipes.WorkspaceConfig_builder{
				Enabled:  true,
				Path:     workspacePath,
				Absolute: true,
			}.Build())
			if execRecipe.GetEntryPoint().GetIdeType() == "" {
				execRecipe.GetEntryPoint().SetIdeType(ideType)
			}
			genCtx := &core.GenerationContext{
				ExecRecipe:    execRecipe,
				OutputCMDOnly: prefs.InstructionFile != "",
			}
			r := executable.ForRecipe(execRecipe)
			_, err := r.Materialize(ctx, genCtx)
			check(err)
			result, err := r.Execute(ctx, genCtx)
			check(err)
			check(ide.WriteLaunchResult(result.LaunchResult))
		},
	}
	cmd.Flags().Int32VarP(&companyID, "company", "c", 0, "Company id for a recipe")
	cmd.Flags().StringVarP(&taskID, "task", "t", "", "Task id to start")
	cmd.Flags().StringVarP(&featureID, "feature", "f", "", "Feature id to start")
	cmd.Flags().StringVarP(&ideType, "ide", "i", "", "IDE to use ('claude', 'cursor-cli' only right now)")
	cmd.Flags().StringVarP(&path, "path", "p", "", "Path to use as workspace. If provided, skip cloning")
	cmd.Flags().StringVarP(&branchName, "branch", "b", "", "Branch to checkout after workspace preparation (task mode only)")
	_ = cmd.MarkFlagRequired("company")
	_ = cmd.MarkFlagRequired("ide")
	return cmd
}
