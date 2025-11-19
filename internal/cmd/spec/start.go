package spec

import (
	"context"
	"fmt"

	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/converters"
	"github.com/devplaninc/devplan-cli/internal/utils/gitws"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/picker"
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
	var ideType string
	var path string
	var branchName string
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start implementation of a task in an AI IDE",
		Run: func(_ *cobra.Command, _ []string) {
			ctx := context.Background()
			cl := devplan.NewClient(devplan.Config{})
			docResp, err := cl.GetDocument(companyID, taskID)
			check(err)
			task := docResp.GetDocument()
			repoPath := path
			if repoPath == "" {
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
				repoPath = cloneRes.RepoPath
			}

			recipe, err := cl.GetTaskRecipe(companyID, taskID)
			check(err)
			executableRecipe := recipes.ExecutableRecipe_builder{
				Recipe: recipe,
				EntryPoint: recipes.EntryPoint_builder{
					IdeType: ideType,
				}.Build(),
			}.Build()
			r := executable.ForRecipe(executableRecipe)
			genCtx := &core.GenerationContext{
				ExecRecipe: executableRecipe,
			}
			res, err := r.Materialize(ctx, genCtx)
			check(err)
			fmt.Println(out.Hf("Materialized: %+v", repoPath))
			check(core.PersistMaterializedResult(ctx, repoPath, res))
			_, err = ide.LaunchIDE(ide.IDE(ideType), repoPath, false)
			check(err)
		},
	}
	cmd.Flags().Int32VarP(&companyID, "company", "c", 0, "Company id for a recipe")
	cmd.Flags().StringVarP(&taskID, "task", "t", "", "Task id for a recipe")
	cmd.Flags().StringVarP(&ideType, "ide", "i", "", "IDE to use ('claude', 'cursor-cli' only right now)")
	cmd.Flags().StringVarP(&path, "path", "p", "", "Path to the repository. If provided, do not clone, load context into the provided path")
	cmd.Flags().StringVarP(&branchName, "branch", "b", "", "Branch to checkout after workspace preparation. If remote with this name exists, it is checked out.")
	_ = cmd.MarkFlagRequired("company")
	_ = cmd.MarkFlagRequired("task")
	_ = cmd.MarkFlagRequired("ide")
	return cmd
}
