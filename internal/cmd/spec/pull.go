package spec

import (
	"context"
	"fmt"
	"os"

	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/opensdd/osdd-api/clients/go/osdd/recipes"
	"github.com/opensdd/osdd-core/core"
	"github.com/opensdd/osdd-core/core/executable"
	"github.com/spf13/cobra"
)

var (
	pullCmd = createPullCmd()
)

func createPullCmd() *cobra.Command {
	var companyID int32
	var taskID string
	var ideType string
	var path string
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Download spec/context files for a task to the current directory",
		Long: `Downloads all spec/context files for the selected IDE directly into
the current working directory (or specified --path), overwriting existing files.
Does not clone repository or launch IDE.`,
		Run: func(_ *cobra.Command, _ []string) {
			ctx := context.Background()

			// Determine output path: use --path flag or current working directory
			outputPath := path
			if outputPath == "" {
				cwd, err := os.Getwd()
				check(err)
				outputPath = cwd
			} else {
				// Validate that the provided path exists
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					check(fmt.Errorf("output path does not exist: %s", outputPath))
				}
			}

			cl := devplan.NewClient(devplan.Config{})
			execRecipe, err := cl.GetTaskExecRecipe(companyID, taskID)
			check(err)
			if execRecipe.GetEntryPoint() == nil {
				execRecipe.SetEntryPoint(&recipes.EntryPoint{})
			}
			execRecipe.GetEntryPoint().SetWorkspace(recipes.WorkspaceConfig_builder{
				Enabled:  true,
				Path:     outputPath,
				Absolute: true,
			}.Build())
			if execRecipe.GetEntryPoint().GetIdeType() == "" {
				execRecipe.GetEntryPoint().SetIdeType(ideType)
			}
			genCtx := &core.GenerationContext{
				ExecRecipe:    execRecipe,
				OutputCMDOnly: true,
			}
			r := executable.ForRecipe(execRecipe)
			_, err = r.Materialize(ctx, genCtx)
			check(err)
			_, err = r.Execute(ctx, genCtx)
			check(err)
			fmt.Printf("Spec files downloaded successfully to: %s\n", outputPath)
		},
	}
	cmd.Flags().Int32VarP(&companyID, "company", "c", 0, "Company ID")
	cmd.Flags().StringVarP(&taskID, "task", "t", "", "Task ID to pull specs for")
	cmd.Flags().StringVarP(&ideType, "ide", "i", "", "IDE type ('claude', 'cursor-cli')")
	cmd.Flags().StringVarP(&path, "path", "p", "", "Output directory (default: current working directory)")
	_ = cmd.MarkFlagRequired("company")
	_ = cmd.MarkFlagRequired("task")
	_ = cmd.MarkFlagRequired("ide")
	return cmd
}
