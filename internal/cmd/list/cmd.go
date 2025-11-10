package list_cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
	"golang.design/x/clipboard"

	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/spf13/cobra"
)

var (
	Cmd = create()
)

func create() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List cloned features",
		Long:    `List all cloned features in the workspace`,
		Run: func(_ *cobra.Command, _ []string) {
			runList()
		},
	}
	return cmd
}

func runList() {
	features, err := workspace.ListClonedFeatures()
	check(err)
	if len(features) == 0 {
		fmt.Println("No cloned features found. Use 'devplan clone' to clone a feature first.")
		os.Exit(0)
	}

	var inputs []huh.Option[string]

	for _, f := range features {
		toCopy := f.FullPath
		if len(f.Repos) == 1 {
			fullNames := f.Repos[0].Repo.FullNames
			if len(fullNames) > 0 {
				parts := strings.Split(fullNames[0], "/")
				if len(parts) == 2 {
					toCopy = path.Join(toCopy, parts[1])
				}
			}
		}
		inputs = append(inputs, huh.NewOption(fmt.Sprintf("%v - %v", f.GetDisplayName(), toCopy), toCopy))
	}

	var featurePath string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a feature to copy its path to the clipboard:").
				Options(inputs...).
				Value(&featurePath),
		),
	)
	check(form.Run())
	clipboard.Write(clipboard.FmtText, []byte(featurePath))
	fmt.Println(out.Successf("Copied %s to clipboard", out.H(featurePath)))
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("Error: %v", err))
		os.Exit(1)
	}
}
