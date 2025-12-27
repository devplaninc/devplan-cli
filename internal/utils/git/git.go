package git

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

var (
	ErrRepositoryNotExists = errors.New("repository does not exist")
)

type RepoInfo struct {
	URLs      []string
	FullNames []string
}

func (r RepoInfo) MatchesName(fullName string) bool {
	for _, fn := range r.FullNames {
		if fn == fullName {
			return true
		}
	}
	return false
}

func (r RepoInfo) GetFullName() string {
	return r.FullNames[0]
}

func IsNotInRepoErr(err error) bool {
	return errors.Is(err, ErrRepositoryNotExists)
}

func EnsureRepoPath(path string) RepoInfo {
	repo, err := RepoAtPath(path)
	if err == nil {
		return repo
	}
	if !IsNotInRepoErr(err) {
		panic(fmt.Sprintf("Failed to get current repository: %v", err))
	}
	panic("Not in a git repository. Please clone a repository first or navigate to the cloned one and run the command from inside the git repository.")
}

func EnsureInRepo() RepoInfo {
	return EnsureRepoPath(".")
}

func CurrentRepo() (RepoInfo, error) {
	return RepoAtPath(".")
}

func RepoAtPath(path string) (RepoInfo, error) {
	urls, err := getRepoURLs(path)
	if err != nil {
		return RepoInfo{}, err
	}
	seenNames := make(map[string]bool)
	var fullNames []string
	for _, u := range urls {
		fn, err := GetFullName(u)
		if err != nil {
			return RepoInfo{}, err
		}
		if seenNames[fn] {
			continue
		}
		seenNames[fn] = true
		fullNames = append(fullNames, fn)
	}
	return RepoInfo{URLs: urls, FullNames: fullNames}, nil
}

func GetRepoInfoFromURL(url string) (RepoInfo, error) {
	fn, err := GetFullName(url)
	if err != nil {
		return RepoInfo{}, err
	}
	return RepoInfo{URLs: []string{url}, FullNames: []string{fn}}, nil
}

func IsValidURL(url string) bool {
	return strings.Contains(url, "://") || strings.Contains(url, "@")
}

func GetFullName(url string) (string, error) {
	if !IsValidURL(url) {
		return "", fmt.Errorf("invalid URL format: %s", url)
	}

	endpoint, err := transport.NewEndpoint(url)
	if err != nil {
		return "", err
	}
	name := strings.TrimSuffix(endpoint.Path, ".git")
	name = strings.TrimPrefix(name, "/")
	return name, nil
}

type CloneOptions struct {
	RepoURL          string
	TargetPath       string
	CreateBranchName string
	OutWriter        io.Writer
}

func Clone(opt CloneOptions) error {
	cmd := exec.Command("git", "clone", opt.RepoURL, opt.TargetPath)
	if o := opt.OutWriter; o != nil {
		cmd.Stdout = o
		cmd.Stderr = o
	}
	err := cmd.Run()
	if err != nil {
		return err
	}

	if opt.CreateBranchName != "" {
		if err := SetupBranch(opt.TargetPath, opt.CreateBranchName); err != nil {
			return err
		}
	}

	return nil
}

// SetupBranch checks if a remote branch exists and checks it out, otherwise creates a new branch.
// This is the isolated branch setup logic used after cloning a repository.
func SetupBranch(repoPath, branchName string) error {
	// Check if remote branch exists
	remoteExists, err := RemoteBranchExists(repoPath, branchName)
	if err != nil {
		return fmt.Errorf("failed to check remote branch: %w", err)
	}

	if remoteExists {
		// Remote branch exists - checkout from remote
		if err := CheckoutRemoteBranch(repoPath, branchName, "origin"); err != nil {
			return fmt.Errorf("failed to checkout remote branch: %w", err)
		}
		return nil
	}

	// Remote branch doesn't exist - create new branch
	if err := CreateAndCheckoutBranch(repoPath, branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}

func GetRoot() (string, error) {
	// Use git command to get the root directory (works with worktrees)
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// LocalBranchExists checks if a local branch with the given name exists
func LocalBranchExists(repoPath, branchName string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				return false, nil // Branch doesn't exist
			}
		}
		return false, err // Actual error
	}
	return true, nil // Branch exists
}

// RemoteBranchExists checks if a remote branch with the given name exists on origin.
// Uses ls-remote to query the remote directly without fetching the entire repo.
func RemoteBranchExists(repoPath, branchName string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "ls-remote", "--heads", "origin", branchName)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// ls-remote returns exit code 2 for connection errors, etc.
			return false, fmt.Errorf("failed to query remote: %w", err)
		}
		return false, err
	}
	// ls-remote returns empty output if branch doesn't exist
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// FetchRemote fetches updates from a remote to ensure remote refs are up-to-date
func FetchRemote(repoPath, remoteName string) error {
	cmd := exec.Command("git", "-C", repoPath, "fetch", remoteName)
	return cmd.Run()
}

// CheckoutLocalBranch checks out an existing local branch
func CheckoutLocalBranch(repoPath, branchName string) error {
	cmd := exec.Command("git", "-C", repoPath, "checkout", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// CheckoutRemoteBranch creates and checks out a local branch that tracks a remote branch
func CheckoutRemoteBranch(repoPath, branchName, remoteName string) error {
	cmd := exec.Command("git", "-C", repoPath, "checkout", "-b", branchName, remoteName+"/"+branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// CreateAndCheckoutBranch creates a new branch and checks it out
func CreateAndCheckoutBranch(repoPath, branchName string) error {
	cmd := exec.Command("git", "-C", repoPath, "checkout", "-b", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// isValidBranchName checks if the branch name follows git naming rules
func isValidBranchName(name string) bool {
	if name == "" {
		return false
	}
	// Git branch name restrictions
	if strings.HasPrefix(name, "-") || strings.HasPrefix(name, ".") {
		return false
	}
	if strings.HasSuffix(name, ".") || strings.HasSuffix(name, ".lock") {
		return false
	}
	// Invalid characters and sequences
	invalidChars := []string{" ", "~", "^", ":", "?", "*", "[", "\\", ".."}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return false
		}
	}
	return true
}

// CheckoutBranch orchestrates the branch checkout logic based on branch existence.
// It handles the following scenarios:
// - Remote branch exists: creates local tracking branch
// - Only local branch exists: checks out local branch
// - Neither exists: creates new local branch
// Note: If already on the requested branch, git checkout succeeds (no-op).
// If checkout fails due to uncommitted changes, the git error is returned.
func CheckoutBranch(repoPath, branchName string) error {
	// Validate branch name
	if !isValidBranchName(branchName) {
		return fmt.Errorf("invalid branch name: %s", branchName)
	}

	// Check if remote branch exists using ls-remote (no fetch required)
	remoteExists, err := RemoteBranchExists(repoPath, branchName)
	if err != nil {
		fmt.Println(out.Warnf("Could not check remote branch, using local refs: %v", err))
		remoteExists = false
		// Continue anyway - remote might not be available but local branches still work
	}

	if remoteExists {
		// Check if local branch already exists (tracking the remote)
		localExists, err := LocalBranchExists(repoPath, branchName)
		if err != nil {
			return fmt.Errorf("failed to check local branch: %w", err)
		}

		if localExists {
			// Local branch exists, just checkout
			if err := CheckoutLocalBranch(repoPath, branchName); err != nil {
				return err
			}
			out.Psuccessf("Checked out branch %s\n", out.H(branchName))
		} else {
			// Fetch to get the remote branch ref before checkout
			if err := FetchRemote(repoPath, "origin"); err != nil {
				return fmt.Errorf("failed to fetch from remote: %w", err)
			}
			// Create local tracking branch from remote
			if err := CheckoutRemoteBranch(repoPath, branchName, "origin"); err != nil {
				return err
			}
			out.Psuccessf("Checked out branch %s from remote origin\n", out.H(branchName))
		}
		return nil
	}

	// Remote doesn't exist, check local
	localExists, err := LocalBranchExists(repoPath, branchName)
	if err != nil {
		return fmt.Errorf("failed to check local branch: %w", err)
	}

	if localExists {
		// Checkout local branch
		if err := CheckoutLocalBranch(repoPath, branchName); err != nil {
			return err
		}
		out.Psuccessf("Checked out local branch %s (note: no remote branch exists)\n", out.H(branchName))
		return nil
	}

	// Neither local nor remote exists, create new branch
	if err := CreateAndCheckoutBranch(repoPath, branchName); err != nil {
		return err
	}
	out.Psuccessf("Created and checked out new branch %s\n", out.H(branchName))
	return nil
}

// WorktreeInfo contains information about a git worktree
type WorktreeInfo struct {
	Path   string
	Branch string
	Commit string
}

// CreateWorktree creates a new git worktree at the specified path with the given branch name
func CreateWorktree(repoPath, worktreePath, branchName string) error {
	// Check if worktree path already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return fmt.Errorf("worktree path already exists: %s", worktreePath)
	}

	// Check if branch exists locally or remotely
	localExists, err := LocalBranchExists(repoPath, branchName)
	if err != nil {
		return fmt.Errorf("failed to check local branch: %w", err)
	}

	remoteExists, err := RemoteBranchExists(repoPath, branchName)
	if err != nil {
		// If we can't check remote, continue with local-only mode
		remoteExists = false
	}

	var cmd *exec.Cmd
	if localExists || remoteExists {
		// Checkout existing branch
		cmd = exec.Command("git", "-C", repoPath, "worktree", "add", worktreePath, branchName)
	} else {
		// Create new branch
		cmd = exec.Command("git", "-C", repoPath, "worktree", "add", "-b", branchName, worktreePath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

// RemoveWorktree removes a worktree at the specified path
func RemoveWorktree(repoPath, worktreePath string) error {
	cmd := exec.Command("git", "-C", repoPath, "worktree", "remove", worktreePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// PruneWorktrees cleans up worktree administrative files
func PruneWorktrees(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "worktree", "prune")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to prune worktrees: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// IsWorktree checks if the given path is a git worktree (not the main repository)
func IsWorktree(path string) (bool, error) {
	gitDirPath := filepath.Join(path, ".git")

	// Check if .git exists
	info, err := os.Stat(gitDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	// If .git is a file (not a directory), it's a worktree
	if !info.IsDir() {
		return true, nil
	}

	// If .git is a directory, it could be the main repo
	// Check if it's a worktree by looking for the worktree admin files
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-common-dir")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	commonDir := strings.TrimSpace(string(output))
	gitDir := filepath.Join(path, ".git")

	// If common-dir is different from .git, it's a worktree
	return commonDir != gitDir && commonDir != ".", nil
}

// GetMainRepoPath returns the path to the main repository from a worktree
func GetMainRepoPath(worktreePath string) (string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "rev-parse", "--git-common-dir")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get common git dir: %w", err)
	}

	commonDir := strings.TrimSpace(string(output))

	// The common dir is the .git directory of the main repo
	// Get the parent directory
	mainRepoPath := filepath.Dir(commonDir)

	return mainRepoPath, nil
}

func getRepoURLs(path string) ([]string, error) {
	// Use git command directly to get remote URL (works with worktrees)
	cmd := exec.Command("git", "-C", path, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Check if it's a "not a git repository" error
			if strings.Contains(string(exitErr.Stderr), "not a git repository") {
				return nil, ErrRepositoryNotExists
			}
		}
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}

	url := strings.TrimSpace(string(output))
	if url == "" {
		return nil, fmt.Errorf("no remote URL found")
	}

	return []string{url}, nil
}
