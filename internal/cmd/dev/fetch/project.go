package fetch

import (
	"fmt"

	"github.com/devplaninc/adcp/clients/go/adcp"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	projectCmd = createProjectCmd()
)

func createProjectCmd() *cobra.Command {
	var companyID int32
	var projectID string
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Get project documents in a Pre Fetch format",
		Run: func(_ *cobra.Command, _ []string) {
			cl := devplan.NewClient(devplan.Config{})
			resp, err := cl.GetProjectDocuments(companyID, projectID)
			check(err)
			var entries []*adcp.FetchedData
			for _, d := range resp.GetDocuments() {
				entries = append(entries, adcp.FetchedData_builder{
					Id:   d.GetId(),
					Data: d.GetContent(),
				}.Build())
			}
			result := adcp.PrefetchResult_builder{Data: entries}.Build()
			m := protojson.MarshalOptions{Indent: "  "}
			fmt.Println(m.Format(result))
		},
	}
	cmd.Flags().Int32VarP(&companyID, "company", "c", -1, "Company id to fetch")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Project id to fetch")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
