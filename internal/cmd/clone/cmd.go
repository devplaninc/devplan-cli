package clone

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/devplaninc/devplan-cli/internal/utils/loaders"
	"github.com/devplaninc/devplan-cli/internal/utils/picker"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"

	"github.com/devplaninc/devplan-cli/internal/components/spinner"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"github.com/manifoldco/promptui"
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
			runClone(targetPicker, repoName, start)
		},
	}
	targetPicker.Prepare(cmd)
	cmd.Flags().StringVarP(&repoName, "repo", "r", "", "Repository to clone (full name or url)")
	cmd.Flags().BoolVar(&start, "start", false, "Start execution immediately after cloning (only supported for ClaudeCode now)")

	return cmd
}

func runClone(targetPicker *picker.TargetCmd, repoName string, start bool) {
	assistants, err := picker.AssistantForIDE(targetPicker.IDEName)
	check(err)
	target, err := picker.Target(targetPicker)
	check(err)
	project := target.ProjectWithDocs

	repo, err := confirmRepository(repoName, project.GetProject().GetCompanyId())
	check(err)

	repoPath, gitRepo, err := prepareRepository(targetPicker, repo, target)
	check(err)
	summary, err := loaders.RepoSummary(target, gitRepo)
	check(err)

	prompt, err := picker.GetTargetPrompt(target, project.GetDocs())
	check(err)

	if err := ide.CleanupCurrentFeaturePrompts(assistants); err != nil {
		out.Pfailf("Failed to clean up prompt files: %v\n", err)
	}

	err = os.Chdir(repoPath)
	check(err)
	check(ide.WriteMultiIDE(assistants, prompt, summary, targetPicker.Yes))

	if targetPicker.IDEName != "" {
		launch(ide.IDE(targetPicker.IDEName), repoPath, start)
		return
	}
	displayPath := pathWithTilde(repoPath)

	fmt.Printf("\nRepository cloned to %s\n", out.H(displayPath))
	fmt.Println("Now you can start your IDE and ask AI assistant to execute current feature. Happy coding!")
}

func prepareRepository(
	featPicker *picker.TargetCmd, repo git.RepoInfo, target picker.DevTarget,
) (string, git.RepoInfo, error) {
	repoPath, exists, err := getRepoPath(repo, target)
	check(err)
	displayPath := pathWithTilde(repoPath)
	if !exists {
		branchName := sanitizeName(target.GetName(), 30)
		check(cloneRepository(repo, repoPath, branchName))
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

func launch(ideName ide.IDE, repoPath string, start bool) {
	launched, err := ide.LaunchIDE(ideName, repoPath, start)
	check(err)
	if launched {
		out.Successf(
			"%v launched at the repository. You can ask AI assitant there to execute current feature. Happy coding!",
			ideName)
	}
}

func checkIfURL(repoName string) (git.RepoInfo, bool, error) {
	if git.IsValidURL(repoName) {
		repo, err := git.GetRepoInfoFromURL(repoName)
		if err == nil {
			prefs.AddExtraGitURL(repoName)
		}
		return repo, err == nil, err
	}
	return git.RepoInfo{}, false, nil
}

func confirmRepository(repoName string, companyID int32) (git.RepoInfo, error) {
	repo, ok, err := checkIfURL(repoName)
	if ok {
		return repo, nil
	}
	if err != nil {
		return git.RepoInfo{}, err
	}
	cl := devplan.NewClient(devplan.Config{})
	resp, err := cl.GetIntegration(companyID, "github")
	if err != nil {
		return git.RepoInfo{}, fmt.Errorf("failed to get integration: %v", err)
	}
	var repoNames []string
	byName := make(map[string]git.RepoInfo)
	for _, repo := range resp.GetInfo().GetGithub().GetRepositories() {
		repoInfo := git.RepoInfo{
			FullNames: []string{repo.GetFullName()},
			URLs:      []string{repo.GetUrl()},
		}
		if len(repoName) > 0 && strings.Contains(strings.ToLower(repo.GetFullName()), strings.ToLower(repoName)) {
			return repoInfo, nil
		}
		repoNames = append(repoNames, repo.GetFullName())
		byName[repo.GetFullName()] = repoInfo
	}
	for _, url := range prefs.GetExtraGitURLs() {
		repo, err := git.GetRepoInfoFromURL(url)
		if err != nil {
			// Ignoring the repo we cannot parse
			continue
		}
		if _, ok := byName[repo.GetFullName()]; !ok {
			repoNames = append(repoNames, repo.GetFullName())
			byName[repo.GetFullName()] = repo
		}
	}

	// Prompt user to select a repository
	prompt := promptui.Select{
		Label: "Select repository",
		Items: repoNames,
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return git.RepoInfo{}, fmt.Errorf("repository selection failed: %v", err)
	}
	selectedRepoName := repoNames[idx]
	return byName[selectedRepoName], nil
}

func sanitizeName(name string, maxLen int) string {
	if len(name) > maxLen {
		name = name[:maxLen]
	}
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	return reg.ReplaceAllString(name, "")
}

func getRepoPath(repo git.RepoInfo, target picker.DevTarget) (string, bool, error) {
	dirName := sanitizeName(target.GetName(), 30)
	repoParent := filepath.Join(workspace.GetFeaturesPath(), fmt.Sprintf("%s", dirName))
	repoFullName := repo.GetFullName()
	parts := strings.Split(repoFullName, "/")
	repoName := parts[len(parts)-1]
	repoPath := filepath.Join(repoParent, repoName)
	if _, err := os.Stat(repoPath); err == nil {
		return repoPath, true, nil
	}
	return repoPath, false, nil
}

type urlDef struct {
	url      string
	protocol prefs.GitProtocol
}

func cloneRepository(repo git.RepoInfo, path string, branchToCreate string) error {
	// Use the last successful protocol from preferences
	protocol := prefs.GetLastGitProtocol()
	httpsURL := fmt.Sprintf("https://github.com/%s", repo.GetFullName())
	sshURL := fmt.Sprintf("git@github.com:%s.git", repo.GetFullName())
	var urls []urlDef
	if protocol == prefs.SSH {
		urls = []urlDef{{sshURL, prefs.SSH}, {httpsURL, prefs.HTTPS}}
	} else {
		urls = []urlDef{{httpsURL, prefs.HTTPS}, {sshURL, prefs.SSH}}
	}
	var err error
	for _, url := range urls {
		err = tryRepoClone(url.url, path, branchToCreate)
		if err == nil {
			if protocol != url.protocol {
				prefs.SetLastGitProtocol(url.protocol)
			}
			return nil
		}
	}

	return fmt.Errorf("failed to clone repository using both SSH and HTTPS.\n"+
		"Please ensure you have:\n"+
		"1. Valid GitHub credentials configured\n"+
		"2. Either SSH keys set up or GitHub Personal Access Token configured\n"+
		"3. Proper network access to GitHub\n"+
		"Original error: %w", err)
}

func tryRepoClone(url string, path string, branchToCreate string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sp := spinner.New(
		fmt.Sprintf("Cloning repository %s into %s", out.H(url), out.H(path)),
		fmt.Sprintf("Repository %v cloned successfully to %v", out.H(url), out.H(path)),
	)

	// Clone in a goroutine so we can show a spinner
	errChan := make(chan error, 1)
	go func() {
		errChan <- git.Clone(git.CloneOptions{
			RepoURL:          url,
			TargetPath:       path,
			OutWriter:        sp.GetProgressWriter(),
			CreateBranchName: branchToCreate,
		})
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
