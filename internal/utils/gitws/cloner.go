package gitws

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
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
	var baseBranch string
	if !mainRepoExists {
		// Clone the main repository with a branch based on project name
		projectBranchName := sanitizeName(project.GetTitle(), 30)
		if err := cloneMainRepository(ctx, repo, mainRepoPath, projectBranchName); err != nil {
			return "", repo, err
		}

		// Write metadata for the main repository
		mainMeta := generateMetadata(repo, target, false)
		if err := metadata.EnsureMetadataSetup(mainRepoPath, mainMeta); err != nil {
			return "", repo, fmt.Errorf("failed to setup main repo metadata: %w", err)
		}
	} else {
		// Check for updates in the main branch
		_ = git.FetchRemote(mainRepoPath, "origin")
		if db, err := git.GetDefaultBranchName(mainRepoPath); err == nil {
			baseBranch = db
			if behind, _ := git.IsBehind(mainRepoPath, baseBranch); behind {
				var pullChanges bool
				err := huh.NewConfirm().
					Title(fmt.Sprintf("There are new changes in the %s branch. Do you want to pull them?", out.H(baseBranch))).
					Value(&pullChanges).
					Run()
				if err != nil {
					return "", repo, err
				}

				if pullChanges {
					if err := git.FastForwardBaseBranch(mainRepoPath, baseBranch); err != nil {
						out.Pwarnf("Failed to pull changes: %v\n", err)
					} else {
						out.Psuccessf("Updated %s branch\n", out.H(baseBranch))
					}
				}
			}
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
	if err := createWorktree(ctx, mainRepoPath, worktreePath, branchName, baseBranch); err != nil {
		return "", repo, err
	}

	// Write metadata for the worktree
	worktreeMeta := generateMetadata(repo, target, true)

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

func createWorktree(_ context.Context, mainRepoPath, worktreePath, branchName, base string) error {
	err := git.CreateWorktree(mainRepoPath, worktreePath, branchName, base)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "already registered worktree") || strings.Contains(errStr, "already exists") {
			var confirmed bool
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
				err = git.CreateWorktree(mainRepoPath, worktreePath, branchName, base)
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

// SanitizeName converts a name to a filesystem-safe format.
func SanitizeName(name string, maxLen int) string {
	return sanitizeName(name, maxLen)
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

// ResolveRepos resolves a list of repository full names to git.RepoInfo objects
// by matching against the company's available repositories.
func ResolveRepos(repoNames []string, companyID int32) ([]git.RepoInfo, error) {
	cl := devplan.NewClient(devplan.Config{})
	allRepos, err := cl.GetAllRepos(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get git repositories: %w", err)
	}
	byName := make(map[string]git.RepoInfo)
	for _, repo := range allRepos {
		byName[repo.GetFullName()] = git.RepoInfo{
			FullNames: []string{repo.GetFullName()},
			URLs:      []string{repo.GetUrl()},
		}
	}

	var resolved []git.RepoInfo
	var unresolved []string
	for _, name := range repoNames {
		// Try exact match first
		if info, ok := byName[name]; ok {
			resolved = append(resolved, info)
			continue
		}
		// Try case-insensitive substring match
		var matches []string
		var matchedInfo git.RepoInfo
		for fullName, info := range byName {
			if strings.Contains(strings.ToLower(fullName), strings.ToLower(name)) {
				matches = append(matches, fullName)
				matchedInfo = info
			}
		}
		found := len(matches) > 0
		if len(matches) > 1 {
			slog.Warn("Ambiguous repository name matched multiple repos, using first match", "name", name, "matches", matches)
		}
		if found {
			resolved = append(resolved, matchedInfo)
		}
		if !found {
			unresolved = append(unresolved, name)
		}
	}
	if len(unresolved) > 0 {
		return nil, fmt.Errorf("failed to resolve repositories: %s", strings.Join(unresolved, ", "))
	}
	return resolved, nil
}

type CloneAllReposResult struct {
	ParentPath string
	Repos      []git.RepoInfo
}

// CloneAllRepos clones multiple repositories into a parent directory using simple git clone (no worktrees).
// If branchName is non-empty, each cloned repo will have that branch created and checked out.
// If any clone fails, it returns immediately with an error naming the failed repo.
// Already-cloned repos are skipped, making the operation idempotent.
func CloneAllRepos(ctx context.Context, repos []git.RepoInfo, parentPath string, branchName string) (CloneAllReposResult, error) {
	if err := os.MkdirAll(parentPath, 0755); err != nil {
		return CloneAllReposResult{}, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	for _, repo := range repos {
		repoFullName := repo.GetFullName()
		parts := strings.Split(repoFullName, "/")
		shortName := parts[len(parts)-1]
		targetPath := filepath.Join(parentPath, shortName)

		if _, err := os.Stat(targetPath); err == nil {
			// Path exists — check if it's already a valid git repo
			if _, gitErr := git.RepoAtPath(targetPath); gitErr == nil {
				out.Psuccessf("Repository %s already cloned at %s, skipping\n", out.H(repoFullName), out.H(targetPath))
				continue
			}
			return CloneAllReposResult{}, fmt.Errorf("path %s already exists but is not a valid git repository", targetPath)
		}

		if err := cloneMainRepository(ctx, repo, targetPath, branchName); err != nil {
			return CloneAllReposResult{}, fmt.Errorf("failed to clone repository %s: %w", repoFullName, err)
		}
	}

	return CloneAllReposResult{ParentPath: parentPath, Repos: repos}, nil
}

func generateMetadata(repo git.RepoInfo, target picker.DevTarget, includeTaskInfo bool) metadata.Metadata {
	project := target.ProjectWithDocs.GetProject()
	meta := metadata.Metadata{
		ProjectID:        project.GetId(),
		ProjectName:      project.GetTitle(),
		RepoURL:          repo.URLs[0],
		RepoName:         repo.GetFullName(),
		ProjectNumericID: fmt.Sprintf("%v", project.GetNumericId()),
	}

	if !includeTaskInfo {
		return meta
	}

	fillStory := func(feat *documents.DocumentEntity) {
		meta.StoryID = feat.GetId()
		meta.StoryName = feat.GetTitle()
		meta.StoryNumericID = fmt.Sprintf("%v", feat.GetNumericId())
	}

	// Add task information if available
	if task := target.Task; task != nil {
		meta.TaskID = task.GetId()
		meta.TaskName = task.GetTitle()
		meta.TaskNumericID = fmt.Sprintf("%v", task.GetNumericId())
		for _, d := range target.ProjectWithDocs.GetDocs() {
			if d.GetId() == task.GetParentId() && d.GetType() == documents.DocumentType_FEATURE {
				fillStory(d)
				break
			}
		}
	} else if feature := target.SpecificFeature; feature != nil {
		fillStory(feature)
	}

	return meta
}
