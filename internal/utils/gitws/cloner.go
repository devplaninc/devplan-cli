package gitws

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/devplaninc/devplan-cli/internal/components/spinner"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/metadata"
	"github.com/devplaninc/devplan-cli/internal/utils/picker"
	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
)

type InteractiveCloneResult struct {
	Target   picker.DevTarget
	RepoPath string
	RepoInfo git.RepoInfo
}

func InteractiveClone(ctx context.Context, targetPicker *picker.TargetCmd, repoName string, branchName string) (InteractiveCloneResult, error) {
	target, err := picker.Target(targetPicker)
	if err != nil {
		return InteractiveCloneResult{}, err
	}
	project := target.ProjectWithDocs

	repo, err := confirmRepository(repoName, project.GetProject().GetCompanyId())
	if err != nil {
		return InteractiveCloneResult{}, err
	}

	repoPath, repo, err := prepareRepository(ctx, targetPicker, repo, target, branchName)
	if err != nil {
		return InteractiveCloneResult{}, err
	}
	return InteractiveCloneResult{
		Target:   target,
		RepoPath: repoPath,
		RepoInfo: repo,
	}, nil
}

func prepareRepository(
	ctx context.Context, featPicker *picker.TargetCmd, repo git.RepoInfo, target picker.DevTarget, branchName string,
) (string, git.RepoInfo, error) {
	// Get project name and repo name
	project := target.ProjectWithDocs.GetProject()
	projectName := sanitizeName(project.GetTitle(), 30)

	// Extract repo name from full name (e.g., "owner/repo" -> "repo")
	repoFullName := repo.GetFullName()
	parts := strings.Split(repoFullName, "/")
	repoName := parts[len(parts)-1]

	mainRepoPath := workspace.GetMainRepoPath(projectName, repoName)

	// Ensure the main repository exists
	mainRepoExists := workspace.MainRepoExists(projectName, repoName)
	if !mainRepoExists {
		// Clone the main repository with a branch based on project name
		projectBranchName := sanitizeName(project.GetTitle(), 30)
		if err := cloneMainRepository(ctx, repo, mainRepoPath, projectBranchName); err != nil {
			return "", repo, err
		}

		// Write metadata for the main repository
		mainMeta := metadata.Metadata{
			ProjectID:   project.GetId(),
			ProjectName: project.GetTitle(),
			RepoURL:     repo.URLs[0],
			RepoName:    repo.GetFullName(),
		}
		if err := metadata.EnsureMetadataSetup(mainRepoPath, mainMeta); err != nil {
			return "", repo, fmt.Errorf("failed to setup main repo metadata: %w", err)
		}
	}

	// Determine the worktree path and branch name
	taskName := sanitizeName(target.GetName(), 30)
	worktreePath := workspace.GetWorktreePath(projectName, taskName)

	// Use provided branch name, or fall back to sanitized task/feature name
	if branchName == "" {
		branchName = taskName
	}

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		// Worktree exists
		displayPath := ide.PathWithTilde(worktreePath)
		if len(featPicker.IDEName) == 0 {
			return "", git.RepoInfo{}, fmt.Errorf("worktree already exists and no IDE to launch selected")
		}
		ideName := ide.IDE(featPicker.IDEName)
		if featPicker.Yes {
			out.Psuccessf("Worktree %s already exists. Opening it in %v.\n", out.H(displayPath), out.H(ideName))
			return worktreePath, git.EnsureRepoPath(worktreePath), nil
		}
		var confirmed bool
		err := huh.NewConfirm().
			Title(fmt.Sprintf("Worktree %s already exists. Do you want to open it in %v?",
				displayPath, ideName)).
			Value(&confirmed).
			Run()
		if err != nil {
			return "", repo, err
		}
		if !confirmed {
			return "", git.RepoInfo{}, fmt.Errorf("worktree already exists, selected not to open it")
		}
		return worktreePath, git.EnsureRepoPath(worktreePath), nil
	}

	// Create the worktree
	if err := createWorktree(ctx, mainRepoPath, worktreePath, branchName); err != nil {
		return "", repo, err
	}

	// Write metadata for the worktree
	worktreeMeta := metadata.Metadata{
		ProjectID:   project.GetId(),
		ProjectName: project.GetTitle(),
		RepoURL:     repo.URLs[0],
		RepoName:    repo.GetFullName(),
	}

	// Add task information if available
	if task := target.Task; task != nil {
		worktreeMeta.TaskID = task.GetId()
		worktreeMeta.TaskName = task.GetTitle()
	} else if feature := target.SpecificFeature; feature != nil {
		worktreeMeta.TaskID = feature.GetId()
		worktreeMeta.TaskName = feature.GetTitle()
	}

	if err := metadata.EnsureMetadataSetup(worktreePath, worktreeMeta); err != nil {
		return "", repo, fmt.Errorf("failed to setup worktree metadata: %w", err)
	}

	repoInfo, err := git.RepoAtPath(worktreePath)
	if err != nil {
		return "", repo, fmt.Errorf("failed to verify worktree as git repository: %w", err)
	}
	return worktreePath, repoInfo, nil
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
	repos, err := cl.GetAllRepos(companyID)
	if err != nil {
		return git.RepoInfo{}, fmt.Errorf("failed to get git repositories: %v", err)
	}
	var repoNames []string
	byName := make(map[string]git.RepoInfo)
	for _, repo := range repos {
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
	var selectedRepoName string
	err = huh.NewSelect[string]().
		Title("Select repository").
		Options(huh.NewOptions(repoNames...)...).
		Value(&selectedRepoName).
		Run()
	if err != nil {
		return git.RepoInfo{}, fmt.Errorf("repository selection failed: %v", err)
	}
	return byName[selectedRepoName], nil
}

func cloneMainRepository(ctx context.Context, repo git.RepoInfo, path string, branchToCreate string) error {
	for _, url := range repo.URLs {
		if strings.Contains(strings.ToLower(url), "github.com") {
			return cloneGithubRepo(ctx, repo, path, branchToCreate)
		}
		if strings.Contains(strings.ToLower(url), "bitbucket.org") {
			return cloneBitbucketRepo(ctx, repo, path, branchToCreate)
		}
	}
	return fmt.Errorf("unsupported repository URL: %+v", repo)
}

func createWorktree(_ context.Context, mainRepoPath, worktreePath, branchName string) error {
	err := git.CreateWorktree(mainRepoPath, worktreePath, branchName)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "already registered worktree") || strings.Contains(errStr, "already exists") {
			confirmed := false
			err := huh.NewConfirm().
				Title(fmt.Sprintf("Worktree at %s already exists or is registered. Do you want to remove/prune it and try again?", worktreePath)).
				Value(&confirmed).
				Run()
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}
			if confirmed {
				fmt.Println("Pruning and removing existing worktree registration...")
				_ = git.PruneWorktrees(mainRepoPath)
				_ = git.RemoveWorktree(mainRepoPath, worktreePath)
				_ = os.RemoveAll(worktreePath)

				// Retry
				err = git.CreateWorktree(mainRepoPath, worktreePath, branchName)
				if err == nil {
					out.Psuccessf("Worktree %s created successfully after cleanup\n", out.H(worktreePath))
					return nil
				}
			}
		}
		_ = os.RemoveAll(worktreePath)
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	out.Psuccessf("Worktree %s created successfully\n", out.H(worktreePath))
	return nil
}

func cloneBitbucketRepo(ctx context.Context, repo git.RepoInfo, path string, branchToCreate string) error {
	var err error
	for _, url := range repo.URLs {
		err = tryRepoClone(ctx, url, path, branchToCreate)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to clone repository .\n"+
		"Please ensure you have:\n"+
		"1. (Recommended) Git Credential Manager installed - https://github.com/git-ecosystem/git-credential-manager \n"+
		"2. Valid Bitbucket credentials configured for SSH is you use SSH-based connection\n"+
		"3. Proper network access to BitBucket\n"+
		"Original error: %w", err)
}

func cloneGithubRepo(ctx context.Context, repo git.RepoInfo, path string, branchToCreate string) error {
	protocol := prefs.GetLastGitProtocol()
	httpsURL := fmt.Sprintf("https://github.com/%s", repo.GetFullName())
	sshURL := fmt.Sprintf("git@github.com:%s.git", repo.GetFullName())
	var urls []urlDef
	if protocol == prefs.HTTPS {
		urls = []urlDef{{httpsURL, prefs.HTTPS}, {sshURL, prefs.SSH}}
	} else {
		urls = []urlDef{{sshURL, prefs.SSH}, {httpsURL, prefs.HTTPS}}
	}
	var err error
	for _, url := range urls {
		err = tryRepoClone(ctx, url.url, path, branchToCreate)
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

func tryRepoClone(ctx context.Context, url string, path string, branchToCreate string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	authMessage := ""
	if strings.Contains(url, "github.com") {
		authMessage = out.Hf("If you experience authentication issue, use 'gh auth login' command to initialize auth")
	}
	if authMessage != "" {
		fmt.Println(authMessage)
	}
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

func sanitizeName(name string, maxLen int) string {
	if len(name) > maxLen {
		name = name[:maxLen]
	}
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	return reg.ReplaceAllString(name, "")
}

type urlDef struct {
	url      string
	protocol prefs.GitProtocol
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
