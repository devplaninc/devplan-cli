package switch_cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/devplaninc/devplan-cli/internal/cmd/common"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/metadata"
	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"github.com/devplaninc/devplan-cli/internal/utils/recentactivity"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
	"github.com/opensdd/osdd-core/core/executable"
	"github.com/spf13/cobra"
)

var (
	Cmd = create()
)

func create() *cobra.Command {
	var ideName string
	cmd := &cobra.Command{
		Use:     "switch",
		Aliases: []string{"sw"},
		Short:   "List and switch between cloned features",
		Long:    `List all cloned features in the workspace and switch to one of them by opening it in your preferred IDE.`,
		Run: func(_ *cobra.Command, _ []string) {
			runSwitch(ideName)
		},
	}
	cmd.Flags().StringVarP(&ideName, "ide", "i", "", "IDE to use (e.g., vscode, intellij, cursor)")
	return cmd
}

func runSwitch(ideName string) {
	ctx := context.Background()
	features, err := workspace.ListClonedRepos()
	check(err)

	if len(features) == 0 {
		fmt.Println(out.Failf("No cloned features found. Use 'devplan clone' to clone a feature first."))
		os.Exit(1)
	}

	features = recentactivity.SortClonedFeatures(features)

	// Build options for selection with full paths
	options, hasAnyChanges := common.BuildFeatureOptions(features)

	// Show legend if there are any uncommitted changes
	if hasAnyChanges {
		common.ShowLegend()
	}

	var selectedIdx int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select a feature to switch to").
				Options(options...).
				Value(&selectedIdx),
		),
	)

	err = form.Run()
	if err != nil {
		fmt.Println(out.Failf("Feature selection failed: %v", err))
		os.Exit(1)
	}

	selectedFeature := features[selectedIdx]
	featurePath := selectedFeature.FullPath
	if meta, err := metadata.ReadMetadata(featurePath); err != nil {
		slog.Debug("Failed to read feature metadata", "path", featurePath, "err", err)
	} else if meta != nil && meta.TaskID != "" {
		if err := recentactivity.RecordTaskActivity(meta.TaskID, "switch"); err != nil {
			slog.Debug("Failed to record recent task activity", "taskID", meta.TaskID, "err", err)
		}
	}

	if ideName != "" {
		ideV := ide.IDE(strings.ToLower(ideName))
		launchSelectedIDE(ctx, ideV, featurePath)
		return
	}

	ides, err := ide.DetectInstalledIDEs()
	if err != nil {
		fmt.Println(out.Failf("Failed to detect installed IDEs: %v", err))
		os.Exit(1)
	}

	if len(ides) == 0 {
		fmt.Println(out.Failf("No supported IDEs found on your system."))
		os.Exit(1)
	}

	// Build IDE options
	ideNames := getIDENames(ides)
	var ideOptions []huh.Option[ide.IDE]
	for _, name := range ideNames {
		ideOptions = append(ideOptions, huh.NewOption(name.DisplayName(), name))
	}

	var selectedIDEName ide.IDE
	ideForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[ide.IDE]().
				Title("Select an IDE to open the feature").
				Options(ideOptions...).
				Value(&selectedIDEName),
		),
	)

	err = ideForm.Run()
	if err != nil {
		fmt.Println(out.Failf("IDE selection failed: %v", err))
		os.Exit(1)
	}

	launchSelectedIDE(ctx, selectedIDEName, featurePath)
}

func launchSelectedIDE(ctx context.Context, ideName ide.IDE, featurePath string) {
	// With worktrees, each feature directory is a worktree that can be directly opened
	fmt.Printf("Opening %s in %s...\n", out.H(featurePath), out.Hf("%v", ideName))
	prefs.SetLastIDE(string(ideName))
	outOnly := prefs.InstructionFile != ""
	res, err := executable.LaunchIDE(ctx, executable.LaunchParams{
		IDE:           string(ideName),
		RepoPath:      featurePath,
		OutputCMDOnly: outOnly,
	})
	check(err)
	check(ide.WriteLaunchResult(res))
	if !outOnly {
		fmt.Println(out.Successf("Successfully opened %s in %s", featurePath, ideName))
	}
}

func getIDENames(ides map[ide.IDE]string) []ide.IDE {
	var names []ide.IDE
	for name := range ides {
		names = append(names, name)
	}

	lastIDE := prefs.GetLastIDE()
	popularity := map[ide.IDE]int{
		ide.Claude:   1,
		ide.Cursor:   2,
		ide.WebStorm: 3,
		ide.PyCharm:  4,
	}

	sort.Slice(names, func(i, j int) bool {
		nameI := names[i]
		nameJ := names[j]

		// Last used one comes on top
		if string(nameI) == lastIDE {
			return true
		}
		if string(nameJ) == lastIDE {
			return false
		}

		// Then by popularity
		weightI, okI := popularity[nameI]
		weightJ, okJ := popularity[nameJ]

		if okI && okJ {
			if weightI != weightJ {
				return weightI < weightJ
			}
		} else if okI {
			return true
		} else if okJ {
			return false
		}

		// Then by alphabet
		return string(nameI) < string(nameJ)
	})

	return names
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("Error: %v", err))
		os.Exit(1)
	}
}
