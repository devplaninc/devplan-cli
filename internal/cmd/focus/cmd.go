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
	targetPicker := &picker.TargetCmd{}
	cmd := &cobra.Command{
		Use:     "focus",
		Aliases: []string{"f"},
		Short:   "Focus on a specific feature of a project",
		PreRunE: targetPicker.PreRun,
		Run: func(_ *cobra.Command, _ []string) {
			runFocus(targetPicker)
		},
	}
	targetPicker.Prepare(cmd)
	return cmd
}

func runFocus(targetPicker *picker.TargetCmd) {
	repo := git.EnsureInRepo()
	out.Psuccessf("Current repository: %+v\n", repo.FullNames[0])
	ides, err := picker.AssistantForIDE(targetPicker.IDEName)
	check(err)
	target, err := picker.Target(targetPicker)
	check(err)
	project := target.ProjectWithDocs
	summary, err := loaders.RepoSummary(target, repo)
	check(err)
	featPrompt, err := picker.GetTargetPrompt(target, project.GetDocs())
	check(err)

	if err := ide.CleanupCurrentFeaturePrompts(ides); err != nil {
		out.Pfailf("Failed to clean up prompt files: %v\n", err)
	}

	check(ide.WriteMultiIDE(ides, featPrompt, summary, targetPicker.Yes))

	fmt.Println("\nNow you can start your IDE and ask AI assistant to execute current feature. Happy coding!")
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("%v", err))
		os.Exit(1)
	}
}
