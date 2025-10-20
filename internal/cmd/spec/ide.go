package spec

import (
	"context"

	"github.com/devplaninc/adcp-core/adcp/core"
	"github.com/devplaninc/adcp-core/adcp/core/recipes"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/spf13/cobra"
)

var (
	ideCmd = createIDECmd()
)

func createIDECmd() *cobra.Command {
	var companyID int32
	cmd := &cobra.Command{
		Use:    "ide",
		Short:  "Get recipe for ide",
		Hidden: true,
		Run: func(_ *cobra.Command, _ []string) {
			cl := devplan.NewClient(devplan.Config{})
			recipe, err := cl.GetIDERecipe(companyID)
			check(err)
			r := recipes.Recipe{}
			res, err := r.Materialize(context.Background(), recipe)
			check(err)
			check(core.PersistMaterializedResult(context.Background(), ".", res))
		},
	}
	cmd.Flags().Int32VarP(&companyID, "company", "c", -1, "Company id for a rule")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
