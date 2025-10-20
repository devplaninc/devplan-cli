package clone

import (
	"context"
	"fmt"
	"os"

	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/gitws"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/loaders"
	"github.com/devplaninc/devplan-cli/internal/utils/picker"
	"github.com/spf13/cobra"
)

var (
	Cmd = create()
)

func create() *cobra.Command {
	targetPicker := &picker.TargetCmd{}
	var repoName string
	var start bool
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone a repository and focus on a feature",
		Long: `Clone a repository and focus on a feature.
This command streamlines the workflow of cloning a repository and focusing on a feature.
It will clone the repository into the configured workplace directory and set up the necessary rules.`,
		PreRunE: targetPicker.PreRun,
		Run: func(_ *cobra.Command, _ []string) {
			ctx := context.Background()
			assistants, err := picker.AssistantForIDE(targetPicker.IDEName)
			check(err)
			cloneRes, err := gitws.InteractiveClone(ctx, targetPicker, repoName)
			check(err)
			target := cloneRes.Target
			gitRepo := cloneRes.RepoInfo
			summary, err := loaders.RepoSummary(target, gitRepo)
			check(err)
			prompt, err := picker.GetTargetPrompt(target, target.ProjectWithDocs.GetDocs())
			check(err)

			if err := ide.CleanupCurrentFeaturePrompts(assistants); err != nil {
				out.Pfailf("Failed to clean up prompt files: %v\n", err)
			}

			check(os.Chdir(cloneRes.RepoPath))
			check(ide.WriteMultiIDE(assistants, prompt, summary, targetPicker.Yes))

			if targetPicker.IDEName != "" {
				check(launch(ide.IDE(targetPicker.IDEName), cloneRes.RepoPath, start))
			}
			displayPath := ide.PathWithTilde(cloneRes.RepoPath)

			fmt.Printf("\nRepository cloned to %s\n", out.H(displayPath))
			fmt.Println("Now you can start your IDE and ask AI assistant to execute current feature. Happy coding!")
		},
	}
	targetPicker.Prepare(cmd)
	cmd.Flags().StringVarP(&repoName, "repo", "r", "", "Repository to clone (full name or url)")
	cmd.Flags().BoolVar(&start, "start", false, "Start execution immediately after cloning (only supported for ClaudeCode now)")

	return cmd
}

func launch(ideName ide.IDE, repoPath string, start bool) error {
	launched, err := ide.LaunchIDE(ideName, repoPath, start)
	if err != nil {
		return err
	}
	if launched {
		out.Successf(
			"%v launched at the repository. You can ask AI assitant there to execute current feature. Happy coding!",
			ideName)
	}
	return nil
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("Error: %v", err))
		os.Exit(1)
	}
}
