package define

import (
	"fmt"
	"os"

	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/services/web/company"
	"github.com/spf13/cobra"
)

func createFeatureCmd() *cobra.Command {
	var companyID int32
	var description string
	var output string

	cmd := &cobra.Command{
		Use:   "feature",
		Short: "Start feature project initialization",
		Long:  "Creates a new feature project immediately and starts background generation of the user story.",
		Run: func(_ *cobra.Command, _ []string) {
			cl := devplan.NewClient(devplan.Config{})
			jsonOut := output == "json"

			req := company.StartQuickWinInitRequest_builder{
				Description: description,
			}.Build()

			resp, err := cl.StartQuickWinInit(companyID, req)
			if err != nil {
				if jsonOut {
					printJSON(featureResult{Error: err.Error()})
				} else {
					fmt.Println(out.Failf("Error: %v", err))
				}
				os.Exit(1)
			}

			if jsonOut {
				printJSON(featureResult{PendingJobID: resp.GetPendingJobId(), ProjectID: resp.GetProjectId()})
			} else {
				fmt.Println(out.Successf("Feature initialization started"))
				fmt.Printf("  Pending Job ID: %s\n", out.H(resp.GetPendingJobId()))
				fmt.Printf("  Project ID:     %s\n", out.H(resp.GetProjectId()))
			}
		},
	}

	cmd.Flags().Int32VarP(&companyID, "company", "c", 0, "Company ID")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description of the feature project")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output format (json)")
	_ = cmd.MarkFlagRequired("company")
	_ = cmd.MarkFlagRequired("description")
	return cmd
}
