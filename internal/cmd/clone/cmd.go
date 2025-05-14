package clone

import (
	"context"
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/utils/loaders"
	"github.com/devplaninc/devplan-cli/internal/utils/picker"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/integrations"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/devplaninc/devplan-cli/internal/components/spinner"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Cmd = create()
)

var defaultWorkspace = path.Join("devplan", "workspace")

const (
	workspaceConfigKey = "workspace_dir"
)

func create() *cobra.Command {
	featPicker := &picker.FeatureCmd{}
	var repoName string
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone a repository and focus on a feature",
		Long: `Clone a repository and focus on a feature.
This command streamlines the workflow of cloning a repository and focusing on a feature.
It will clone the repository into the configured workplace directory and set up the necessary rules.`,
		PreRunE: featPicker.PreRun,
		Run: func(_ *cobra.Command, _ []string) {
			runClone(featPicker, repoName)
		},
	}
	featPicker.Prepare(cmd)
	cmd.Flags().StringVarP(&repoName, "repo", "r", "", "Repository to clone (full name or url)")

	return cmd
}

func runClone(featPicker *picker.FeatureCmd, repoName string) {
	assistants, err := picker.AssistantForIDE(featPicker.IDEName)
	check(err)
	feat, err := picker.Feature(featPicker)
	check(err)
	feature := feat.Feature
	project := feat.ProjectWithDocs

	repo, err := confirmRepository(repoName, feature.GetCompanyId())
	check(err)
	repoURL := fmt.Sprintf("git@github.com:%s.git", repo.GetFullName())

	repoPath, gitRepo, err := prepareRepository(featPicker, repoURL, feature)
	check(err)
	summary, err := loaders.RepoSummary(feature, gitRepo)
	check(err)
	featPrompt, err := picker.GetFeaturePrompt(feature.GetId(), project.GetDocs())
	check(err)

	err = os.Chdir(repoPath)
	check(err)
	check(ide.WriteMultiIDE(assistants, featPrompt, summary, featPicker.Yes))

	if featPicker.IDEName != "" {
		launch(ide.IDE(featPicker.IDEName), repoPath)
		return
	}
	displayPath := pathWithTilde(repoPath)

	fmt.Printf("\nRepository cloned to %s\n", out.H(displayPath))
	fmt.Println("Now you can start your IDE and ask AI assistant to execute current feature. Happy coding!")
}

func prepareRepository(featPicker *picker.FeatureCmd, repoURL string, feature *documents.DocumentEntity) (string, git.RepoInfo, error) {
	repoPath, exists, err := getRepoPath(repoURL, feature)
	check(err)
	displayPath := pathWithTilde(repoPath)
	if !exists {
		check(cloneRepository(repoURL, repoPath))
		return repoPath, git.EnsureRepoPath(repoPath), nil
	}
	if len(featPicker.IDEName) == 0 {
		return "", git.RepoInfo{}, fmt.Errorf("repository already exists and no IDE to launch selected")
	}
	ideName := ide.IDE(featPicker.IDEName)
	if featPicker.Yes {
		out.Psuccessf("Repository %s already exists. Opening it in %v.\n", out.H(displayPath), out.H(ideName))
		return repoPath, git.EnsureRepoPath(repoPath), nil
	}
	p := promptui.Prompt{
		Label: fmt.Sprintf("Repository %s already exists. Do you want to open it in %v",
			displayPath, ideName),
		IsConfirm: true,
	}
	resp, err := p.Run()
	check(err)
	if resp != "y" {
		return "", git.RepoInfo{}, fmt.Errorf("repository already exists, selected not to open it")
	}
	return repoPath, git.EnsureRepoPath(repoPath), nil
}

func launch(ideName ide.IDE, repoPath string) {
	check(ide.LaunchIDE(ideName, repoPath))
	out.Successf(
		"%v launched at the repository. You can ask AI assitant there to execute current feature. Happy coding!",
		ideName)
}

func confirmRepository(repoName string, companyID int32) (*integrations.GitHubRepository, error) {
	cl := devplan.NewClient(devplan.Config{})
	resp, err := cl.GetIntegration(companyID, "github")
	if err != nil {
		return nil, fmt.Errorf("failed to get integration: %v", err)
	}
	var repoNames []string
	byName := make(map[string]*integrations.GitHubRepository)
	for _, repo := range resp.GetInfo().GetGithub().GetRepositories() {
		if len(repoName) > 0 && strings.Contains(strings.ToLower(repo.GetFullName()), strings.ToLower(repoName)) {
			return repo, nil
		}
		repoNames = append(repoNames, repo.GetFullName())
		byName[repo.GetFullName()] = repo
	}

	// Prompt user to select a repository
	prompt := promptui.Select{
		Label: "Select repository",
		Items: repoNames,
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("repository selection failed: %v", err)
	}
	selectedRepoName := repoNames[idx]
	return byName[selectedRepoName], nil
}

func getWorkspaceDir() string {
	workspaceDir := viper.GetString(workspaceConfigKey)
	if workspaceDir == "" {
		// Use default directory in user's home
		home, err := os.UserHomeDir()
		if err != nil {
			out.Pfailf("Failed to get user home directory: %v\n", err)
			os.Exit(1)
		}
		workspaceDir = filepath.Join(home, defaultWorkspace)

		// Save to config for future use
		viper.Set(workspaceConfigKey, workspaceDir)
		err = viper.WriteConfig()
		if err != nil {
			out.Pfailf("Failed to write config: %v\n", err)
			// Continue anyway, just log the error
		}
	}

	// Ensure directory exists
	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		err = os.MkdirAll(workspaceDir, 0755)
		if err != nil {
			out.Pfailf("Failed to create workplace directory: %v\n", err)
			os.Exit(1)
		}
	}

	return workspaceDir
}

func sanitizeDirName(name string) string {
	if len(name) > 30 {
		name = name[:30]
	}
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	return reg.ReplaceAllString(name, "")
}

func getRepoPath(url string, feature *documents.DocumentEntity) (string, bool, error) {
	workspaceDir := getWorkspaceDir()
	dirName := sanitizeDirName(feature.GetTitle())
	repoParent := filepath.Join(workspaceDir, "features", fmt.Sprintf("%s", dirName))
	repoFullName, err := git.GetFullName(url)
	parts := strings.Split(repoFullName, "/")
	repoName := parts[len(parts)-1]
	if err != nil {
		return "", false, fmt.Errorf("failed to get repository name from URL: %v", err)
	}
	repoPath := filepath.Join(repoParent, repoName)
	if _, err := os.Stat(repoPath); err == nil {
		return repoPath, true, nil
	}
	return repoPath, false, nil
}

func cloneRepository(url string, path string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sp := spinner.New(
		fmt.Sprintf("Cloning repository %s into %s", out.H(url), out.H(path)),
		fmt.Sprintf("Repository cloned successfully to %v", out.H(path)),
	)

	// Clone in a goroutine so we can show a spinner
	errChan := make(chan error, 1)
	go func() {
		errChan <- git.Clone(url, path, sp.GetProgressWriter())
		cancel()
	}()

	err := sp.Run(ctx)
	if err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	err = <-errChan
	if err != nil {
		_ = os.RemoveAll(path)
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	return nil
}

func pathWithTilde(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	pref := fmt.Sprintf("%s/", home)
	if strings.HasPrefix(path, pref) {
		return "~/" + strings.TrimPrefix(path, pref)
	}
	return path
}

func check(err error) {
	if err != nil {
		fmt.Println(out.Failf("Error: %v", err))
		os.Exit(1)
	}
}
