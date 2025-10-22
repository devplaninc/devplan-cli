package gitws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/devplaninc/devplan-cli/internal/components/spinner"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/picker"
	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
	"github.com/manifoldco/promptui"
)

type InteractiveCloneResult struct {
	Target   picker.DevTarget
	RepoPath string
	RepoInfo git.RepoInfo
}

func InteractiveClone(ctx context.Context, targetPicker *picker.TargetCmd, repoName string) (InteractiveCloneResult, error) {
	target, err := picker.Target(targetPicker)
	if err != nil {
		return InteractiveCloneResult{}, err
	}
	project := target.ProjectWithDocs

	repo, err := confirmRepository(repoName, project.GetProject().GetCompanyId())
	if err != nil {
		return InteractiveCloneResult{}, err
	}

	repoPath, repo, err := prepareRepository(ctx, targetPicker, repo, target)
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
	ctx context.Context, featPicker *picker.TargetCmd, repo git.RepoInfo, target picker.DevTarget,
) (string, git.RepoInfo, error) {
	repoPath, exists, err := getRepoPath(repo, target)
	if err != nil {
		return "", repo, err
	}
	displayPath := ide.PathWithTilde(repoPath)
	if !exists {
		branchName := sanitizeName(target.GetName(), 30)
		if err := cloneRepository(ctx, repo, repoPath, branchName); err != nil {
			return "", repo, err
		}
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
	if err != nil {
		return "", repo, err
	}
	if resp != "y" {
		return "", git.RepoInfo{}, fmt.Errorf("repository already exists, selected not to open it")
	}
	return repoPath, git.EnsureRepoPath(repoPath), nil
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

func cloneRepository(ctx context.Context, repo git.RepoInfo, path string, branchToCreate string) error {
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
	if protocol == prefs.SSH {
		urls = []urlDef{{sshURL, prefs.SSH}, {httpsURL, prefs.HTTPS}}
	} else {
		urls = []urlDef{{httpsURL, prefs.HTTPS}, {sshURL, prefs.SSH}}
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
