package spec

import (
	"context"
	"fmt"

	"github.com/devplaninc/adcp-core/adcp/core"
	"github.com/devplaninc/adcp-core/adcp/core/executable"
	"github.com/devplaninc/adcp/clients/go/adcp"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/converters"
	"github.com/devplaninc/devplan-cli/internal/utils/gitws"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/picker"
	"github.com/spf13/cobra"
)

var (
	startCmd = createStartCmd()
)

func createStartCmd() *cobra.Command {
	var companyID int32
	var taskID string
	var ideType string
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start implementation of a task in an AI IDE",
		Run: func(_ *cobra.Command, _ []string) {
			ctx := context.Background()
			cl := devplan.NewClient(devplan.Config{})
			docResp, err := cl.GetDocument(companyID, taskID)
			check(err)
			task := docResp.GetDocument()
			details, err := converters.GetTaskDetails(task)
			check(err)
			cloneRes, err := gitws.InteractiveClone(ctx, &picker.TargetCmd{
				CompanyID: companyID,
				ProjectID: task.GetProjectId(),
				FeatureID: task.GetParentId(),
				TaskID:    taskID,
				IDEName:   ideType,
				Yes:       true,
			}, details.GetRepoName())
			check(err)
			recipe, err := cl.GetTaskRecipe(companyID, taskID)
			check(err)
			executableRecipe := adcp.ExecutableRecipe_builder{
				Recipe: recipe,
				EntryPoint: adcp.EntryPoint_builder{
					IdeType: ideType,
				}.Build(),
			}.Build()
			r := executable.ForRecipe(executableRecipe)
			res, err := r.Materialize(ctx)
			check(err)
			fmt.Println(out.Hf("Materialized: %+v", cloneRes.RepoPath))
			check(core.PersistMaterializedResult(ctx, cloneRes.RepoPath, res))
			_, err = ide.LaunchIDE(ide.IDE(ideType), cloneRes.RepoPath, false)
			check(err)
		},
	}
	cmd.Flags().Int32VarP(&companyID, "company", "c", 0, "Company id for a recipe")
	cmd.Flags().StringVarP(&taskID, "task", "t", "", "Task id for a recipe")
	cmd.Flags().StringVarP(&ideType, "ide", "i", "", "IDE to use ('claude', 'cursor-cli' only right now)")
	_ = cmd.MarkFlagRequired("company")
	_ = cmd.MarkFlagRequired("task")
	_ = cmd.MarkFlagRequired("ide")
	return cmd
}
