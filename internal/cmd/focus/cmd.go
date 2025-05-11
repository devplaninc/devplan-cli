package focus

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/loaders"
	"github.com/devplaninc/devplan-cli/internal/utils/picker"
	"github.com/spf13/cobra"
	"os"
)

var (
	Cmd = create()
)

func create() *cobra.Command {
	featPicker := &picker.FeatureCmd{}
	cmd := &cobra.Command{
		Use:     "focus",
		Aliases: []string{"f"},
		Short:   "Focus on a specific feature of a project",
		PreRunE: featPicker.PreRun,
		Run: func(_ *cobra.Command, _ []string) {
			runFocus(featPicker)
		},
	}
	return cmd
}

func runFocus(featPicker *picker.FeatureCmd) {
	repo := git.EnsureInRepo()
	out.Psuccessf("Current repository: %+v\n", repo.FullNames[0])
	ides, err := picker.IDEs(featPicker.IDEName)
	check(err)
	feat, err := picker.Feature(featPicker)
	check(err)
	feature := feat.Feature
	project := feat.ProjectWithDocs
	summary, err := loaders.RepoSummary(feature, repo)
	check(err)
	featPrompt, err := picker.GetFeaturePrompt(feature.GetId(), project.GetDocs())
	check(err)
	check(ide.WriteMultiIDE(ides, featPrompt, summary, featPicker.Yes))
	fmt.Println("\nNow you can start your IDE and ask AI assistant to execute current feature. Happy coding!")
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("%v", err))
		os.Exit(1)
	}
}
