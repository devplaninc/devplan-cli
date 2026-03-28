package define

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/spf13/cobra"
)

const (
	pollInterval = 5 * time.Second
	maxPollTime  = 10 * time.Minute
)

func createStatusCmd() *cobra.Command {
	var companyID int32
	var pendingJobID string
	var wait bool
	var output string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check feature initialization status",
		Long:  "Fetches initialization status by pending job ID. With --wait, polls until done or failed.",
		Run: func(_ *cobra.Command, _ []string) {
			cl := devplan.NewClient(devplan.Config{})
			jsonOut := output == "json"

			if !wait {
				// Single status check
				resp, err := cl.GetQuickWinInitStatus(companyID, pendingJobID)
				if err != nil {
					if jsonOut {
						printJSON(statusResult{Error: err.Error()})
					} else {
						fmt.Println(out.Failf("Error: %v", err))
					}
					os.Exit(1)
				}
				if jsonOut {
					printJSON(statusResult{Status: resp.GetStatus(), ProjectID: resp.GetProjectId()})
				} else {
					printStatus(resp.GetStatus(), resp.GetProjectId())
				}
				return
			}

			// Polling mode
			deadline := time.Now().Add(maxPollTime)
			for {
				resp, err := cl.GetQuickWinInitStatus(companyID, pendingJobID)
				if err != nil {
					if jsonOut {
						printJSON(statusResult{Error: err.Error()})
					} else {
						fmt.Println(out.Failf("Error: %v", err))
					}
					os.Exit(1)
				}

				if !jsonOut {
					printStatus(resp.GetStatus(), resp.GetProjectId())
				}

				if resp.GetStatus() == "completed" {
					if jsonOut {
						printJSON(statusResult{Status: resp.GetStatus(), ProjectID: resp.GetProjectId()})
					}
					os.Exit(0)
				}
				if resp.GetStatus() == "failed" {
					if jsonOut {
						printJSON(statusResult{Status: resp.GetStatus(), ProjectID: resp.GetProjectId()})
					}
					os.Exit(1)
				}

				if time.Now().After(deadline) {
					if jsonOut {
						printJSON(statusResult{Error: "timed out waiting for initialization to complete"})
					} else {
						fmt.Println(out.Failf("Timed out waiting for initialization to complete"))
					}
					os.Exit(1)
				}

				time.Sleep(pollInterval)
			}
		},
	}

	cmd.Flags().Int32VarP(&companyID, "company", "c", 0, "Company ID")
	cmd.Flags().StringVarP(&pendingJobID, "job", "j", "", "Pending job ID")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "Poll until terminal status (completed or failed)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output format (json)")
	_ = cmd.MarkFlagRequired("company")
	_ = cmd.MarkFlagRequired("job")
	return cmd
}

type statusResult struct {
	Status    string `json:"status"`
	ProjectID string `json:"projectId,omitempty"`
	Error     string `json:"error,omitempty"`
}

type featureResult struct {
	PendingJobID string `json:"pendingJobId,omitempty"`
	ProjectID    string `json:"projectId,omitempty"`
	Error        string `json:"error,omitempty"`
}

func printStatus(status string, projectID string) {
	switch status {
	case "completed":
		fmt.Println(out.Successf("Status: %s", status))
	case "failed":
		fmt.Println(out.Failf("Status: %s", status))
	default:
		fmt.Printf("  Status: %s\n", out.H(status))
	}
	if projectID != "" {
		fmt.Printf("  Project ID: %s\n", projectID)
	}
}

func printJSON(v any) {
	data, _ := json.Marshal(v)
	fmt.Println(string(data))
}
