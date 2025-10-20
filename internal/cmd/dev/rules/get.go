package rules

import (
	"fmt"

	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/spf13/cobra"
)

var (
	getCmd = createGetCmd()
)

func createGetCmd() *cobra.Command {
	var ruleName string
	var companyID int32
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a rule by name",
		Run: func(_ *cobra.Command, _ []string) {
			cl := devplan.NewClient(devplan.Config{})
			resp, err := cl.GetDevRule(companyID, ruleName)
			check(err)
			fmt.Println(resp.GetRule())
		},
	}
	cmd.Flags().StringVarP(&ruleName, "name", "n", "", "name of the rule")
	cmd.Flags().Int32VarP(&companyID, "company", "c", -1, "Company id for a rule")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
