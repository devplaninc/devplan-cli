package spec

import (
	"context"

	"github.com/devplaninc/adcp/clients/go/adcp"
	"github.com/devplaninc/devplan-cli/internal/utils/recipes"
	"github.com/spf13/cobra"
)

var (
	taskCmd = createTaskCmd()
)

func createTaskCmd() *cobra.Command {
	var companyID int32
	var taskID string
	var ideType string
	ctx := context.Background()
	cmd := &cobra.Command{
		Use:    "task",
		Short:  "Get recipe for a task",
		Hidden: true,
		Run: func(_ *cobra.Command, _ []string) {
			err := recipes.PersistForTask(ctx, companyID, taskID, recipes.PersistOptions{
				Path: ".",
				EntryPoint: adcp.EntryPoint_builder{
					IdeType: ideType,
				}.Build(),
			})
			check(err)
		},
	}
	cmd.Flags().Int32VarP(&companyID, "company", "c", -1, "Company id for a recipe")
	cmd.Flags().StringVarP(&taskID, "task", "t", "", "Task id for a recipe")
	cmd.Flags().StringVarP(&ideType, "ide", "i", "", "IDE to use (claude only right now)")
	_ = cmd.MarkFlagRequired("company")
	_ = cmd.MarkFlagRequired("task")
	_ = cmd.MarkFlagRequired("ide")
	return cmd
}
